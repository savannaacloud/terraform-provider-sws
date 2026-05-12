---
page_title: "sws_network Resource - terraform-provider-sws"
description: |-
  A tenant network with an optional inline IPv4 subnet.
---

# sws_network

A tenant network with an optional inline IPv4 subnet.

## Example

```hcl
resource "sws_network" "app" {
  name = "app-net"
  cidr = "10.0.0.0/24"   # optional, this is the default. Set "" to skip subnet.
}
```

## Argument Reference

### Required

- `name` (string) — Display name.

### Optional

- `description` (string) — Free-form description.
- `cidr` (string) — IPv4 CIDR for the inline subnet. Default `10.0.0.0/24`. Pass `""` to skip subnet creation.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Network UUID.
- `subnet_id` (string) — ID of the inline subnet, if one was created.


## Import

```
terraform import sws_network.<local_name> <id>
```
