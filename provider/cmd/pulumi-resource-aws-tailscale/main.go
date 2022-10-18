//go:generate go run ./generate.go

package main

import (
	"github.com/lbrlabs/pulumi-aws-tailscale/pkg/provider"
	"github.com/lbrlabs/pulumi-aws-tailscale/pkg/version"
)

var (
	providerName = "aws-tailscale"
)

func main() {
	provider.Serve(providerName, version.Version, pulumiSchema)
}
