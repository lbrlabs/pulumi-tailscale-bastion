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
                 location: pulumi.Input[str],
                 resource_group_name: pulumi.Input[str],
                 route: pulumi.Input[str],
                 subnet_id: pulumi.Input[str],
                 instance_sku: Optional[pulumi.Input[str]] = None):
        """
        The set of arguments for constructing a Bastion resource.
        :param pulumi.Input[str] location: The Azure region you're using.
        :param pulumi.Input[str] resource_group_name: The Azure resource group to create the bastion in.
        :param pulumi.Input[str] route: The route you'd like to advertise via tailscale.
        :param pulumi.Input[str] subnet_id: The subnet Ids to launch instances in.
        :param pulumi.Input[str] instance_sku: The Azure instance SKU to use for the bastion.
        """
        pulumi.set(__self__, "location", location)
        pulumi.set(__self__, "resource_group_name", resource_group_name)
        pulumi.set(__self__, "route", route)
        pulumi.set(__self__, "subnet_id", subnet_id)
        if instance_sku is not None:
            pulumi.set(__self__, "instance_sku", instance_sku)

    @property
    @pulumi.getter
    def location(self) -> pulumi.Input[str]:
        """
        The Azure region you're using.
        """
        return pulumi.get(self, "location")

    @location.setter
    def location(self, value: pulumi.Input[str]):
        pulumi.set(self, "location", value)

    @property
    @pulumi.getter(name="resourceGroupName")
    def resource_group_name(self) -> pulumi.Input[str]:
        """
        The Azure resource group to create the bastion in.
        """
        return pulumi.get(self, "resource_group_name")

    @resource_group_name.setter
    def resource_group_name(self, value: pulumi.Input[str]):
        pulumi.set(self, "resource_group_name", value)

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
    @pulumi.getter(name="subnetId")
    def subnet_id(self) -> pulumi.Input[str]:
        """
        The subnet Ids to launch instances in.
        """
        return pulumi.get(self, "subnet_id")

    @subnet_id.setter
    def subnet_id(self, value: pulumi.Input[str]):
        pulumi.set(self, "subnet_id", value)

    @property
    @pulumi.getter(name="instanceSku")
    def instance_sku(self) -> Optional[pulumi.Input[str]]:
        """
        The Azure instance SKU to use for the bastion.
        """
        return pulumi.get(self, "instance_sku")

    @instance_sku.setter
    def instance_sku(self, value: Optional[pulumi.Input[str]]):
        pulumi.set(self, "instance_sku", value)


class Bastion(pulumi.ComponentResource):
    @overload
    def __init__(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 instance_sku: Optional[pulumi.Input[str]] = None,
                 location: Optional[pulumi.Input[str]] = None,
                 resource_group_name: Optional[pulumi.Input[str]] = None,
                 route: Optional[pulumi.Input[str]] = None,
                 subnet_id: Optional[pulumi.Input[str]] = None,
                 __props__=None):
        """
        Create a Bastion resource with the given unique name, props, and options.
        :param str resource_name: The name of the resource.
        :param pulumi.ResourceOptions opts: Options for the resource.
        :param pulumi.Input[str] instance_sku: The Azure instance SKU to use for the bastion.
        :param pulumi.Input[str] location: The Azure region you're using.
        :param pulumi.Input[str] resource_group_name: The Azure resource group to create the bastion in.
        :param pulumi.Input[str] route: The route you'd like to advertise via tailscale.
        :param pulumi.Input[str] subnet_id: The subnet Ids to launch instances in.
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
                 instance_sku: Optional[pulumi.Input[str]] = None,
                 location: Optional[pulumi.Input[str]] = None,
                 resource_group_name: Optional[pulumi.Input[str]] = None,
                 route: Optional[pulumi.Input[str]] = None,
                 subnet_id: Optional[pulumi.Input[str]] = None,
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

            __props__.__dict__["instance_sku"] = instance_sku
            if location is None and not opts.urn:
                raise TypeError("Missing required property 'location'")
            __props__.__dict__["location"] = location
            if resource_group_name is None and not opts.urn:
                raise TypeError("Missing required property 'resource_group_name'")
            __props__.__dict__["resource_group_name"] = resource_group_name
            if route is None and not opts.urn:
                raise TypeError("Missing required property 'route'")
            __props__.__dict__["route"] = route
            if subnet_id is None and not opts.urn:
                raise TypeError("Missing required property 'subnet_id'")
            __props__.__dict__["subnet_id"] = subnet_id
            __props__.__dict__["private_key"] = None
            __props__.__dict__["scale_set_name"] = None
        super(Bastion, __self__).__init__(
            'tailscale-bastion:azure:Bastion',
            resource_name,
            __props__,
            opts,
            remote=True)

    @property
    @pulumi.getter(name="privateKey")
    def private_key(self) -> pulumi.Output[str]:
        """
        The SSH private key to access your bastion
        """
        return pulumi.get(self, "private_key")

    @property
    @pulumi.getter(name="scaleSetName")
    def scale_set_name(self) -> pulumi.Output[str]:
        """
        The name of the Scaleset that managed the bastion instances
        """
        return pulumi.get(self, "scale_set_name")

