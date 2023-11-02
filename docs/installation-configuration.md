---
title: Tailscale Bastion Installation & Configuration
meta_desc: Information on how to install and use the Tailscale Bastion Provider.
layout: package
---

## Installation

The Pulumi Grafana provider is available as a package in all Pulumi languages:

* JavaScript/TypeScript: [`@lbrlabs/pulumi-tailscalebastion`](https://www.npmjs.com/package/@lbrlabs/pulumi-tailscalebastion)
* Python: [`lbrlabs_pulumi_tailscalebastion`](https://pypi.org/project/lbrlabs-pulumi-tailscalebastion//)
* Go: [`github.com/lbrlabs/pulumi-tailscale-bastion/sdk/go/grafana`](https://pkg.go.dev/github.com/lbrlabs/pulumi-tailscale-bastion/sdk)
* .NET: [`Lbrlabs.PulumiPackage.TailscaleBastion`](https://www.nuget.org/packages/Lbrlabs.PulumiPackage.TailscaleBastion)

### Provider Binary

The Tailscale Bastion provider binary is a third party binary. It can be installed using the `pulumi plugin` command.

```bash
pulumi plugin install resource tailscale-bastion --server github://api.github.com/lbrlabs
```

## Setup

To provision resources with the Pulumi Tailscale bastion provider, you'll need tailscale credentials and credentials for the cloud provider you're provisioning the bastion in.

Please visit the Pulumi registry for your chosen cloud provider to learn how to provision credentials:

Azure: https://www.pulumi.com/registry/packages/azure/
AWS: https://www.pulumi.com/registry/packages/aws/
Kubernetes: https://www.pulumi.com/registry/packages/kubernetes/
GCP: Coming soon
