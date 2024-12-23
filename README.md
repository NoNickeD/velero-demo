# Velero Demo

This repository contains a demo for setting up Velero with AWS EKS, including infrastructure setup, application deployment, backup, and restore functionality.

## Prerequisites

- AWS CLI configured with a profile.
- Tofu CLI installed for infrastructure management.
- Kubectl installed and configured for your EKS cluster.
- Taskfile CLI installed for task automation.
- Velero CLI installed.

## Infrastructure Setup

1. Navigate to the `infrastructure` directory:

   ```bash
   cd velero-demo/infrastructure
   ```

2. Initialize OpenTofu:

   ```bash
   tofu init
   ```

3. Plan and apply the infrastructure:

   ```bash
   tofu plan -var-file=./conf/deploy.tfvars
   tofu apply -var-file=./conf/deploy.tfvars
   ```

4. Configure your kubeconfig:

   ```bash
   aws eks --region eu-central-1 update-kubeconfig --name velero-demo --profile sandbox --alias velero-demo
   ```

## Application Deployment

1. Navigate to the `app` directory:

   ```bash
   cd ../app
   ```

2. Build the application locally:

   ```bash
   task build-local
   ```

3. Update the image registry and tag in `log-generator.yaml` (line 35):

   ```bash
   nvim log-generator.yaml
   ```

4. Apply the log generator:

   ```bash
   kubectl apply --filename log-generator.yaml
   ```

## Setting Up Velero

1. Navigate back to the `infrastructure` directory:

   ```bash
   cd ../infrastructure
   ```

2. Run the Velero setup task:

   ```bash
   task setup-velero
   ```

3. Verify Velero setup:

   ```bash
   kubectl --namespace velero logs deployment/velero
   ```

## Backup Management

### Create Scheduled Backups

1. Create a daily backup schedule:

   ```bash
   velero schedule create daily-backup --include-namespaces default --schedule="@daily" --ttl 168h
   ```

2. Verify the schedule:

   ```bash
   kubectl --namespace velero describe schedule daily-backup
   kubectl get schedules -n velero
   ```

### Create On-Demand Backups

1. Create a backup:

   ```bash
   velero backup create test-backup --include-namespaces default
   ```

2. Check the backup logs:

   ```bash
   velero backup logs test-backup
   ```

## Restore Management

### Restore Backups

1. List available backups:

   ```bash
   velero backup get
   ```

2. Restore a backup:

   ```bash
   velero restore create --from-backup test-backup
   ```

3. Check the restore status:

   ```bash
   velero restore get
   ```

4. Describe the restore for more details:

   ```bash
   velero restore describe <restore-name>
   ```

## Destroying the Infrastructure

1. Navigate to the `infrastructure` directory:

   ```bash
   cd velero-demo/infrastructure
   ```

2. Destroy the infrastructure:

   ```bash
   tofu destroy -var-file=./conf/deploy.tfvars
   ```

## Notes

- Ensure that Velero is configured with the correct AWS credentials and permissions.
- Test the backups and restores in a non-production environment before applying to production.
- Modify the `log-generator.yaml` file with your actual container image and registry details.

---

For additional information, refer to the [Velero Documentation](https://velero.io/docs/).
