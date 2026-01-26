# Get documentation for aws_instance resource
data "registry_provider_docs" "aws_instance" {
  namespace = "hashicorp"
  name      = "aws"
  version   = "5.0.0"
  doc_type  = "resources"
  doc_name  = "instance"
}

output "aws_instance_docs" {
  value = {
    title       = data.registry_provider_docs.aws_instance.title
    category    = data.registry_provider_docs.aws_instance.category
    subcategory = data.registry_provider_docs.aws_instance.subcategory
    # content contains the full markdown documentation
    content_preview = substr(data.registry_provider_docs.aws_instance.content, 0, 500)
  }
}

# Get documentation for a data source
data "registry_provider_docs" "aws_ami" {
  namespace = "hashicorp"
  name      = "aws"
  version   = "5.0.0"
  doc_type  = "data-sources"
  doc_name  = "ami"
}

output "aws_ami_docs_title" {
  value = data.registry_provider_docs.aws_ami.title
}

# Get documentation for S3 bucket
data "registry_provider_docs" "s3_bucket" {
  namespace = "hashicorp"
  name      = "aws"
  version   = "5.0.0"
  doc_type  = "resources"
  doc_name  = "s3_bucket"
}

output "s3_docs_subcategory" {
  value = data.registry_provider_docs.s3_bucket.subcategory
}

# Write documentation to a local file (useful for doc generation)
resource "local_file" "aws_instance_docs" {
  content  = data.registry_provider_docs.aws_instance.content
  filename = "${path.module}/output/aws_instance.md"
}
