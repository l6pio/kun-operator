package templates

import (
	"context"
	_ "embed"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:embed server/cluster-role-binding.yaml
var ServerClusterRoleBindingTemplate string

func GetClusterRoleBinding(c client.Client, ctx context.Context, name string) (*v1.ClusterRoleBinding, error) {
	var roleBindings v1.ClusterRoleBindingList
	if err := c.List(ctx, &roleBindings); err != nil {
		return nil, err
	}

	if len(roleBindings.Items) > 0 {
		for _, roleBinding := range roleBindings.Items {
			if roleBinding.Name == name {
				return &roleBinding, nil
			}
		}
	}
	return nil, nil
}

func CreateClusterRoleBinding(c client.Client, ctx context.Context, name string, namespace string, template string) error {
	var roleBinding, err = GetClusterRoleBinding(c, ctx, name)
	if err != nil {
		return err
	} else if roleBinding == nil {
		decoder := serializer.NewCodecFactory(scheme.Scheme).UniversalDecoder()

		roleBinding = new(v1.ClusterRoleBinding)
		if err := runtime.DecodeInto(decoder, []byte(template), roleBinding); err != nil {
			return err
		}
		roleBinding.Name = name
		roleBinding.Subjects[0].Namespace = namespace

		if err := c.Create(ctx, roleBinding); err != nil {
			return err
		}
	}
	return nil
}

func DeleteClusterRoleBinding(c client.Client, ctx context.Context, name string) error {
	roleBinding, err := GetClusterRoleBinding(c, ctx, name)
	if err != nil {
		return err
	} else if roleBinding != nil {
		if err := c.Delete(ctx, roleBinding); err != nil {
			return err
		}
	}
	return nil
}
