# Pulumi Tailscale Bastion

This repo provides a [multi-language](https://www.pulumi.com/blog/pulumiup-pulumi-packages-multi-language-components/) component that creates a [Tailscale](https://tailscale.com/) [Subnet Router](https://tailscale.com/kb/1019/subnets/) in your chosen cloud provider

## Examples

Examples for all languages are in the [examples](examples/) directory. 

Note, you need to create a VPC, and also add your worker nodes. 

## FAQs

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