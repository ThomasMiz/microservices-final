output "bucket-static-website-url" {
    value       = format("http://%s", aws_s3_bucket_website_configuration.frontend_website.website_endpoint)
    description = "The HTTP address of the static website hosting for this bucket"
}