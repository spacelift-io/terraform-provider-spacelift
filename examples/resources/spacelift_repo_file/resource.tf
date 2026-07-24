resource "spacelift_repo_file" "main" {
  repo_id = spacelift_repo.this.id
  path    = "modules/vpc/main.tf"
  content = file("${path.module}/src/vpc.tf")

  commit_message = "Sync the VPC module"
  author_name    = "Terraform"
}

# Spacelift never returns the contents of an encrypted file, so Terraform cannot
# detect changes made to it outside of Terraform.
resource "spacelift_repo_file" "secrets" {
  repo_id = spacelift_repo.this.id
  path    = "secrets.auto.tfvars"
  content = "token = \"bacon\""
  encrypt = true
}
