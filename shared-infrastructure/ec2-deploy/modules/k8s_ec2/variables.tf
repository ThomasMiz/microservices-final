variable ec2_ami {
  type        = string
  default     = "ami-080e1f13689e07408" # Ubuntu Server 22.04 LTS (HVM), SSD Volume Type - ami-080e1f13689e07408 (us-east-1)
  description = "EC2 AMI ID for the Kubernetes nodes"
}

variable instance_type {
  type        = string
  default     = "t2.medium"
  description = "EC2 instance type for the Kubernetes nodes"
}

variable ec2_key_name {
  type        = string
  description = "Name of the EC2 Key Pair to access the instances"
}

variable worker_node_count {
  type        = number
  default     = 1
  description = "Number of worker nodes in the Kubernetes cluster"
}

variable security_group_ids {
  type        = list(string)
  description = "List of security group IDs to associate with the EC2 instances"
}

variable subnet_id {
  type        = string
  description = "Subnet ID to launch the EC2 instances in"
}

variable pod_network_cidr {
  type        = string
  default     = "192.168.0.0/16"
  description = "CIDR for the pod network"
}

variable iam_instance_profile {
  type        = string
  default     = null
  description = "Optional IAM instance profile name to associate with EC2 instances"
}

