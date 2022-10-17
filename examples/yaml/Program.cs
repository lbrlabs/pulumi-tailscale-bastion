using System.Collections.Generic;
using Pulumi;
using AwsTailscale = Lbrlabs.PulumiPackage.AwsTailscale;
using Awsx = Pulumi.Awsx;

return await Deployment.RunAsync(() => 
{
    var vpc = new Awsx.Ec2.Vpc("vpc", new()
    {
        CidrBlock = "172.20.0.0/22",
    });

    var bastion = new AwsTailscale.Bastion("bastion", new()
    {
        VpcId = vpc.VpcId,
        SubnetIds = vpc.PrivateSubnetIds,
        Route = "172.20.0.0/22",
        Region = "us-west-2",
    });

    return new Dictionary<string, object?>
    {
        ["url"] = vpc.VpcId,
    };
});

