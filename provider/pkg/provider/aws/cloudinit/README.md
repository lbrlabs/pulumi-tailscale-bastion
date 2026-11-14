# Upstream cloud-init scripts

Vendored unchanged from https://github.com/tailscale/terraform-cloudinit-tailscale
at commit `342bd51fea138784845c3a879150d8dacb94d816` (see LICENSE).

`files/{ip_forwarding,udp_offloads}.sh` and
`templates/{install_tailscale,setup_tailscale}.sh.tmpl` are embedded in the
provider. `../cloudinit.go` renders their Terraform-style interpolations and
assembles MIME cloud-init parts without requiring Terraform at deploy time.
The optional daemon configuration template is omitted because this component
uses the default tailscaled configuration.

To update, copy these four files and LICENSE from a reviewed upstream commit,
update this pin, and run the AWS provider tests. The renderer rejects unknown
template variables. The AWS wrapper fetches legacy credentials from SSM at boot,
removes the temporary credential file, and fails if authentication retries exhaust
or a `tailscale set` command fails.

The AWS Bastion exposes the installation and setup options as inputs:

| Inputs | Defaults |
| --- | --- |
| `track`, `maxRetries`, `retryDelay` | `stable`, `10`, `10` seconds |
| `json`, `reset`, `forceReauth` | `false` |
| `loginServer`, `timeout` | `https://controlplane.tailscale.com`, `60s` |
| `acceptDNS`, `snatSubnetRoutes` | `true` |
| `acceptRoutes`, `shieldsUp`, `statefulFiltering`, `exitNodeAllowLANAccess` | `false` |
| `netfilterMode` | `on` (`off` and `nodivert` also supported) |
| `exitNode`, `operator` | empty (unset) |

`hostname`, `tailscaleTags`, `routes`, `enableSSH`, `enableExitNode`,
`enableAppConnector`, and `peerRelaySettings` supply the remaining node settings.
Authentication variables are managed through `oauthClientSecret`,
`workloadIdentity`, or the automatically generated tailnet key. AWS workload
identity obtains its token automatically rather than accepting a static ID token.
The default generated tailnet key requires the Tailscale provider; a custom
`loginServer` must support the selected authentication mechanism.
