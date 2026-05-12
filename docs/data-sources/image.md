---
page_title: "sws_image Data Source - terraform-provider-sws"
description: |-
  Look up an image by name.
---

# sws_image (Data Source)

Look up an image by name.

## Example

```hcl
data "sws_image" "ubuntu" {
  name = "Ubuntu 22.04 LTS"
}
```

## Argument Reference

### Required

- `name` (string) — Image name (exact match).


## Attribute Reference

### Read-only

- `id` (string) — Image UUID.

