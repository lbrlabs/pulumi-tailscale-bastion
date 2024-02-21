"""A Python Pulumi program"""

import pulumi
import pulumi_awsx as awsx
import lbrlabs_pulumi_tailscalebastion as tailscale

vpc = awsx.ec2.Vpc(
    "example",
    cidr_block="172.20.0.0/22",
)

bastion = tailscale.aws.Bastion(
    "example",
    vpc_id=vpc.vpc_id,
    subnet_ids=vpc.public_subnet_ids,
    route="172.20.0.0/22",
    tailscale_tags=["tag:bastion"],
    region="us-west-2",
    high_availability=True,
    enable_ssh=True,
    public=True,
)


pulumi.export("vpcId", vpc.vpc_id)
