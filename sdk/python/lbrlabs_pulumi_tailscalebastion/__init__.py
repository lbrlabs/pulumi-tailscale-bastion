# coding=utf-8
# *** WARNING: this file was generated by Pulumi SDK Generator. ***
# *** Do not edit by hand unless you're certain you know what you are doing! ***

from . import _utilities
import typing
# Export this package's modules as members:
from .provider import *

# Make subpackages available:
if typing.TYPE_CHECKING:
    import lbrlabs_pulumi_tailscalebastion.aws as __aws
    aws = __aws
    import lbrlabs_pulumi_tailscalebastion.azure as __azure
    azure = __azure
    import lbrlabs_pulumi_tailscalebastion.kubernetes as __kubernetes
    kubernetes = __kubernetes
else:
    aws = _utilities.lazy_import('lbrlabs_pulumi_tailscalebastion.aws')
    azure = _utilities.lazy_import('lbrlabs_pulumi_tailscalebastion.azure')
    kubernetes = _utilities.lazy_import('lbrlabs_pulumi_tailscalebastion.kubernetes')

_utilities.register(
    resource_modules="""
[
 {
  "pkg": "tailscale-bastion",
  "mod": "aws",
  "fqn": "lbrlabs_pulumi_tailscalebastion.aws",
  "classes": {
   "tailscale-bastion:aws:Bastion": "Bastion"
  }
 },
 {
  "pkg": "tailscale-bastion",
  "mod": "azure",
  "fqn": "lbrlabs_pulumi_tailscalebastion.azure",
  "classes": {
   "tailscale-bastion:azure:Bastion": "Bastion"
  }
 },
 {
  "pkg": "tailscale-bastion",
  "mod": "kubernetes",
  "fqn": "lbrlabs_pulumi_tailscalebastion.kubernetes",
  "classes": {
   "tailscale-bastion:kubernetes:Bastion": "Bastion"
  }
 }
]
""",
    resource_packages="""
[
 {
  "pkg": "tailscale-bastion",
  "token": "pulumi:providers:tailscale-bastion",
  "fqn": "lbrlabs_pulumi_tailscalebastion",
  "class": "Provider"
 }
]
"""
)
