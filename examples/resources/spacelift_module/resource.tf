resource "spacelift_module" "k8s-module" {
  administrative = true
  branch         = "master"
  description    = "Infra terraform module"
  repository     = "terraform-super-module"
}
