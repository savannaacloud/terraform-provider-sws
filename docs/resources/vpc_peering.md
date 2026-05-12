---
page_title: "sws_vpc_peering Resource - terraform-provider-sws"
description: |-
  A VPC peering connection.
---

# sws_vpc_peering

A VPC peering connection.

## Example

```hcl
resource "sws_vpc_peering" "to_prod" { name = "dev-to-prod" }
```

## Argument Reference

### Required

- `name` (string) — Connection name.

### Optional

- `config` (string) — JSON-encoded config (peer project + network).


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Peering UUID.
- `status` (string) — Status.


## Import

```
terraform import sws_vpc_peering.<local_name> <id>
```
