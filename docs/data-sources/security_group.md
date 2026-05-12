---
page_title: "sws_security_group Data Source - terraform-provider-sws"
description: |-
  Look up an existing security group by name.
---

# sws_security_group (Data Source)

Look up an existing security group by name.

## Example

```hcl
data "sws_security_group" "default" {
  name = "default"
}
```

## Argument Reference

### Required

- `name` (string) — Security group name.


## Attribute Reference

### Read-only

- `id` (string) — Security group UUID.

