package kubernetes

import (
	"fmt"
	"strings"

	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	rbacv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/rbac/v1"
	"github.com/pulumi/pulumi-tailscale/sdk/go/tailscale"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The set of arguments for creating a Bastion component resource.
type BastionArgs struct {
	CreateNamespace bool                    `pulumi:"createNamespace"`
	Namespace       *corev1.Namespace       `pulumi:"namespace"`
	Routes          pulumi.StringArrayInput `pulumi:"routes"`
	TailscaleTags   pulumi.StringArrayInput `pulumi:"tailscaleTags"`
}

// The Bastion component resource.
type Bastion struct {
	pulumi.ResourceState

	DeploymentName pulumi.StringOutput `pulumi:"deploymentName"`
}

// NewBastion creates a new Bastion component resource.
func NewBastion(ctx *pulumi.Context,
	name string, args *BastionArgs, opts ...pulumi.ResourceOption) (*Bastion, error) {
	if args == nil {
		args = &BastionArgs{}
	}

	component := &Bastion{}

	err := ctx.RegisterComponentResource("tailscale-bastion:kubernetes:Bastion", name, component, opts...)
	if err != nil {
		return nil, err
	}

	// create a tailnet key to auth devices
	tailnetKey, err := tailscale.NewTailnetKey(ctx, name, &tailscale.TailnetKeyArgs{
		Ephemeral:     pulumi.Bool(true),
		Preauthorized: pulumi.Bool(true),
		Reusable:      pulumi.Bool(true),
		Tags:          args.TailscaleTags,
	}, pulumi.Parent(component))
	if err != nil {
		return nil, fmt.Errorf("error creating tailnet key: %v", err)
	}

	var namespace *corev1.Namespace

	if args.CreateNamespace {
		namespace, err = corev1.NewNamespace(ctx, name, &corev1.NamespaceArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name: pulumi.String(name),
			},
		}, pulumi.Parent(component))
		if err != nil {
			return nil, fmt.Errorf("error creating namespace: %v", err)
		}
	} else {
		namespace = args.Namespace
	}

	secret, err := corev1.NewSecret(ctx, name, &corev1.SecretArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Namespace: namespace.Metadata.Name(),
		},
		StringData: pulumi.StringMap{
			"TS_AUTH_KEY": tailnetKey.Key,
		},
	}, pulumi.Parent(namespace))
	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes secret: %v", err)
	}

	serviceAccount, err := corev1.NewServiceAccount(ctx, name, &corev1.ServiceAccountArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Namespace: namespace.Metadata.Name(),
		},
	}, pulumi.Parent(namespace))
	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes service account: %v", err)
	}

	role, err := rbacv1.NewRole(ctx, name, &rbacv1.RoleArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Namespace: namespace.Metadata.Name(),
		},
		Rules: rbacv1.PolicyRuleArray{
			&rbacv1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{
					pulumi.String(""), // Core API group
				},
				Resources: pulumi.StringArray{
					pulumi.String("secrets"),
				},
				Verbs: pulumi.StringArray{
					pulumi.String("create"),
				},
			},
			&rbacv1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{
					pulumi.String(""),
				},
				Resources: pulumi.StringArray{
					pulumi.String("secrets"),
				},
				Verbs: pulumi.StringArray{
					pulumi.String("get"),
					pulumi.String("update"),
				},
			},
		},
	}, pulumi.Parent(namespace))
	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes role: %v", err)
	}

	_, err = rbacv1.NewRoleBinding(ctx, name, &rbacv1.RoleBindingArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Namespace: namespace.Metadata.Name(),
		},
		Subjects: rbacv1.SubjectArray{
			&rbacv1.SubjectArgs{
				Kind: pulumi.String("ServiceAccount"),
				Name: serviceAccount.Metadata.Name().Elem(),
			},
		},
		RoleRef: &rbacv1.RoleRefArgs{
			Kind:     pulumi.String("Role"),
			Name:     role.Metadata.Name().Elem(),
			ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
		},
	}, pulumi.Parent(role))
	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes role binding: %v", err)
	}

	// build a csv from the input routes
	routes := args.Routes.ToStringArrayOutput().ApplyT(
		func(routes interface{}) string {
			return strings.Join(routes.([]string), ",")
		},
	).(pulumi.StringOutput)

	deployment, err := appsv1.NewDeployment(ctx, name, &appsv1.DeploymentArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Namespace: namespace.Metadata.Name(),
		},
		Spec: &appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"name":        pulumi.String(name),
					"application": pulumi.String("tailscale"),
				},
			},
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: pulumi.StringMap{
						"name":        pulumi.String(name),
						"application": pulumi.String("tailscale"),
					},
				},
				Spec: &corev1.PodSpecArgs{
					ServiceAccountName: serviceAccount.Metadata.Name(),
					Containers: corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:            pulumi.String("tailscale"),
							Image:           pulumi.String("ghcr.io/tailscale/tailscale:latest"),
							ImagePullPolicy: pulumi.String("Always"),
							SecurityContext: &corev1.SecurityContextArgs{
								RunAsUser:  pulumi.Int(1000),
								RunAsGroup: pulumi.Int(1000),
							},
							Env: corev1.EnvVarArray{
								corev1.EnvVarArgs{
									Name: pulumi.String("TS_AUTH_KEY"),
									ValueFrom: &corev1.EnvVarSourceArgs{
										SecretKeyRef: &corev1.SecretKeySelectorArgs{
											Name: secret.Metadata.Name(),
											Key:  pulumi.String("TS_AUTH_KEY"),
										},
									},
								},
								corev1.EnvVarArgs{
									Name:  pulumi.String("TS_USERSPACE"),
									Value: pulumi.String("true"),
								},
								corev1.EnvVarArgs{
									Name:  pulumi.String("TS_ROUTES"),
									Value: routes,
								},
							},
						},
					},
				},
			},
		},
	}, pulumi.Parent(namespace))
	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes deployment: %v", err)
	}

	component.DeploymentName = deployment.Metadata.Name().Elem()

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"deploymentName": deployment.Metadata.Name(),
	}); err != nil {
		return nil, err
	}

	return component, nil
}
