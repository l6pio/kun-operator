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

//go:embed server/cluster-role.yaml
var ServerClusterRoleTemplate string

func GetClusterRole(c client.Client, ctx context.Context, name string) (*v1.ClusterRole, error) {
	var roles v1.ClusterRoleList
	if err := c.List(ctx, &roles); err != nil {
		return nil, err
	}

	if len(roles.Items) > 0 {
		for _, role := range roles.Items {
			if role.Name == name {
				return &role, nil
			}
		}
	}
	return nil, nil
}

func CreateClusterRole(c client.Client, ctx context.Context, name string, template string) error {
	var role, err = GetClusterRole(c, ctx, name)
	if err != nil {
		return err
	} else if role == nil {
		decoder := serializer.NewCodecFactory(scheme.Scheme).UniversalDecoder()

		role = new(v1.ClusterRole)
		if err := runtime.DecodeInto(decoder, []byte(template), role); err != nil {
			return err
		}
		role.Name = name

		if err := c.Create(ctx, role); err != nil {
			return err
		}
	}
	return nil
}

func DeleteClusterRole(c client.Client, ctx context.Context, name string) error {
	role, err := GetClusterRole(c, ctx, name)
	if err != nil {
		return err
	} else if role != nil {
		if err := c.Delete(ctx, role); err != nil {
			return err
		}
	}
	return nil
}
