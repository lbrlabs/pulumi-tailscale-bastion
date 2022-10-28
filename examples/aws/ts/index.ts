import * as pulumi from "@pulumi/pulumi";
import * as awstailscale from "@lbrlabs/pulumi-aws-tailscalebastion";
import * as awsx from "@pulumi/awsx";

const vpc = new awsx.ec2.Vpc("tailscale", {
  cidrBlock: "172.20.0.0/22",
  tags: {
    Owner: "lbriggs",
    owner: "lbriggs",
    purpose: "infra",
  },
});

export const vpcId = vpc.vpcId

// this creates a bastion for me!
const bastion = new awstailscale.Bastion("example", {
    vpcId: vpc.vpcId,
    subnetIds: vpc.privateSubnetIds,
    route: "172.20.0.0/22",
    region: "us-west-2",
    instanceType: "t3.small"
})
