# coding=utf-8
# *** WARNING: this file was generated by Pulumi SDK Generator. ***
# *** Do not edit by hand unless you're certain you know what you are doing! ***

import copy
import warnings
import pulumi
import pulumi.runtime
from typing import Any, Mapping, Optional, Sequence, Union, overload
from .. import _utilities

__all__ = ['BastionArgs', 'Bastion']

@pulumi.input_type
class BastionArgs:
    def __init__(__self__, *,
                 route: pulumi.Input[str],
                 subnet_network_id: pulumi.Input[str],
                 machine_type: Optional[pulumi.Input[str]] = None):
        """
        The set of arguments for constructing a Bastion resource.
        :param pulumi.Input[str] route: The route you'd like to advertise via tailscale.
        :param pulumi.Input[str] subnet_network_id: The subnetwork to create the bastion in.
        :param pulumi.Input[str] machine_type: The GCP machine type to launch. Defaults to f1-micro.
        """
        pulumi.set(__self__, "route", route)
        pulumi.set(__self__, "subnet_network_id", subnet_network_id)
        if machine_type is not None:
            pulumi.set(__self__, "machine_type", machine_type)

    @property
    @pulumi.getter
    def route(self) -> pulumi.Input[str]:
        """
        The route you'd like to advertise via tailscale.
        """
        return pulumi.get(self, "route")

    @route.setter
    def route(self, value: pulumi.Input[str]):
        pulumi.set(self, "route", value)

    @property
    @pulumi.getter(name="subnetNetworkId")
    def subnet_network_id(self) -> pulumi.Input[str]:
        """
        The subnetwork to create the bastion in.
        """
        return pulumi.get(self, "subnet_network_id")

    @subnet_network_id.setter
    def subnet_network_id(self, value: pulumi.Input[str]):
        pulumi.set(self, "subnet_network_id", value)

    @property
    @pulumi.getter(name="machineType")
    def machine_type(self) -> Optional[pulumi.Input[str]]:
        """
        The GCP machine type to launch. Defaults to f1-micro.
        """
        return pulumi.get(self, "machine_type")

    @machine_type.setter
    def machine_type(self, value: Optional[pulumi.Input[str]]):
        pulumi.set(self, "machine_type", value)


class Bastion(pulumi.ComponentResource):
    @overload
    def __init__(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 machine_type: Optional[pulumi.Input[str]] = None,
                 route: Optional[pulumi.Input[str]] = None,
                 subnet_network_id: Optional[pulumi.Input[str]] = None,
                 __props__=None):
        """
        Create a Bastion resource with the given unique name, props, and options.
        :param str resource_name: The name of the resource.
        :param pulumi.ResourceOptions opts: Options for the resource.
        :param pulumi.Input[str] machine_type: The GCP machine type to launch. Defaults to f1-micro.
        :param pulumi.Input[str] route: The route you'd like to advertise via tailscale.
        :param pulumi.Input[str] subnet_network_id: The subnetwork to create the bastion in.
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
                 machine_type: Optional[pulumi.Input[str]] = None,
                 route: Optional[pulumi.Input[str]] = None,
                 subnet_network_id: Optional[pulumi.Input[str]] = None,
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

            __props__.__dict__["machine_type"] = machine_type
            if route is None and not opts.urn:
                raise TypeError("Missing required property 'route'")
            __props__.__dict__["route"] = route
            if subnet_network_id is None and not opts.urn:
                raise TypeError("Missing required property 'subnet_network_id'")
            __props__.__dict__["subnet_network_id"] = subnet_network_id
            __props__.__dict__["autoscaler_name"] = None
            __props__.__dict__["group_manager_name"] = None
            __props__.__dict__["private_key"] = None
            __props__.__dict__["target_pool_name"] = None
        super(Bastion, __self__).__init__(
            'tailscale-bastion:gcp:Bastion',
            resource_name,
            __props__,
            opts,
            remote=True)

    @property
    @pulumi.getter(name="autoscalerName")
    def autoscaler_name(self) -> pulumi.Output[str]:
        """
        The name of the autoscaler that manages the instances.
        """
        return pulumi.get(self, "autoscaler_name")

    @property
    @pulumi.getter(name="groupManagerName")
    def group_manager_name(self) -> pulumi.Output[str]:
        """
        The name of the group manager that manages the instances.
        """
        return pulumi.get(self, "group_manager_name")

    @property
    @pulumi.getter(name="privateKey")
    def private_key(self) -> pulumi.Output[str]:
        """
        The SSH private key to access your bastion.
        """
        return pulumi.get(self, "private_key")

    @property
    @pulumi.getter(name="targetPoolName")
    def target_pool_name(self) -> pulumi.Output[str]:
        """
        The name of the target that manages the instances.
        """
        return pulumi.get(self, "target_pool_name")

