# coding=utf-8
# *** WARNING: this file was generated by Pulumi SDK Generator. ***
# *** Do not edit by hand unless you're certain you know what you are doing! ***

import copy
import warnings
import sys
import pulumi
import pulumi.runtime
from typing import Any, Mapping, Optional, Sequence, Union, overload
if sys.version_info >= (3, 11):
    from typing import NotRequired, TypedDict, TypeAlias
else:
    from typing_extensions import NotRequired, TypedDict, TypeAlias
from .. import _utilities

__all__ = ['BastionArgs', 'Bastion']

@pulumi.input_type
class BastionArgs:
    def __init__(__self__, *,
                 high_availability: Optional[pulumi.Input[bool]] = None,
                 region: pulumi.Input[str],
                 subnet_ids: pulumi.Input[Sequence[pulumi.Input[str]]],
                 tailscale_tags: pulumi.Input[Sequence[pulumi.Input[str]]],
                 vpc_id: pulumi.Input[str],
                 enable_app_connector: Optional[pulumi.Input[bool]] = None,
                 enable_exit_node: Optional[pulumi.Input[bool]] = None,
                 enable_ssh: Optional[pulumi.Input[bool]] = None,
                 hostname: Optional[pulumi.Input[str]] = None,
                 instance_type: Optional[pulumi.Input[str]] = None,
                 oauth_client_secret: Optional[pulumi.Input[str]] = None,
                 public: Optional[pulumi.Input[bool]] = None,
                 routes: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]] = None):
        """
        The set of arguments for constructing a Bastion resource.
        :param pulumi.Input[bool] high_availability: Whether the bastion should be highly available.
        :param pulumi.Input[str] region: The AWS region you're using.
        :param pulumi.Input[Sequence[pulumi.Input[str]]] subnet_ids: The subnet Ids to launch instances in.
        :param pulumi.Input[Sequence[pulumi.Input[str]]] tailscale_tags: The tags to apply to the tailnet device andauth key. This tag should be added to your oauth key and ACL.
        :param pulumi.Input[str] vpc_id: The VPC the Bastion should be created in.
        :param pulumi.Input[bool] enable_app_connector: Whether the bastion advertises itself as an app connector.
        :param pulumi.Input[bool] enable_exit_node: Whether the subnet router can advertise itself as an exit node.
        :param pulumi.Input[bool] enable_ssh: Whether to enable SSH access to the bastion.
        :param pulumi.Input[str] hostname: The hostname of the bastion.
        :param pulumi.Input[str] instance_type: The EC2 instance type to use for the bastion.
        :param pulumi.Input[str] oauth_client_secret: An OAuth Client Secret to use for authenticating Tailscale clients.
        :param pulumi.Input[bool] public: Whether the bastion is going in public subnets.
        :param pulumi.Input[Sequence[pulumi.Input[str]]] routes: The routes you'd like to advertise via tailscale.
        """
        if high_availability is None:
            high_availability = False
        pulumi.set(__self__, "high_availability", high_availability)
        pulumi.set(__self__, "region", region)
        pulumi.set(__self__, "subnet_ids", subnet_ids)
        pulumi.set(__self__, "tailscale_tags", tailscale_tags)
        pulumi.set(__self__, "vpc_id", vpc_id)
        if enable_app_connector is None:
            enable_app_connector = False
        if enable_app_connector is not None:
            pulumi.set(__self__, "enable_app_connector", enable_app_connector)
        if enable_exit_node is None:
            enable_exit_node = False
        if enable_exit_node is not None:
            pulumi.set(__self__, "enable_exit_node", enable_exit_node)
        if enable_ssh is None:
            enable_ssh = True
        if enable_ssh is not None:
            pulumi.set(__self__, "enable_ssh", enable_ssh)
        if hostname is not None:
            pulumi.set(__self__, "hostname", hostname)
        if instance_type is not None:
            pulumi.set(__self__, "instance_type", instance_type)
        if oauth_client_secret is not None:
            pulumi.set(__self__, "oauth_client_secret", oauth_client_secret)
        if public is None:
            public = False
        if public is not None:
            pulumi.set(__self__, "public", public)
        if routes is not None:
            pulumi.set(__self__, "routes", routes)

    @property
    @pulumi.getter(name="highAvailability")
    def high_availability(self) -> pulumi.Input[bool]:
        """
        Whether the bastion should be highly available.
        """
        return pulumi.get(self, "high_availability")

    @high_availability.setter
    def high_availability(self, value: pulumi.Input[bool]):
        pulumi.set(self, "high_availability", value)

    @property
    @pulumi.getter
    def region(self) -> pulumi.Input[str]:
        """
        The AWS region you're using.
        """
        return pulumi.get(self, "region")

    @region.setter
    def region(self, value: pulumi.Input[str]):
        pulumi.set(self, "region", value)

    @property
    @pulumi.getter(name="subnetIds")
    def subnet_ids(self) -> pulumi.Input[Sequence[pulumi.Input[str]]]:
        """
        The subnet Ids to launch instances in.
        """
        return pulumi.get(self, "subnet_ids")

    @subnet_ids.setter
    def subnet_ids(self, value: pulumi.Input[Sequence[pulumi.Input[str]]]):
        pulumi.set(self, "subnet_ids", value)

    @property
    @pulumi.getter(name="tailscaleTags")
    def tailscale_tags(self) -> pulumi.Input[Sequence[pulumi.Input[str]]]:
        """
        The tags to apply to the tailnet device andauth key. This tag should be added to your oauth key and ACL.
        """
        return pulumi.get(self, "tailscale_tags")

    @tailscale_tags.setter
    def tailscale_tags(self, value: pulumi.Input[Sequence[pulumi.Input[str]]]):
        pulumi.set(self, "tailscale_tags", value)

    @property
    @pulumi.getter(name="vpcId")
    def vpc_id(self) -> pulumi.Input[str]:
        """
        The VPC the Bastion should be created in.
        """
        return pulumi.get(self, "vpc_id")

    @vpc_id.setter
    def vpc_id(self, value: pulumi.Input[str]):
        pulumi.set(self, "vpc_id", value)

    @property
    @pulumi.getter(name="enableAppConnector")
    def enable_app_connector(self) -> Optional[pulumi.Input[bool]]:
        """
        Whether the bastion advertises itself as an app connector.
        """
        return pulumi.get(self, "enable_app_connector")

    @enable_app_connector.setter
    def enable_app_connector(self, value: Optional[pulumi.Input[bool]]):
        pulumi.set(self, "enable_app_connector", value)

    @property
    @pulumi.getter(name="enableExitNode")
    def enable_exit_node(self) -> Optional[pulumi.Input[bool]]:
        """
        Whether the subnet router can advertise itself as an exit node.
        """
        return pulumi.get(self, "enable_exit_node")

    @enable_exit_node.setter
    def enable_exit_node(self, value: Optional[pulumi.Input[bool]]):
        pulumi.set(self, "enable_exit_node", value)

    @property
    @pulumi.getter(name="enableSSH")
    def enable_ssh(self) -> Optional[pulumi.Input[bool]]:
        """
        Whether to enable SSH access to the bastion.
        """
        return pulumi.get(self, "enable_ssh")

    @enable_ssh.setter
    def enable_ssh(self, value: Optional[pulumi.Input[bool]]):
        pulumi.set(self, "enable_ssh", value)

    @property
    @pulumi.getter
    def hostname(self) -> Optional[pulumi.Input[str]]:
        """
        The hostname of the bastion.
        """
        return pulumi.get(self, "hostname")

    @hostname.setter
    def hostname(self, value: Optional[pulumi.Input[str]]):
        pulumi.set(self, "hostname", value)

    @property
    @pulumi.getter(name="instanceType")
    def instance_type(self) -> Optional[pulumi.Input[str]]:
        """
        The EC2 instance type to use for the bastion.
        """
        return pulumi.get(self, "instance_type")

    @instance_type.setter
    def instance_type(self, value: Optional[pulumi.Input[str]]):
        pulumi.set(self, "instance_type", value)

    @property
    @pulumi.getter(name="oauthClientSecret")
    def oauth_client_secret(self) -> Optional[pulumi.Input[str]]:
        """
        An OAuth Client Secret to use for authenticating Tailscale clients.
        """
        return pulumi.get(self, "oauth_client_secret")

    @oauth_client_secret.setter
    def oauth_client_secret(self, value: Optional[pulumi.Input[str]]):
        pulumi.set(self, "oauth_client_secret", value)

    @property
    @pulumi.getter
    def public(self) -> Optional[pulumi.Input[bool]]:
        """
        Whether the bastion is going in public subnets.
        """
        return pulumi.get(self, "public")

    @public.setter
    def public(self, value: Optional[pulumi.Input[bool]]):
        pulumi.set(self, "public", value)

    @property
    @pulumi.getter
    def routes(self) -> Optional[pulumi.Input[Sequence[pulumi.Input[str]]]]:
        """
        The routes you'd like to advertise via tailscale.
        """
        return pulumi.get(self, "routes")

    @routes.setter
    def routes(self, value: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]]):
        pulumi.set(self, "routes", value)


class Bastion(pulumi.ComponentResource):
    @overload
    def __init__(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 enable_app_connector: Optional[pulumi.Input[bool]] = None,
                 enable_exit_node: Optional[pulumi.Input[bool]] = None,
                 enable_ssh: Optional[pulumi.Input[bool]] = None,
                 high_availability: Optional[pulumi.Input[bool]] = None,
                 hostname: Optional[pulumi.Input[str]] = None,
                 instance_type: Optional[pulumi.Input[str]] = None,
                 oauth_client_secret: Optional[pulumi.Input[str]] = None,
                 public: Optional[pulumi.Input[bool]] = None,
                 region: Optional[pulumi.Input[str]] = None,
                 routes: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]] = None,
                 subnet_ids: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]] = None,
                 tailscale_tags: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]] = None,
                 vpc_id: Optional[pulumi.Input[str]] = None,
                 __props__=None):
        """
        Create a Bastion resource with the given unique name, props, and options.
        :param str resource_name: The name of the resource.
        :param pulumi.ResourceOptions opts: Options for the resource.
        :param pulumi.Input[bool] enable_app_connector: Whether the bastion advertises itself as an app connector.
        :param pulumi.Input[bool] enable_exit_node: Whether the subnet router can advertise itself as an exit node.
        :param pulumi.Input[bool] enable_ssh: Whether to enable SSH access to the bastion.
        :param pulumi.Input[bool] high_availability: Whether the bastion should be highly available.
        :param pulumi.Input[str] hostname: The hostname of the bastion.
        :param pulumi.Input[str] instance_type: The EC2 instance type to use for the bastion.
        :param pulumi.Input[str] oauth_client_secret: An OAuth Client Secret to use for authenticating Tailscale clients.
        :param pulumi.Input[bool] public: Whether the bastion is going in public subnets.
        :param pulumi.Input[str] region: The AWS region you're using.
        :param pulumi.Input[Sequence[pulumi.Input[str]]] routes: The routes you'd like to advertise via tailscale.
        :param pulumi.Input[Sequence[pulumi.Input[str]]] subnet_ids: The subnet Ids to launch instances in.
        :param pulumi.Input[Sequence[pulumi.Input[str]]] tailscale_tags: The tags to apply to the tailnet device andauth key. This tag should be added to your oauth key and ACL.
        :param pulumi.Input[str] vpc_id: The VPC the Bastion should be created in.
        """
        ...
    @overload
    def __init__(__self__,
                 resource_name: str,
                 args: BastionArgs,
                 opts: Optional[pulumi.ResourceOptions] = None):
        """
        Create a Bastion resource with the given unique name, props, and options.
        :param str resource_name: The name of the resource.
        :param BastionArgs args: The arguments to use to populate this resource's properties.
        :param pulumi.ResourceOptions opts: Options for the resource.
        """
        ...
    def __init__(__self__, resource_name: str, *args, **kwargs):
        resource_args, opts = _utilities.get_resource_args_opts(BastionArgs, pulumi.ResourceOptions, *args, **kwargs)
        if resource_args is not None:
            __self__._internal_init(resource_name, opts, **resource_args.__dict__)
        else:
            __self__._internal_init(resource_name, *args, **kwargs)

    def _internal_init(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 enable_app_connector: Optional[pulumi.Input[bool]] = None,
                 enable_exit_node: Optional[pulumi.Input[bool]] = None,
                 enable_ssh: Optional[pulumi.Input[bool]] = None,
                 high_availability: Optional[pulumi.Input[bool]] = None,
                 hostname: Optional[pulumi.Input[str]] = None,
                 instance_type: Optional[pulumi.Input[str]] = None,
                 oauth_client_secret: Optional[pulumi.Input[str]] = None,
                 public: Optional[pulumi.Input[bool]] = None,
                 region: Optional[pulumi.Input[str]] = None,
                 routes: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]] = None,
                 subnet_ids: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]] = None,
                 tailscale_tags: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]] = None,
                 vpc_id: Optional[pulumi.Input[str]] = None,
                 __props__=None):
        opts = pulumi.ResourceOptions.merge(_utilities.get_resource_opts_defaults(), opts)
        if not isinstance(opts, pulumi.ResourceOptions):
            raise TypeError('Expected resource options to be a ResourceOptions instance')
        if opts.id is not None:
            raise ValueError('ComponentResource classes do not support opts.id')
        else:
            if __props__ is not None:
                raise TypeError('__props__ is only valid when passed in combination with a valid opts.id to get an existing resource')
            __props__ = BastionArgs.__new__(BastionArgs)

            if enable_app_connector is None:
                enable_app_connector = False
            __props__.__dict__["enable_app_connector"] = enable_app_connector
            if enable_exit_node is None:
                enable_exit_node = False
            __props__.__dict__["enable_exit_node"] = enable_exit_node
            if enable_ssh is None:
                enable_ssh = True
            __props__.__dict__["enable_ssh"] = enable_ssh
            if high_availability is None:
                high_availability = False
            if high_availability is None and not opts.urn:
                raise TypeError("Missing required property 'high_availability'")
            __props__.__dict__["high_availability"] = high_availability
            __props__.__dict__["hostname"] = hostname
            __props__.__dict__["instance_type"] = instance_type
            __props__.__dict__["oauth_client_secret"] = oauth_client_secret
            if public is None:
                public = False
            __props__.__dict__["public"] = public
            if region is None and not opts.urn:
                raise TypeError("Missing required property 'region'")
            __props__.__dict__["region"] = region
            __props__.__dict__["routes"] = routes
            if subnet_ids is None and not opts.urn:
                raise TypeError("Missing required property 'subnet_ids'")
            __props__.__dict__["subnet_ids"] = subnet_ids
            if tailscale_tags is None and not opts.urn:
                raise TypeError("Missing required property 'tailscale_tags'")
            __props__.__dict__["tailscale_tags"] = tailscale_tags
            if vpc_id is None and not opts.urn:
                raise TypeError("Missing required property 'vpc_id'")
            __props__.__dict__["vpc_id"] = vpc_id
            __props__.__dict__["asg_name"] = None
            __props__.__dict__["private_key"] = None
        super(Bastion, __self__).__init__(
            'tailscale-bastion:aws:Bastion',
            resource_name,
            __props__,
            opts,
            remote=True)

    @property
    @pulumi.getter(name="asgName")
    def asg_name(self) -> pulumi.Output[str]:
        """
        The name of the ASG that managed the bastion instances
        """
        return pulumi.get(self, "asg_name")

    @property
    @pulumi.getter(name="privateKey")
    def private_key(self) -> pulumi.Output[str]:
        """
        The SSH private key to access your bastion
        """
        return pulumi.get(self, "private_key")

