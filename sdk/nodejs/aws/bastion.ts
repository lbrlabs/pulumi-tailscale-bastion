// *** WARNING: this file was generated by Pulumi SDK Generator. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as utilities from "../utilities";

export class Bastion extends pulumi.ComponentResource {
    /** @internal */
    public static readonly __pulumiType = 'tailscale-bastion:aws:Bastion';

    /**
     * Returns true if the given object is an instance of Bastion.  This is designed to work even
     * when multiple copies of the Pulumi SDK have been loaded into the same process.
     */
    public static isInstance(obj: any): obj is Bastion {
        if (obj === undefined || obj === null) {
            return false;
        }
        return obj['__pulumiType'] === Bastion.__pulumiType;
    }

    /**
     * The name of the ASG that managed the bastion instances
     */
    public /*out*/ readonly asgName!: pulumi.Output<string>;
    /**
     * The SSH private key to access your bastion
     */
    public /*out*/ readonly privateKey!: pulumi.Output<string>;

    /**
     * Create a Bastion resource with the given unique name, arguments, and options.
     *
     * @param name The _unique_ name of the resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param opts A bag of options that control this resource's behavior.
     */
    constructor(name: string, args: BastionArgs, opts?: pulumi.ComponentResourceOptions) {
        let resourceInputs: pulumi.Inputs = {};
        opts = opts || {};
        if (!opts.id) {
            if ((!args || args.highAvailability === undefined) && !opts.urn) {
                throw new Error("Missing required property 'highAvailability'");
            }
            if ((!args || args.region === undefined) && !opts.urn) {
                throw new Error("Missing required property 'region'");
            }
            if ((!args || args.subnetIds === undefined) && !opts.urn) {
                throw new Error("Missing required property 'subnetIds'");
            }
            if ((!args || args.tailscaleTags === undefined) && !opts.urn) {
                throw new Error("Missing required property 'tailscaleTags'");
            }
            if ((!args || args.vpcId === undefined) && !opts.urn) {
                throw new Error("Missing required property 'vpcId'");
            }
            resourceInputs["enableAppConnector"] = (args ? args.enableAppConnector : undefined) ?? false;
            resourceInputs["enableExitNode"] = (args ? args.enableExitNode : undefined) ?? false;
            resourceInputs["enableSSH"] = (args ? args.enableSSH : undefined) ?? true;
            resourceInputs["highAvailability"] = (args ? args.highAvailability : undefined) ?? false;
            resourceInputs["hostname"] = args ? args.hostname : undefined;
            resourceInputs["instanceType"] = args ? args.instanceType : undefined;
            resourceInputs["oauthClientSecret"] = args ? args.oauthClientSecret : undefined;
            resourceInputs["public"] = (args ? args.public : undefined) ?? false;
            resourceInputs["region"] = args ? args.region : undefined;
            resourceInputs["routes"] = args ? args.routes : undefined;
            resourceInputs["subnetIds"] = args ? args.subnetIds : undefined;
            resourceInputs["tailscaleTags"] = args ? args.tailscaleTags : undefined;
            resourceInputs["vpcId"] = args ? args.vpcId : undefined;
            resourceInputs["asgName"] = undefined /*out*/;
            resourceInputs["privateKey"] = undefined /*out*/;
        } else {
            resourceInputs["asgName"] = undefined /*out*/;
            resourceInputs["privateKey"] = undefined /*out*/;
        }
        opts = pulumi.mergeOptions(utilities.resourceOptsDefaults(), opts);
        super(Bastion.__pulumiType, name, resourceInputs, opts, true /*remote*/);
    }
}

/**
 * The set of arguments for constructing a Bastion resource.
 */
export interface BastionArgs {
    /**
     * Whether the bastion advertises itself as an app connector.
     */
    enableAppConnector?: pulumi.Input<boolean>;
    /**
     * Whether the subnet router can advertise itself as an exit node.
     */
    enableExitNode?: pulumi.Input<boolean>;
    /**
     * Whether to enable SSH access to the bastion.
     */
    enableSSH?: pulumi.Input<boolean>;
    /**
     * Whether the bastion should be highly available.
     */
    highAvailability: pulumi.Input<boolean>;
    /**
     * The hostname of the bastion.
     */
    hostname?: pulumi.Input<string>;
    /**
     * The EC2 instance type to use for the bastion.
     */
    instanceType?: pulumi.Input<string>;
    /**
     * An OAuth Client Secret to use for authenticating Tailscale clients.
     */
    oauthClientSecret?: pulumi.Input<string>;
    /**
     * Whether the bastion is going in public subnets.
     */
    public?: pulumi.Input<boolean>;
    /**
     * The AWS region you're using.
     */
    region: pulumi.Input<string>;
    /**
     * The routes you'd like to advertise via tailscale.
     */
    routes?: pulumi.Input<pulumi.Input<string>[]>;
    /**
     * The subnet Ids to launch instances in.
     */
    subnetIds: pulumi.Input<pulumi.Input<string>[]>;
    /**
     * The tags to apply to the tailnet device andauth key. This tag should be added to your oauth key and ACL.
     */
    tailscaleTags: pulumi.Input<pulumi.Input<string>[]>;
    /**
     * The VPC the Bastion should be created in.
     */
    vpcId: pulumi.Input<string>;
}
