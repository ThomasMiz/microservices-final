resource "aws_key_pair" "node_key" {
  key_name   = "k8s_cluster_keypair"
  public_key = file(local.key_file_name)
}

data "aws_availability_zones" "available_azs" {
  state = "available"
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.13.0"

  name = "k8s_vpc"
  cidr = "10.0.0.0/16"

  azs = slice(data.aws_availability_zones.available_azs.names, 0, 2)

  public_subnets      = ["10.0.110.0/24", "10.0.120.0/24"]
  public_subnet_names = ["public_subnet_1", "public_subnet_2"]

  private_subnets      = ["10.0.10.0/24", "10.0.20.0/24"] # private subnet with nat gw
  private_subnet_names = ["lambda_subnet_1", "lambda_subnet_2"]


  enable_nat_gateway = false
  enable_vpn_gateway = false

  enable_dns_hostnames = true
  create_igw           = true

  tags = {
    Terraform   = "true"
    Environment = "dev"
  }

}

resource "aws_security_group" "k8s_sg" {
  name        = "k8s_security_group"
  vpc_id      = module.vpc.vpc_id
  description = "Allow all inbound and outbound traffic"
}
resource "aws_vpc_security_group_ingress_rule" "k8s_sg_ingress_rule" {
  security_group_id = aws_security_group.k8s_sg.id
  ip_protocol = -1
  cidr_ipv4   = "0.0.0.0/0"
}

resource "aws_vpc_security_group_egress_rule" "k8s_sg_egress_rule" {
  security_group_id = aws_security_group.k8s_sg.id

  ip_protocol = -1
  cidr_ipv4   = "0.0.0.0/0"
}

# IAM role for nodes (existing role expected)
data "aws_iam_role" "labrole" {
  name = "LabRole"
}

# Create an instance profile that attaches the existing role to EC2 instances
resource "aws_iam_instance_profile" "lab_profile" {
  name = "${data.aws_iam_role.labrole.name}-instance-profile"
  role = data.aws_iam_role.labrole.name
}

module "k8s_ec2" {
    source = "./modules/k8s_ec2"
    depends_on = [module.vpc]

    ec2_key_name = aws_key_pair.node_key.key_name
    security_group_ids = [aws_security_group.k8s_sg.id]
    subnet_id    = module.vpc.public_subnets[0]
    worker_node_count = local.worker_count
    iam_instance_profile = aws_iam_instance_profile.lab_profile.name
}

output "ssh_commands_to_master_node" {
  description = "SSH command to access the Kubernetes master node"
  value       = "ssh ubuntu@${module.k8s_ec2.master-node-fqdn}"
}
output "ssh_commands_to_worker_nodes" {
  description = "SSH commands to access the Kubernetes worker nodes"
  value       = [for fqdn in module.k8s_ec2.worker-nodes-fqdn : "ssh ubuntu@${fqdn}"]
}

module "front-s3" {
  source               = "./modules/front-s3"
  bucket_name          = local.s3_frontend_bucket_name
  api_endpoint         = local.api_url
  region               = local.region
}

output "frontend_s3_static_website_url" {
  description = "The HTTP address of the static website hosting for this bucket"
  value       = module.front-s3.bucket-static-website-url
}
