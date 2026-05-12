# Firewall topologies on Savannaa Cloud

Two opinionated firewall designs you can deploy with the `sws` Terraform
provider. Pick one based on whether you need **deep inspection** (NVA) or
**SDN-level segmentation** (native firewall via security groups).

```
examples/firewall/
├── native/   ← 4 hubs ↔ 1 spoke, security-group enforcement only
└── nva/      ← 1 hub + NVA appliance ↔ 4 spokes, all traffic inspected
```

---

## 1. Native firewall (`native/`)

Flat peering with east-west security enforced by the platform's native
firewall (security groups). No inspection point.

```
    hub-1 (10.10.1.0/24) ──┐
    hub-2 (10.10.2.0/24) ──┤
                            ├──peering──► spoke (10.10.0.0/24)
    hub-3 (10.10.3.0/24) ──┤
    hub-4 (10.10.4.0/24) ──┘

    Security groups:
      • spoke SG:  ingress TCP 22/80/443 from each hub CIDR
      • hub-N SG:  ingress TCP 22/80/443 from spoke CIDR only
                   (no cross-hub paths — hubs can't reach each other)
```

**When this fits**

- You need L3/L4 segmentation, not packet inspection.
- You want zero-latency between networks (no extra VM hop).
- You don't need NAT, VPN termination, IDS, or L7 firewalling.
- You don't want to operate an extra appliance.

---

## 2. NVA appliance (`nva/`)

Classic enterprise hub-and-spoke where a Network Virtual Appliance lives
on the hub and inspects all cross-spoke + north-south traffic.

```
                    ┌── peering ─► spoke-1 (10.20.1.0/24)
                    │
                    ├── peering ─► spoke-2 (10.20.2.0/24)
   hub (10.20.0.0/24)
    └─► NVA  ───────┤
                    ├── peering ─► spoke-3 (10.20.3.0/24)
                    │
                    └── peering ─► spoke-4 (10.20.4.0/24)

   Security groups:
     • NVA SG:    SSH/0.0.0.0 mgmt + ALL TCP from every spoke CIDR
     • spoke-N SG: ingress TCP from NVA's fixed IP (/32) only
                  (forces every packet through the appliance)
```

The spoke SGs pin ingress to the NVA's `/32`, not the hub CIDR. That
guarantees no future hub-side VM can bypass the appliance.

**When this fits**

- You need L7 inspection, IDS/IPS, VPN termination, or custom NAT.
- You're running a marketplace firewall image (pfSense, OPNsense,
  Fortinet, Sophos, VyOS — see Marketplace > NVA Firewall).
- Centralised east-west policy is more important than a couple ms of
  added latency.

---

## Comparison

|                    | `native/`                | `nva/`                          |
|--------------------|--------------------------|---------------------------------|
| Inspection         | none — pure SDN policy   | deep packet (per your image)   |
| Latency            | minimal, kernel-level    | one extra VM hop                |
| Cost               | $0 (SGs are free)        | NVA flavor × month + storage    |
| Complexity         | low                      | medium-high                      |
| Bypass risk        | n/a                      | None — spoke SGs pin to `/32`   |
| Failure domain     | none                     | NVA goes down → spokes isolated|
| HA path            | n/a                      | run a second NVA + LB (TODO)   |

---

## Prerequisites

1. A **Savannaa account** with API access — get the API URL + key from
   **Account > API Keys** in the console.
2. **Terraform ≥ 1.5** installed locally.
3. (Optional) An SSH public key on disk if you want to log into the NVA
   or any test VMs.

## Step-by-step

```bash
# 1. Choose the topology
cd examples/firewall/native        # or  cd examples/firewall/nva

# 2. Set credentials and region (env-var driven — no hardcoded secrets)
export SWS_API_URL=https://savannaa.com
export SWS_API_KEY=<your-api-key>
export SWS_REGION=ng-lagos-1        # or  ng-abuja-1

# 3. Initialise — downloads the provider on first run
terraform init

# 4. Preview — confirm what will be created
terraform plan

# 5. Apply
#    For native/: no extra vars needed
terraform apply -auto-approve

#    For nva/: pass your SSH key (optional) + a marketplace firewall image
#    (or accept the Ubuntu placeholder and configure iptables yourself).
terraform apply -auto-approve \
  -var "ssh_public_key=$(cat ~/.ssh/id_rsa.pub)" \
  -var "nva_image_name=pfSense 2.7"

# 6. Verify in the console
#    Networking & Delivery > Networks      — confirm the 5 networks
#    Networking & Delivery > Network Peering — confirm the 4 peerings
#    (NVA topology only) Compute > Instances  — confirm the NVA is ACTIVE

# 7. Read the outputs (network IDs, peering IDs, NVA IP)
terraform output

# 8. Tear it down when you're done
terraform destroy -auto-approve
```

## Common gotchas

- **Plan a CIDR scheme up front.** Don't peer two networks whose CIDRs
  overlap; the route tables can't disambiguate.
- **The NVA placeholder is Ubuntu.** It boots but doesn't forward traffic
  until you configure it. For plug-and-play inspection, pass a
  marketplace NVA Firewall image via `-var "nva_image_name=…"`.
- **External egress.** Neither topology configures NAT outbound by
  default. Add an `external_gateway` to your router (or use Public IPs
  on individual VMs) when you need internet egress.
- **Quota.** Each example creates 5 networks + several SGs + 4 peerings.
  Defaults on a new account are 10 networks / 20 SGs / 5 peerings — fine
  for one example at a time, tight if you stand both up side-by-side.

## Cleanup

`terraform destroy` will remove everything in the state file. If you
manually created any extra resources outside Terraform (test VMs,
floating IPs), delete those first in the console — destroy will fail if
a peered network has unmanaged ports attached.
