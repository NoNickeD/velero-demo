terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.30"
    }
    helm = {
      source  = "hashicorp/helm"
      version = ">= 2.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = ">= 4.0"
    }
  }
}

# AWS Provider Configuration
provider "aws" {
  region  = var.region
  profile = var.profile
}

provider "kubernetes" {
  host                   = module.eks.cluster_endpoint
  cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)

  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args = [
      "eks", "get-token",
      "--cluster-name", module.eks.cluster_name,
      "--region", var.region,
      "--profile", var.profile
    ]
  }
}


# Helm Provider Configuration for managing Kubernetes resources with Helm
provider "helm" {
  kubernetes {
    host                   = module.eks.cluster_endpoint
    token                  = data.aws_eks_cluster_auth.cluster.token
    cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)
  }
}

# TLS Provider Configuration (for certificates, if needed in other parts of the configuration)
provider "tls" {}


module "velero" {
  source               = "./modules/velero"
  cluster_name         = module.eks.cluster_name
  provider_url         = module.eks.oidc_provider
  velero_policy_arn    = aws_iam_policy.velero_policy.arn
  namespace            = "velero" # optional override
  service_account_name = "velero" # optional override
}
