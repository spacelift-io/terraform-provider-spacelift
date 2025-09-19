resource "spacelift_worker_pool" "production" {
  name = "production-workers"
  csr  = filebase64("/path/to/csr")
}

resource "spacelift_worker_pool_recycle" "monthly_refresh" {
  worker_pool_id = spacelift_worker_pool.production.id

  # Keepers map that triggers recycle when values change
  # Using a monthly timestamp ensures workers get recycled monthly
  keepers = {
    month = formatdate("YYYY-MM", timestamp())

    # You can also trigger manual recycles by changing this value
    manual_trigger = "initial"

    # Or specify a self-hosted version to trigger a recycle when the version changes
    self_hosted_version = var.self_hosted_version
  }
}
