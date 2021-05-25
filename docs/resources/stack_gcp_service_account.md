---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "spacelift_stack_gcp_service_account Resource - terraform-provider-spacelift"
subcategory: ""
description: |-
  spacelift_gcp_service_account represents a Google Cloud Platform service account that's linked to a particular Stack or Module. These accounts are created by Spacelift on per-stack basis, and can be added as members to as many organizations and projects as needed. During a Run or a Task, temporary credentials for those service accounts are injected into the environment, which allows credential-less GCP Terraform provider setup.
---

# spacelift_stack_gcp_service_account (Resource)

`spacelift_gcp_service_account` represents a Google Cloud Platform service account that's linked to a particular Stack or Module. These accounts are created by Spacelift on per-stack basis, and can be added as members to as many organizations and projects as needed. During a Run or a Task, temporary credentials for those service accounts are injected into the environment, which allows credential-less GCP Terraform provider setup.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **token_scopes** (Set of String) List of scopes that will be requested when generating temporary GCP service account credentials

### Optional

- **id** (String) The ID of this resource.
- **module_id** (String) ID of the module which uses GCP service account credentials
- **stack_id** (String) ID of the stack which uses GCP service account credentials

### Read-Only

- **service_account_email** (String) Email address of the GCP service account dedicated for this stack

