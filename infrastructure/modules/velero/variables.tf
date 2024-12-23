variable "cluster_name" {
  type        = string
  description = "Name of the EKS cluster for labeling the IRSA role"
}

variable "provider_url" {
  type        = string
  description = "The EKS OIDC provider URL (e.g. from module.eks.oidc_provider)"
}

variable "velero_policy_arn" {
  type        = string
  description = "ARN of the AWS IAM Policy for Velero"
}

variable "namespace" {
  type        = string
  description = "Kubernetes namespace for Velero"
  default     = "velero"
}

variable "service_account_name" {
  type        = string
  description = "Name of the Velero service account"
  default     = "velero"
}

