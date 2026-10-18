# Pulumi Tailscale Bastion

This repo provides a [multi-language](https://www.pulumi.com/blog/pulumiup-pulumi-packages-multi-language-components/) component that creates a [Tailscale](https://tailscale.com/) [Subnet Router](https://tailscale.com/kb/1019/subnets/) in your chosen cloud provider

## Examples

Examples for all languages are in the [examples](examples/) directory. 

Note, you need to create a VPC, and also add your worker nodes. 

## FAQs

### AWS workload identity federation

AWS bastions can join Tailscale using their EC2 instance role, without storing
an OAuth secret or auth key:

```typescript
const bastion = new tailscalebastion.aws.Bastion("router", {
    vpcId: vpc.id,
    subnetIds: privateSubnetIds,
    region: "us-west-2",
    tailscaleTags: ["tag:router"],
    routes: ["10.0.0.0/16"],
    workloadIdentity: {
        clientId: config.require("tailscaleClientId"),
        audience: config.require("tailscaleAudience"),
    },
});
export const roleArn = bastion.roleArn;
```

Configure a [Tailscale federated identity](https://tailscale.com/kb/1581/workload-identity-federation)
that trusts your AWS token issuer and matches the instance role's ARN as its
subject. Grant it the `auth_keys` scope and the advertised tags. AWS outbound
identity federation must be enabled/configured for your account. The client ID
and audience are public identifiers; `roleArn` is exported to help configure
the trust. Establish that trust before instances boot (or refresh the ASG after
configuring it).

The component grants `sts:GetWebIdentityToken` for the configured audience and
a maximum token lifetime of 300 seconds. Tailscale discovers the instance role
and obtains its token at boot using `--client-id` and `--audience` (requires
Tailscale 1.94+). Nodes remain ephemeral and preauthorized.

`workloadIdentity` and `oauthClientSecret` are mutually exclusive. Omitting both
retains provider-created auth keys. AWS bootstrap uses pinned, embedded scripts
from [terraform-cloudinit-tailscale](https://github.com/tailscale/terraform-cloudinit-tailscale),
installs the current stable Tailscale release, and supports both CPU architectures.

### Can you add support for X

Add an issue, but this is mainly designed to be useful for cloud providers I use, so I reserve the right to refuse.

### Can you make X optional?

I have no plans to make any of the batteries included optional at this time

## Installing

This package is available in many languages in the standard packaging formats.

### Node.js (Java/TypeScript)

To use from JavaScript or TypeScript in Node.js, install using either `npm`:

```
$ npm install @lbrlabs/pulumi-tailscalebastion
```

or `yarn`:

```
$ yarn add @lbrlabs/pulumi-tailscalebastion
```

### Python

To use from Python, install using `pip`:

```
$ pip install lbrlabs_pulumi_tailscalebastion
```

### Go

To use from Go, use `go get` to grab the latest version of the library

```
$ go get github.com/lbrlabs/pulumi-tailscale-bastion/sdk/go/...
```

### .NET

To use from Dotnet, use `dotnet add package` to install into your project. You must specify the version if it is a pre-release version.


```
$ dotnet add package Lbrlabs.PulumiPackage.TailscaleBastion
```

## Reference

See the Pulumi registry for API docs:

https://www.pulumi.com/registry/packages/lbrlabs-tailscale-bastion/api-docs/
