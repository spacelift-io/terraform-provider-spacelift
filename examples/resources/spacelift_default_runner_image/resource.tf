resource "spacelift_default_runner_image" "example" {
  public  = "registry.example.com/my-public-image:latest"
  private = "registry.example.com/my-private-image:latest"
}
