---
page_title: "sws_bastion Resource - terraform-provider-sws"
description: |-
  A bastion host for SSH access into private subnets.
---

# sws_bastion

A bastion host for SSH access into private subnets.

## Example

```hcl
resource "sws_bastion" "ops" { name = "ops-bastion" }
```

## Argument Reference

### Required

- `name` (string) — Bastion name.

### Optional

- `config` (string) — JSON-encoded config.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Bastion UUID.
- `status` (string) — Status.


## Import

```
terraform import sws_bastion.<local_name> <id>
```
