---
page_title: "sws_vault_secret Resource - terraform-provider-sws"
description: |-
  A secret stored in the project vault.
---

# sws_vault_secret

A secret stored in the project vault.

## Example

```hcl
resource "sws_vault_secret" "db_password" { name = "db_password" }
```

## Argument Reference

### Required

- `name` (string) — Secret name.

### Optional

- `config` (string) — JSON-encoded payload.


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Secret UUID.
- `status` (string) — Status.


## Import

```
terraform import sws_vault_secret.<local_name> <id>
```
