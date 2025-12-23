#!/bin/bash

set -e  # Exit on error

# Usage: ./connect_kubectl.sh /path/to/ssh_key
if [ -z "$1" ]; then
    echo "Usage: $0 /path/to/ssh_key"
    exit 1
fi
SSH_KEY_PATH="$1"
if [ ! -f "$SSH_KEY_PATH" ]; then
    echo "Error: SSH key file not found at $SSH_KEY_PATH"
    exit 1
fi

# Start ssh-agent and add the key
if [ -z "$SSH_AUTH_SOCK" ]; then
    eval "$(ssh-agent -s)"
fi
ssh-add "$SSH_KEY_PATH"

# Check for jq
if ! command -v jq &>/dev/null; then
    echo "Error: jq is required but not installed."
    exit 1
fi

# Ensure tf_output.json exists or generate it
if [ ! -f tf_output.json ]; then
    terraform output -json > tf_output.json
fi

# Extract master SSH command and hostname
MASTER_SSH=$(jq -r '.ssh_commands_to_master_node.value' tf_output.json)
username=$(echo "$MASTER_SSH" | awk '{print $2}' | cut -d'@' -f1)
hostname=$(echo "$MASTER_SSH" | awk '{print $2}' | cut -d'@' -f2)
ip_address=$(echo "$hostname" | sed -E 's/ec2-([0-9]+)-([0-9]+)-([0-9]+)-([0-9]+).*/\1.\2.\3.\4/')

kubeconfig_path="${HOME}/.kube/config"

# Make sure the .kube directory exists
mkdir -p "${HOME}/.kube"

echo "Master node's IP address: $ip_address"
echo "Master node's EC2 hostname: $hostname"
echo "Copying the kubeconfig file to the local machine..."

# Check if the remote host is reachable
if ! ssh -q -o ConnectTimeout=5 "$username@$hostname" exit &>/dev/null; then
    echo "Error: Cannot connect to $hostname. Please check your SSH access."
    exit 1
fi

# Copy the kubeconfig file
scp "$username@$hostname:~/.kube/config" "$kubeconfig_path"
if [ $? -ne 0 ]; then
    echo "Error: Failed to copy kubeconfig file."
    exit 1
fi

echo "Modifying the kubeconfig file to use the master node's IP address..."
cp "$kubeconfig_path" "${kubeconfig_path}.backup"

sed -i.bak '4d' "$kubeconfig_path"
sed -i.bak "s|server: https://.*:6443|server: https://$ip_address:6443\n    insecure-skip-tls-verify: true|g" "$kubeconfig_path"

rm -f "${kubeconfig_path}.bak"

echo "Success! Your kubectl is now configured to connect to the Kubernetes cluster."
echo "Try running 'kubectl get nodes' to verify the connection."