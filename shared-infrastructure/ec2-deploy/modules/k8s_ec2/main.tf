terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.99.1"
    }
  }
}
resource "aws_instance" "master_node" {
    ami                    = var.ec2_ami
    instance_type          = var.instance_type
    key_name               = var.ec2_key_name
    vpc_security_group_ids = var.security_group_ids
    subnet_id              = var.subnet_id
    iam_instance_profile   = var.iam_instance_profile

    user_data_base64 = base64encode(templatefile("./scripts/initMasterCRIO.sh", {
        KUBERNETES_VERSION = "1.29.0-1.1",
        CRIO_OS            = "xUbuntu_22.04",
        CRIO_VERSION       = "1.27",
        NODE_NAME          = "master",
        POD_NETWORK_CIDR   = var.pod_network_cidr
    }))
  #So that the master node can be accessed from the internet
  associate_public_ip_address = true

    root_block_device {
        volume_size = 30
        volume_type = "gp3"
    }

    tags = {
        Name = "master_node"
    }
}
resource "aws_instance" "worker_nodes" {
    count                  = var.worker_node_count
    ami                    = var.ec2_ami
    instance_type          = var.instance_type
    key_name               = var.ec2_key_name
    vpc_security_group_ids = var.security_group_ids
    subnet_id              = var.subnet_id
    iam_instance_profile   = var.iam_instance_profile

    user_data_base64 = base64encode(templatefile("./scripts/initWorkerCRIO.sh", {
        HOSTNAME           = "worker-node-${count.index + 1}",
        CRIO_OS            = "xUbuntu_22.04",
        CRIO_VERSION       = "1.28",
        KUBERNETES_VERSION = "1.29.0-1.1",
    }))

    associate_public_ip_address = true
    root_block_device {
        volume_size = 30
        volume_type = "gp3"
    }

    tags = {
        Name = "worker-node-${count.index + 1}"
    }
}