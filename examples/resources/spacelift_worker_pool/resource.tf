resource "spacelift_worker_pool" "k8s-core" {
  name        = "Main worker"
  csr         = filebase64("/path/to/csr")
  description = "Used for all type jobs"
}
