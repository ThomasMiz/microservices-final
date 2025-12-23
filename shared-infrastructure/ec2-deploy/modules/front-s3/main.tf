resource "aws_s3_bucket" "frontend_bucket" {
  bucket        = var.bucket_name
  force_destroy = true
}

resource "aws_s3_bucket_website_configuration" "frontend_website" {
  bucket = aws_s3_bucket.frontend_bucket.id

  index_document {
    suffix = "index.html"
  }

  error_document {
    key = "index.html"
  }
}

resource "aws_s3_bucket_public_access_block" "public_access_block" {
  bucket                  = aws_s3_bucket.frontend_bucket.id
  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_policy" "frontend_bucket_policy" {
  bucket = aws_s3_bucket.frontend_bucket.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = "*"
      Action    = "s3:GetObject"
      Resource  = "${aws_s3_bucket.frontend_bucket.arn}/*"
    }]
  })
}

resource "null_resource" "build_frontend" {
  provisioner "local-exec" {
    working_dir = "${path.root}/../../chotel-front"
        command     = <<EOT
      echo "const DEFAULT_API_BASE = '$API_ENDPOINT';" > ./js/config.js;

      DEST_DIR="tmp_build"

      find . -type f \
        ! -name 'tmp*' \
        \( -iname '*.css' -o -iname '*.js' -o -iname '*.ico' -o -iname '*.html' \) \
        -print0 |
      while IFS= read -r -d '' file; do
        rel_path="$${file#./}"
        mkdir -p "$DEST_DIR/$(dirname "$rel_path")"
        cp "$file" "$DEST_DIR/$rel_path"
      done

    EOT
    environment = {
      "API_ENDPOINT" = var.api_endpoint
      "REGION"       = var.region
    }
    interpreter = ["bash", "-c"]
  }
  triggers = {
    "run_always" = plantimestamp()
  }
}

resource "null_resource" "upload_frontend" {
  depends_on = [aws_s3_bucket.frontend_bucket, null_resource.build_frontend]
  provisioner "local-exec" {
    working_dir = "${path.root}/../../chotel-front"
            command     = <<EOT
      aws s3 cp ./tmp_build s3://$BUCKET --recursive
      rm -rf ./tmp_build
    EOT
    environment = {
      "BUCKET" = var.bucket_name
    }
    interpreter = ["bash", "-c"]
  }
  triggers = {
    "run_always" = plantimestamp()
  }
}
