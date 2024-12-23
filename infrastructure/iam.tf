#--------------------------------
# IAM Roles and Policies
#--------------------------------
resource "aws_iam_policy" "eks_full_access" {
  name        = "EKSFullAccessPolicy"
  description = "Full access to EKS Cluster resources"
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "eks:*",
          "ec2:*",
          "iam:PassRole"
        ],
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_policy" "velero_policy" {
  name        = "VeleroPolicy"
  description = "IAM Policy granting only the necessary permissions for Velero"
  policy      = data.aws_iam_policy_document.velero_policy.json
}
