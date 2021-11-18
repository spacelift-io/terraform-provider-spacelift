resource "spacelift_stack" "k8s-core" {
  branch     = "master"
  name       = "Kubernetes core services"
  repository = "core-infra"
}

resource "spacelift_stack_gcp_service_account" "k8s-core" {
  stack_id = spacelift_stack.k8s-core.id

  token_scopes = [
    "https://www.googleapis.com/auth/compute",
    "https://www.googleapis.com/auth/cloud-platform",
    "https://www.googleapis.com/auth/devstorage.full_control",
  ]
}

resource "google_project" "k8s-core" {
  name       = "Kubernetes code"
  project_id = "unicorn-k8s-core"
  org_id     = var.gcp_organization_id
}

resource "google_project_iam_member" "k8s-core" {
  project = google_project.k8s-core.id
  role    = "roles/owner"
  member  = spacelift_stack_gcp_service_account.k8s-core.service_account_email
}
