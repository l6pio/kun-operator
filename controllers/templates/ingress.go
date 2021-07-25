package templates

import (
	"context"
	_ "embed"
	networking "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:embed server/ingress.yaml
var ServerIngressTemplate string

//go:embed ui/ingress.yaml
var UIIngressTemplate string

func GetIngress(c client.Client, ctx context.Context, name string, namespace string) (*networking.Ingress, error) {
	var ingresses networking.IngressList
	if err := c.List(ctx, &ingresses, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	if len(ingresses.Items) > 0 {
		for _, ingress := range ingresses.Items {
			if ingress.Name == name {
				return &ingress, nil
			}
		}
	}
	return nil, nil
}

func CreateIngress(
	c client.Client, ctx context.Context, name string, namespace string,
	template string, host string, path string, port int32,
) error {
	var ingress, err = GetIngress(c, ctx, name, namespace)
	if err != nil {
		return err
	} else if ingress == nil {
		decoder := serializer.NewCodecFactory(scheme.Scheme).UniversalDecoder()

		ingress = new(networking.Ingress)
		if err := runtime.DecodeInto(decoder, []byte(template), ingress); err != nil {
			return err
		}
		ingress.Name = name
		ingress.Namespace = namespace

		pathTypePrefix := networking.PathTypePrefix
		ingress.Spec.Rules = []networking.IngressRule{
			{
				Host: host,
				IngressRuleValue: networking.IngressRuleValue{
					HTTP: &networking.HTTPIngressRuleValue{
						Paths: []networking.HTTPIngressPath{
							{
								Path:     path,
								PathType: &pathTypePrefix,
								Backend: networking.IngressBackend{
									Service: &networking.IngressServiceBackend{
										Name: name,
										Port: networking.ServiceBackendPort{
											Number: port,
										},
									},
								},
							},
						},
					},
				},
			},
		}
		if err := c.Create(ctx, ingress); err != nil {
			return err
		}
	}
	return nil
}

func DeleteIngress(c client.Client, ctx context.Context, name string, namespace string) error {
	ingress, err := GetIngress(c, ctx, name, namespace)
	if err != nil {
		return err
	} else if ingress != nil {
		if err := c.Delete(ctx, ingress); err != nil {
			return err
		}
	}
	return nil
}
