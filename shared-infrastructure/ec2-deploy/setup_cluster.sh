#!/bin/bash

# Usage: ./setup_cluster.sh /path/to/ssh_key
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


# Get terraform output as JSON
terraform output -json > tf_output.json

MASTER_SSH=$(jq -r '.ssh_commands_to_master_node.value' tf_output.json)
WORKER_SSHS=$(jq -r '.ssh_commands_to_worker_nodes.value[]' tf_output.json)


JOIN_CMD=$($MASTER_SSH "sudo kubeadm token create --print-join-command")
if [ -z "$JOIN_CMD" ]; then
    echo "Failed to get join command from master"
    exit 1
fi
echo "Join command: $JOIN_CMD"

for WORKER in $WORKER_SSHS; do
    echo "Joining worker: $WORKER"
    (ssh -A "$WORKER" "sudo $JOIN_CMD")
done