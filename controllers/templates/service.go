package templates

import (
	"context"
	_ "embed"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:embed server/service.yaml
var ServerServiceTemplate string

//go:embed ui/service.yaml
var UIServiceTemplate string

func GetService(c client.Client, ctx context.Context, name string, namespace string) (*core.Service, error) {
	var services core.ServiceList
	if err := c.List(ctx, &services, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	if len(services.Items) > 0 {
		for _, service := range services.Items {
			if service.Name == name {
				return &service, nil
			}
		}
	}
	return nil, nil
}

func CreateService(
	c client.Client, ctx context.Context, name string, namespace string,
	template string, port int32,
) error {
	var service, err = GetService(c, ctx, name, namespace)
	if err != nil {
		return err
	} else if service == nil {
		decoder := serializer.NewCodecFactory(scheme.Scheme).UniversalDecoder()

		service = new(core.Service)
		if err := runtime.DecodeInto(decoder, []byte(template), service); err != nil {
			return err
		}
		service.Name = name
		service.Namespace = namespace
		service.Spec.Ports[0].Port = port

		if err := c.Create(ctx, service); err != nil {
			return err
		}
	}
	return nil
}

func DeleteService(c client.Client, ctx context.Context, name string, namespace string) error {
	service, err := GetService(c, ctx, name, namespace)
	if err != nil {
		return err
	} else if service != nil {
		if err := c.Delete(ctx, service); err != nil {
			return err
		}
	}
	return nil
}
