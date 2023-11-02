---
title: Tailscale Bastion
meta_desc: Create a Tailscale Bastion with Pulumi.
layout: package
---

The Tailscale Bastion package for Pulumi can be used to provision a tailscale router in a variety of cloud providers.

## Example

{{< chooser language "typescript,python" >}}
{{% choosable language typescript %}}

```typescript
import * as lbrlabs from "@lbrlabs/pulumi-tailscale-bastion";
// this creates a bastion for me!
const bastion = new lbrlabs.aws.Bastion("example", {
    vpcId: vpc.vpcId,
    subnetIds: vpc.privateSubnetIds,
    route: "172.20.0.0/22",
    region: "us-west-2",
    instanceType: "t3.small"
})
```

{{% /choosable %}}
{{% choosable language python %}}

```python
import lbrlabs_pulumi_tailscalebastion as lbrlabs

bastion = lbrlabs.aws.Bastion(
    "example",
    vpc_id=vpc.vpc_id,
    subnet_ids=vpc.private_subnet_ids,
    route="172.20.0.0/22",
    region="us-west-2",
)
```

{{% /choosable %}}
{{< /chooser >}}
