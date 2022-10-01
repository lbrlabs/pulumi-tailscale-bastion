//go:generate go run ./generate.go

package main

import (
	"github.com/lbrlabs/pulumi-aws-tailscalebastion/pkg/provider"
	"github.com/lbrlabs/pulumi-aws-tailscalebastion/pkg/version"
)

var providerName = "aws-tailscalebastion"

func main() {
	provider.Serve(providerName, version.Version, pulumiSchema)
}
