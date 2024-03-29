---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "spacelift_ips Data Source - terraform-provider-spacelift"
subcategory: ""
description: |-
  spacelift_ips returns the list of Spacelift's outgoing IP addresses, which you can use to whitelist connections coming from the Spacelift's "mothership". NOTE: this does not include the IP addresses of the workers in Spacelift's public worker pool. If you need to ensure that requests made during runs originate from a known set of IP addresses, please consider setting up a private worker pool https://docs.spacelift.io/concepts/worker-pools.
---

# spacelift_ips (Data Source)

`spacelift_ips` returns the list of Spacelift's outgoing IP addresses, which you can use to whitelist connections coming from the Spacelift's "mothership". **NOTE:** this does not include the IP addresses of the workers in Spacelift's public worker pool. If you need to ensure that requests made during runs originate from a known set of IP addresses, please consider setting up a [private worker pool](https://docs.spacelift.io/concepts/worker-pools).

## Example Usage

```terraform
data "spacelift_ips" "ips" {}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `id` (String) The ID of this resource.
- `ips` (Set of String) the list of spacelift.io outgoing IP addresses
