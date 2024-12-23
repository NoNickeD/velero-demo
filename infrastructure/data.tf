data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_eks_cluster_auth" "cluster" {
  name = module.eks.cluster_name
}

data "aws_iam_policy" "ebs_csi_policy" {
  arn = "arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy"
}

# data "aws_iam_policy_document" "assume_role" {
#   statement {
#     effect = "Allow"

#     principals {
#       type        = "Service"
#       identifiers = ["eks.amazonaws.com", "ec2.amazonaws.com", "eks-fargate-pods.amazonaws.com", "eks-nodegroup.amazonaws.com", "lambda.amazonaws.com"]
#     }
#     actions = ["sts:AssumeRole"]
#   }
# }

data "aws_ssm_parameter" "eks_ami_release_version" {
  name = "/aws/service/eks/optimized-ami/${module.eks.cluster_version}/amazon-linux-2/recommended/release_version"
}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

data "aws_iam_policy_document" "velero_policy" {
  statement {
    actions = [
      "ec2:DescribeVolumes",
      "ec2:DescribeSnapshots",
      "ec2:CreateTags",
      "ec2:CreateVolume",
      "ec2:CreateSnapshot",
      "ec2:DeleteSnapshot"
    ]
    resources = ["*"]
  }

  # Restrict S3 actions
  statement {
    actions = [
      "s3:PutObject",
      "s3:GetObject",
      "s3:DeleteObject",
      "s3:AbortMultipartUpload"
    ]
    resources = [
      "arn:aws:s3:::${var.velero_bucket_name}/*"
    ]
  }

  # For listing the bucket
  statement {
    actions   = ["s3:ListBucket"]
    resources = ["arn:aws:s3:::${var.velero_bucket_name}"]
  }
}

