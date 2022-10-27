//go:generate go run ./generate.go

package main

import (
	"github.com/lbrlabs/pulumi-tailscale-bastion/pkg/provider"
	"github.com/lbrlabs/pulumi-tailscale-bastion/pkg/version"
)

var (
	providerName = "tailscale-bastion"
)

func main() {
	provider.Serve(providerName, version.Version, pulumiSchema)
}
