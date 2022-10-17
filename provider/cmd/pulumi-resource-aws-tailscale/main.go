//go:generate go run ./generate.go

package main

import (
	"github.com/lbrlabs/pulumi-aws-tailscalebastion/pkg/provider"
	"github.com/lbrlabs/pulumi-aws-tailscalebastion/pkg/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	providerName = "aws-tailscalebastion"
)

func main() {
	kingpin.Version(version.Version)
	kingpin.Parse()
	provider.Serve(providerName, version.Version, pulumiSchema)
}
