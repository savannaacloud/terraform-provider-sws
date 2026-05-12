---
page_title: "sws_backup_policy Resource - terraform-provider-sws"
description: |-
  A managed backup policy.
---

# sws_backup_policy

A managed backup policy.

## Example

```hcl
resource "sws_backup_policy" "nightly" { name = "nightly" }
```

## Argument Reference

### Required

- `name` (string) — Policy name.

### Optional

- `config` (string) — JSON-encoded config (schedule, retention, targets).


## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` (string) — Policy UUID.
- `status` (string) — Status.


## Import

```
terraform import sws_backup_policy.<local_name> <id>
```
