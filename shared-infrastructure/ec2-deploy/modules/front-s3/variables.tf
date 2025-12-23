variable "bucket_name" {
  description = "S3 bucket name"
  type        = string
}

variable "api_endpoint" {
  description = "The URL of the API Gateway to access the backend"
  type        = string
}

variable "region" {
  type        = string
}
