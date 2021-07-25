/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/go-logr/logr"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/workqueue"
	kunapi "l6p.io/KunOperator/api/v1"
	"l6p.io/KunOperator/controllers/templates"
	"path"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const KunApi = "kun-api"
const KunUI = "kun-ui"

// KunInstallationReconciler reconciles a KunInstallation object
type KunInstallationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=l6p.io,resources=kuninstallations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=l6p.io,resources=kuninstallations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=l6p.io,resources=kuninstallations/finalizers,verbs=update

//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KunInstallation object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *KunInstallationReconciler) Reconcile(_ context.Context, _ ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

type EnqueueRequestForKunInstallation struct {
	Logger logr.Logger
	client.Client
}

func (e EnqueueRequestForKunInstallation) Create(createEvent event.CreateEvent, _ workqueue.RateLimitingInterface) {
	install := createEvent.Object.(*kunapi.KunInstallation)

	var deploys apps.DeploymentList
	if err := e.Client.List(context.TODO(), &deploys,
		&client.ListOptions{
			LabelSelector: labels.SelectorFromSet(labels.Set{"l6p-app": KunUI}),
			Namespace:     install.Namespace,
		},
	); err != nil {
		e.Logger.Error(err, "failed to get Deployments")
		return
	}
	if len(deploys.Items) == 0 {
		if err := DeployUI(e.Client, context.Background(), install); err != nil {
			e.Logger.Error(err, "failed to deploy Kun UI")
			return
		}
	}

	if err := e.Client.List(context.TODO(), &deploys,
		&client.ListOptions{
			LabelSelector: labels.SelectorFromSet(labels.Set{"l6p-app": KunApi}),
			Namespace:     install.Namespace,
		},
	); err != nil {
		e.Logger.Error(err, "failed to get Deployments")
		return
	}
	if len(deploys.Items) == 0 {
		if err := DeployServer(e.Client, context.Background(), install); err != nil {
			e.Logger.Error(err, "failed to deploy Kun server")
			return
		}
	}
}

func (e EnqueueRequestForKunInstallation) Update(updateEvent event.UpdateEvent, _ workqueue.RateLimitingInterface) {
	oldInstall := updateEvent.ObjectOld.(*kunapi.KunInstallation)
	newInstall := updateEvent.ObjectNew.(*kunapi.KunInstallation)

	if err := UndeployUI(e.Client, context.Background(), oldInstall.Namespace); err != nil {
		e.Logger.Error(err, "failed to undeploy Kun UI")
		return
	}

	if err := UndeployServer(e.Client, context.Background(), oldInstall.Namespace); err != nil {
		e.Logger.Error(err, "failed to undeploy Kun server")
		return
	}

	if err := DeployUI(e.Client, context.Background(), newInstall); err != nil {
		e.Logger.Error(err, "failed to deploy Kun UI")
		return
	}

	if err := DeployServer(e.Client, context.Background(), newInstall); err != nil {
		e.Logger.Error(err, "failed to deploy Kun server")
		return
	}
}

func (e EnqueueRequestForKunInstallation) Delete(deleteEvent event.DeleteEvent, _ workqueue.RateLimitingInterface) {
	install := deleteEvent.Object.(*kunapi.KunInstallation)

	if err := UndeployUI(e.Client, context.Background(), install.Namespace); err != nil {
		e.Logger.Error(err, "failed to undeploy Kun UI")
		return
	}

	if err := UndeployServer(e.Client, context.Background(), install.Namespace); err != nil {
		e.Logger.Error(err, "failed to undeploy Kun server")
		return
	}
}

func (e EnqueueRequestForKunInstallation) Generic(event.GenericEvent, workqueue.RateLimitingInterface) {
}

// SetupWithManager sets up the controller with the Manager.
func (r *KunInstallationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	logger := logf.Log.WithName("kuninstallation-controller")

	return ctrl.NewControllerManagedBy(mgr).
		For(&kunapi.KunInstallation{}).
		Watches(
			&source.Kind{Type: &kunapi.KunInstallation{}},
			&EnqueueRequestForKunInstallation{Logger: logger, Client: r.Client},
		).
		Complete(r)
}

func DeployServer(c client.Client, ctx context.Context, install *kunapi.KunInstallation) error {
	if err := templates.CreateServiceAccount(
		c, ctx, KunApi, install.Namespace, templates.ServerServiceAccountTemplate,
	); err != nil {
		return err
	}

	if err := templates.CreateClusterRole(
		c, ctx, KunApi, templates.ServerClusterRoleTemplate,
	); err != nil {
		return err
	}

	if err := templates.CreateClusterRoleBinding(
		c, ctx, KunApi, install.Namespace, templates.ServerClusterRoleBindingTemplate,
	); err != nil {
		return err
	}

	if err := templates.CreateDeployment(
		c, ctx, KunApi, install.Namespace, templates.ServerDeployTemplate,
		install.Spec.Server.Replicas,
		path.Join(install.Spec.Hub, fmt.Sprintf("%s:latest", KunApi)),
		&install.Spec.Server.Resources,
		[]core.EnvVar{
			{
				Name:  "MONGODB_ADDR",
				Value: install.Spec.Mongodb.Addr,
			},
			{
				Name:  "MONGODB_USER",
				Value: install.Spec.Mongodb.User,
			},
			{
				Name:  "MONGODB_PASS",
				Value: install.Spec.Mongodb.Pass,
			},
		},
	); err != nil {
		return err
	}

	if err := templates.CreateService(
		c, ctx, KunApi, install.Namespace, templates.ServerServiceTemplate,
		*install.Spec.Server.Port,
	); err != nil {
		return err
	}

	if install.Spec.Server.Ingress.Enabled {
		if err := templates.CreateIngress(
			c, ctx, KunApi, install.Namespace, templates.ServerIngressTemplate,
			install.Spec.Server.Ingress.Host, install.Spec.Server.Ingress.Path, *install.Spec.Server.Port,
		); err != nil {
			return err
		}
	}
	return nil
}

func UndeployServer(c client.Client, ctx context.Context, namespace string) error {
	if err := templates.DeleteIngress(c, ctx, KunApi, namespace); err != nil {
		return err
	}

	if err := templates.DeleteService(c, ctx, KunApi, namespace); err != nil {
		return err
	}

	if err := templates.DeleteDeployment(c, ctx, KunApi, namespace); err != nil {
		return err
	}

	if err := templates.DeleteClusterRole(c, ctx, KunApi); err != nil {
		return err
	}

	if err := templates.DeleteClusterRoleBinding(c, ctx, KunApi); err != nil {
		return err
	}

	if err := templates.DeleteServiceAccount(c, ctx, KunApi, namespace); err != nil {
		return err
	}
	return nil
}

func DeployUI(c client.Client, ctx context.Context, install *kunapi.KunInstallation) error {
	if !install.Spec.UI.Enabled {
		return nil
	}

	if err := templates.CreateServiceAccount(
		c, ctx, KunUI, install.Namespace, templates.UIServiceAccountTemplate,
	); err != nil {
		return err
	}

	if err := templates.CreateDeployment(
		c, ctx, KunUI, install.Namespace, templates.UIDeployTemplate,
		install.Spec.UI.Replicas,
		path.Join(install.Spec.Hub, fmt.Sprintf("%s:latest", KunUI)),
		&install.Spec.UI.Resources,
		nil,
	); err != nil {
		return err
	}

	if err := templates.CreateService(
		c, ctx, KunUI, install.Namespace, templates.UIServiceTemplate,
		*install.Spec.UI.Port,
	); err != nil {
		return err
	}

	if install.Spec.UI.Ingress.Enabled {
		if err := templates.CreateIngress(
			c, ctx, KunUI, install.Namespace, templates.UIIngressTemplate,
			install.Spec.UI.Ingress.Host, install.Spec.UI.Ingress.Path, *install.Spec.UI.Port,
		); err != nil {
			return err
		}
	}
	return nil
}

func UndeployUI(c client.Client, ctx context.Context, namespace string) error {
	if err := templates.DeleteServiceAccount(c, ctx, KunUI, namespace); err != nil {
		return err
	}

	if err := templates.DeleteDeployment(c, ctx, KunUI, namespace); err != nil {
		return err
	}

	if err := templates.DeleteService(c, ctx, KunUI, namespace); err != nil {
		return err
	}

	if err := templates.DeleteIngress(c, ctx, KunUI, namespace); err != nil {
		return err
	}
	return nil
}
