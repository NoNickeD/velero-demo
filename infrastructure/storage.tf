module "velero_s3_bucket" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = "4.2.2"

  bucket        = var.velero_bucket_name
  acl           = null
  force_destroy = true

  object_ownership = "BucketOwnerEnforced"

  tags = merge(local.tags, { Name = "${var.velero_bucket_name}" })

  versioning = {
    enabled = true
  }
}
