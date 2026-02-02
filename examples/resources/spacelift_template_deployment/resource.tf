resource "spacelift_template" "test" {
  name  = "test-template"
  space = "root"
}

resource "spacelift_template_version" "first_version" {
  template_id    = spacelift_template.test.id
  version_number = "1.0.0"
  state          = "PUBLISHED"
  template       = <<-EOT
stacks:
- name: "test-stack-$${{ inputs.env_name }}"
  key: test
  autodeploy: true
  vcs:
    reference:
      value: main
      type: branch
    repository: demo
    provider: GITHUB
  vendor:
    terraform:
      manage_state: true
      version: "1.5.0"
inputs:
- id: env_name
  name: Environment Name
  default: dev
EOT
}

resource "spacelift_template_deployment" "test" {
  template_version_id = spacelift_template_version.first_version.id
  space               = "root"
  name                = "test"
  description         = "description"

  input {
    id        = "env_name"
    value     = "production"
    sensitive = false
  }
}
