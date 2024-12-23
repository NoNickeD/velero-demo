terraform {
  required_providers {
    kubernetes = {
      source = "hashicorp/kubernetes"
    }
  }
}


resource "kubernetes_namespace" "velero" {
  metadata {
    name = var.namespace
  }
}

# IRSA Role for Velero
module "irsa_velero" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-assumable-role-with-oidc"
  version = "5.39.0"

  create_role = true
  role_name   = "VeleroRole-${var.cluster_name}"

  # The EKS OIDC provider URL
  provider_url = var.provider_url

  # Attach the pre-created Velero Policy
  role_policy_arns = [
    var.velero_policy_arn
  ]

  # Only allow the Velero service account in this namespace
  oidc_fully_qualified_subjects = [
    "system:serviceaccount:${var.namespace}:${var.service_account_name}"
  ]
}

# Create the Velero service account in Kubernetes
resource "kubernetes_service_account" "velero" {
  depends_on = [module.irsa_velero] # Ensure role is created before SA references it

  metadata {
    name      = var.service_account_name
    namespace = kubernetes_namespace.velero.metadata[0].name

    # Annotate so the service account can assume the IRSA role
    annotations = {
      "eks.amazonaws.com/role-arn" = module.irsa_velero.iam_role_arn
    }
  }
}
