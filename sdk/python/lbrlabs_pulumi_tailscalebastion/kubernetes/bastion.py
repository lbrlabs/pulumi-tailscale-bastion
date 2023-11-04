# coding=utf-8
# *** WARNING: this file was generated by Pulumi SDK Generator. ***
# *** Do not edit by hand unless you're certain you know what you are doing! ***

import copy
import warnings
import pulumi
import pulumi.runtime
from typing import Any, Mapping, Optional, Sequence, Union, overload
from .. import _utilities
import pulumi_kubernetes

__all__ = ['BastionArgs', 'Bastion']

@pulumi.input_type
class BastionArgs:
    def __init__(__self__, *,
                 create_namespace: bool,
                 routes: pulumi.Input[Sequence[pulumi.Input[str]]],
                 namespace: Optional[pulumi.Input['pulumi_kubernetes.core.v1.Namespace']] = None,
                 tailscale_tags: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]] = None):
        """
        The set of arguments for constructing a Bastion resource.
        :param bool create_namespace: Whether we should create a new namespace.
        :param pulumi.Input[Sequence[pulumi.Input[str]]] routes: The routes to advertise to tailscale. This is likely the Pod and Service CIDR.
        :param pulumi.Input['pulumi_kubernetes.core.v1.Namespace'] namespace: The bucket resource.
        :param pulumi.Input[Sequence[pulumi.Input[str]]] tailscale_tags: The tags to apply to the tailnet device andauth key. This tag should be added to your oauth key and ACL.
        """
        pulumi.set(__self__, "create_namespace", create_namespace)
        pulumi.set(__self__, "routes", routes)
        if namespace is not None:
            pulumi.set(__self__, "namespace", namespace)
        if tailscale_tags is not None:
            pulumi.set(__self__, "tailscale_tags", tailscale_tags)

    @property
    @pulumi.getter(name="createNamespace")
    def create_namespace(self) -> bool:
        """
        Whether we should create a new namespace.
        """
        return pulumi.get(self, "create_namespace")

    @create_namespace.setter
    def create_namespace(self, value: bool):
        pulumi.set(self, "create_namespace", value)

    @property
    @pulumi.getter
    def routes(self) -> pulumi.Input[Sequence[pulumi.Input[str]]]:
        """
        The routes to advertise to tailscale. This is likely the Pod and Service CIDR.
        """
        return pulumi.get(self, "routes")

    @routes.setter
    def routes(self, value: pulumi.Input[Sequence[pulumi.Input[str]]]):
        pulumi.set(self, "routes", value)

    @property
    @pulumi.getter
    def namespace(self) -> Optional[pulumi.Input['pulumi_kubernetes.core.v1.Namespace']]:
        """
        The bucket resource.
        """
        return pulumi.get(self, "namespace")

    @namespace.setter
    def namespace(self, value: Optional[pulumi.Input['pulumi_kubernetes.core.v1.Namespace']]):
        pulumi.set(self, "namespace", value)

    @property
    @pulumi.getter(name="tailscaleTags")
    def tailscale_tags(self) -> Optional[pulumi.Input[Sequence[pulumi.Input[str]]]]:
        """
        The tags to apply to the tailnet device andauth key. This tag should be added to your oauth key and ACL.
        """
        return pulumi.get(self, "tailscale_tags")

    @tailscale_tags.setter
    def tailscale_tags(self, value: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]]):
        pulumi.set(self, "tailscale_tags", value)


class Bastion(pulumi.ComponentResource):
    @overload
    def __init__(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 create_namespace: Optional[bool] = None,
                 namespace: Optional[pulumi.Input['pulumi_kubernetes.core.v1.Namespace']] = None,
                 routes: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]] = None,
                 tailscale_tags: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]] = None,
                 __props__=None):
        """
        Create a Bastion resource with the given unique name, props, and options.
        :param str resource_name: The name of the resource.
        :param pulumi.ResourceOptions opts: Options for the resource.
        :param bool create_namespace: Whether we should create a new namespace.
        :param pulumi.Input['pulumi_kubernetes.core.v1.Namespace'] namespace: The bucket resource.
        :param pulumi.Input[Sequence[pulumi.Input[str]]] routes: The routes to advertise to tailscale. This is likely the Pod and Service CIDR.
        :param pulumi.Input[Sequence[pulumi.Input[str]]] tailscale_tags: The tags to apply to the tailnet device andauth key. This tag should be added to your oauth key and ACL.
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
                 create_namespace: Optional[bool] = None,
                 namespace: Optional[pulumi.Input['pulumi_kubernetes.core.v1.Namespace']] = None,
                 routes: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]] = None,
                 tailscale_tags: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]] = None,
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

            if create_namespace is None and not opts.urn:
                raise TypeError("Missing required property 'create_namespace'")
            __props__.__dict__["create_namespace"] = create_namespace
            __props__.__dict__["namespace"] = namespace
            if routes is None and not opts.urn:
                raise TypeError("Missing required property 'routes'")
            __props__.__dict__["routes"] = routes
            __props__.__dict__["tailscale_tags"] = tailscale_tags
            __props__.__dict__["deployment_name"] = None
        super(Bastion, __self__).__init__(
            'tailscale-bastion:kubernetes:Bastion',
            resource_name,
            __props__,
            opts,
            remote=True)

    @property
    @pulumi.getter(name="deploymentName")
    def deployment_name(self) -> pulumi.Output[str]:
        """
        The name of the kubernetes deployment that contains the tailscale bastion
        """
        return pulumi.get(self, "deployment_name")

