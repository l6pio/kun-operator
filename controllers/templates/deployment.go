package templates

import (
	"context"
	_ "embed"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:embed server/deployment.yaml
var ServerDeployTemplate string

//go:embed ui/deployment.yaml
var UIDeployTemplate string

func GetDeployment(c client.Client, ctx context.Context, name string, namespace string) (*apps.Deployment, error) {
	var deployments apps.DeploymentList
	if err := c.List(ctx, &deployments, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	if len(deployments.Items) > 0 {
		for _, deploy := range deployments.Items {
			if deploy.Name == name {
				return &deploy, nil
			}
		}
	}
	return nil, nil
}

func CreateDeployment(
	c client.Client, ctx context.Context, name string, namespace string,
	template string, replicas *int32, image string,
	resources *core.ResourceRequirements, env []core.EnvVar,
) error {
	var deploy, err = GetDeployment(c, ctx, name, namespace)
	if err != nil {
		return err
	} else if deploy == nil {
		decoder := serializer.NewCodecFactory(scheme.Scheme).UniversalDecoder()

		deploy = new(apps.Deployment)
		if err := runtime.DecodeInto(decoder, []byte(template), deploy); err != nil {
			return err
		}
		deploy.Name = name
		deploy.Namespace = namespace
		deploy.Spec.Replicas = replicas

		container := &deploy.Spec.Template.Spec.Containers[0]
		container.Image = image
		container.Resources = *resources

		if env != nil {
			container.Env = env
		}

		if err := c.Create(ctx, deploy); err != nil {
			return err
		}
	}
	return nil
}

func DeleteDeployment(c client.Client, ctx context.Context, name string, namespace string) error {
	deploy, err := GetDeployment(c, ctx, name, namespace)
	if err != nil {
		return err
	} else if deploy != nil {
		if err := c.Delete(ctx, deploy); err != nil {
			return err
		}
	}
	return nil
}
