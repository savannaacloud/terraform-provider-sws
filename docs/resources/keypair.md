---
page_title: "sws_keypair Resource - terraform-provider-sws"
description: |-
  An SSH keypair for instance access.
---

# sws_keypair

An SSH keypair for instance access.

## Example

```hcl
resource "sws_keypair" "admin" {
  name       = "admin"
  public_key = file("~/.ssh/id_rsa.pub")
}
```

## Argument Reference

### Required

- `name` (string) — Keypair name. Used as the ID.
- `public_key` (string) — OpenSSH-format public key.

### Optional

_None._


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Same as name.
- `fingerprint` (string) — Key fingerprint.


## Import

```
terraform import sws_keypair.<local_name> <id>
```
