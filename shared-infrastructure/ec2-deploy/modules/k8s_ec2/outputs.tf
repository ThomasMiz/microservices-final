output "master-node-fqdn" {
    value       = aws_instance.master_node.public_dns
    description = "Public DNS of the Kubernetes master node"
}
output "worker-nodes-fqdn" {
    value       = [for instance in aws_instance.worker_nodes : instance.public_dns]
    description = "Public DNS of the Kubernetes worker nodes"
}