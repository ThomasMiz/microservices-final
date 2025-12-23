locals {
  key_file_name           = "/mnt/final/shared-infrastructure/ec2-deploy/tmp/id_ed25519.pub"
  profile                 = "default"
  region                  = "us-east-1"
  worker_count            = 2
  s3_frontend_bucket_name = "chotel-frontend" # Must be a globally unique value!
  api_url                 = "http://chotel.cuini.me/api"
}
