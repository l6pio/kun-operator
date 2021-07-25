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

//go:embed server/service-account.yaml
var ServerServiceAccountTemplate string

//go:embed ui/service-account.yaml
var UIServiceAccountTemplate string

func GetServiceAccount(c client.Client, ctx context.Context, name string, namespace string) (*core.ServiceAccount, error) {
	var serviceAccounts core.ServiceAccountList
	if err := c.List(ctx, &serviceAccounts, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	if len(serviceAccounts.Items) > 0 {
		for _, sa := range serviceAccounts.Items {
			if sa.Name == name {
				return &sa, nil
			}
		}
	}
	return nil, nil
}

func CreateServiceAccount(c client.Client, ctx context.Context, name string, namespace string, template string) error {
	var serviceAccount, err = GetServiceAccount(c, ctx, name, namespace)
	if err != nil {
		return err
	} else if serviceAccount == nil {
		decoder := serializer.NewCodecFactory(scheme.Scheme).UniversalDecoder()

		serviceAccount = new(core.ServiceAccount)
		if err := runtime.DecodeInto(decoder, []byte(template), serviceAccount); err != nil {
			return err
		}
		serviceAccount.Name = name
		serviceAccount.Namespace = namespace

		if err := c.Create(ctx, serviceAccount); err != nil {
			return err
		}
	}
	return nil
}

func DeleteServiceAccount(c client.Client, ctx context.Context, name string, namespace string) error {
	serviceAccount, err := GetServiceAccount(c, ctx, name, namespace)
	if err != nil {
		return err
	} else if serviceAccount != nil {
		if err := c.Delete(ctx, serviceAccount); err != nil {
			return err
		}
	}
	return nil
}
