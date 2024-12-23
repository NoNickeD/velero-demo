output "service_account_name" {
  value = kubernetes_service_account.velero.metadata[0].name
}

output "service_account_namespace" {
  value = kubernetes_service_account.velero.metadata[0].namespace
}

output "irsa_role_arn" {
  value = module.irsa_velero.iam_role_arn
}
