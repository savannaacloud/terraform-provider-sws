---
page_title: "sws_security_group Resource - terraform-provider-sws"
description: |-
  A network security group (stateful firewall).
---

# sws_security_group

A network security group (stateful firewall).

## Example

```hcl
resource "sws_security_group" "web" {
  name        = "web"
  description = "allow 80/443"
}
```

## Argument Reference

### Required

- `name` (string) — Display name.

### Optional

- `description` (string) — Free-form description.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Security group UUID.


## Import

```
terraform import sws_security_group.<local_name> <id>
```
