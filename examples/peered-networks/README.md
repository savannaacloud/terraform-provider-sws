# Two peered networks

Minimal VPC peering example: net-a ↔ net-b with one Ubuntu test VM in each.

```
   net-a (10.30.1.0/24) ◄──peering──► net-b (10.30.2.0/24)
        │                                  │
        └──► vm-a                          └──► vm-b
```

## What gets created

| Resource | Count |
|---|---|
| `sws_network` | 2 |
| `sws_vpc_peering` | 1 |
| `sws_security_group` | 1 (shared) |
| `sws_security_group_rule` | 2 (ICMP intra-/16, SSH /0) |
| `sws_instance` | 2 |

## Apply

```bash
export SWS_API_URL=https://savannaa.com
export SWS_API_KEY=...
export SWS_REGION=ng-lagos-1
terraform init
terraform apply -var "ssh_public_key=$(cat ~/.ssh/id_rsa.pub)"
```

## Verify

```bash
ssh ubuntu@$(terraform output -raw vm_a_ip)
ping <vm-b-fixed-ip>     # should succeed across the peering
```

The `vm_b` private IP shows in the Compute > Instances detail page. The
peering allows the entire `10.30.0.0/16` range so both halves can reach
each other.
