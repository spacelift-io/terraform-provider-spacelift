resource "spacelift_bitbucket_datacenter_integration" "example" {
  name             = "Bitbucket integration"
  is_default       = false
  space_id         = "root"
  api_host         = "private://bitbucket_spacelift/bitbucket"
  user_facing_host = "https://bitbucket.spacelift.io/bitbucket"
  username         = "bitbucket_user_name"
  access_token     = "ABCD-MDQ3NzgxMzg3NzZ0VpOejJZmQfBBlpxxJuK9j1LLQG8g"
}