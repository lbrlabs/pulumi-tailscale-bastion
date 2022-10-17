"""A Python Pulumi program"""

import pulumi
import pulumi_awsx as awsx
import lbrlabs_pulumi_aws_tailscalebastion as lbrlabs

vpc = awsx.ec2.Vpc(
    "example",
    cidr_block="172.20.0.0/22",
)

bastion = lbrlabs.Bastion(
    "example",
    vpc_id=vpc.vpc_id,
    subnet_ids=vpc.private_subnet_ids,
    route="172.20.0.0/22",
    region="us-west-2",
)


pulumi.export("vpcId", vpc.id)
