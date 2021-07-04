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
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"github.com/go-logr/logr"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/workqueue"
	kunapi "l6p.io/KunOperator/api/v1"
	"net/http"
	"path"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

const KunApi = "kun-api"
const KunUI = "kun-ui"

var StartTime = time.Now()
var ServerNamespaces = make([]string, 0)
var ServerImage = ""
var UIImage = ""

var PodCreateMessageChannel = make(chan string, 10000)

//go:embed templates/server/service-account.yaml
var ServerServiceAccountTemplate string

//go:embed templates/server/deployment.yaml
var ServerDeployTemplate string

//go:embed templates/server/service.yaml
var ServerServiceTemplate string

//go:embed templates/server/ingress.yaml
var ServerIngressTemplate string

//go:embed templates/ui/service-account.yaml
var UIServiceAccountTemplate string

//go:embed templates/ui/deployment.yaml
var UIDeployTemplate string

//go:embed templates/ui/service.yaml
var UIServiceTemplate string

//go:embed templates/ui/ingress.yaml
var UIIngressTemplate string

// KunInstallationReconciler reconciles a KunInstallation object
type KunInstallationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=l6p.io,resources=kuninstallations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=l6p.io,resources=kuninstallations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=l6p.io,resources=kuninstallations/finalizers,verbs=update

//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses,verbs=get;list;watch;create;update;patch;delete

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

//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete

type EnqueueRequestForPod struct {
	Logger logr.Logger
}

func (e EnqueueRequestForPod) Create(event event.CreateEvent, _ workqueue.RateLimitingInterface) {
	pod := event.Object.(*core.Pod)
	if StartTime.After(pod.CreationTimestamp.Time) {
		return
	}

	for _, c := range pod.Spec.Containers {
		if c.Image == ServerImage || c.Image == UIImage {
			continue
		}
		e.Logger.Info(fmt.Sprintf("image '%s' is used to create pod '%s'", c.Image, pod.Name))
		PodCreateMessageChannel <- c.Image
	}
}

func PodCreateMessageProcessor(logger logr.Logger) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	sendRequest := func(namespace string, image string) {
		req, err := http.NewRequest(
			"POST",
			fmt.Sprintf("http://kun-api.%s/api/v1/img/up", namespace),
			bytes.NewBuffer([]byte(fmt.Sprintf(`{"image":"%s"}`, image))),
		)
		if err != nil {
			logger.Error(err, "failed to create http request")
			return
		}

		req.Header.Set("Content-Type", "application/json")
		resp, err := httpClient.Do(req)
		if err != nil {
			logger.Error(err, "failed to send http request")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			logger.Error(err, "API invocation failed")
			return
		}
	}

	for {
		image := <-PodCreateMessageChannel
		for _, namespace := range ServerNamespaces {
			sendRequest(namespace, image)
		}
	}
}

func (e EnqueueRequestForPod) Update(event.UpdateEvent, workqueue.RateLimitingInterface) {
}

func (e EnqueueRequestForPod) Delete(event.DeleteEvent, workqueue.RateLimitingInterface) {
}

func (e EnqueueRequestForPod) Generic(event.GenericEvent, workqueue.RateLimitingInterface) {
}

type EnqueueRequestForKunInstallation struct {
	Logger logr.Logger
	client.Client
}

func (e EnqueueRequestForKunInstallation) Create(createEvent event.CreateEvent, _ workqueue.RateLimitingInterface) {
	install := createEvent.Object.(*kunapi.KunInstallation)

	exist := false
	for _, namespace := range ServerNamespaces {
		if namespace == install.Namespace {
			exist = true
			break
		}
	}
	if !exist {
		ServerNamespaces = append(ServerNamespaces, install.Namespace)
	}

	if StartTime.After(install.CreationTimestamp.Time) {
		return
	}

	if err := DeployUI(e.Client, context.Background(), install); err != nil {
		e.Logger.Error(err, "failed to deploy Kun UI")
		return
	}

	if err := DeployServer(e.Client, context.Background(), install); err != nil {
		e.Logger.Error(err, "failed to deploy Kun server")
		return
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

	tmp := make([]string, 0)
	for _, namespace := range ServerNamespaces {
		if namespace != install.Namespace {
			tmp = append(tmp, namespace)
		}
	}
	ServerNamespaces = tmp

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
	go PodCreateMessageProcessor(logger)

	return ctrl.NewControllerManagedBy(mgr).
		For(&kunapi.KunInstallation{}).
		Watches(
			&source.Kind{Type: &kunapi.KunInstallation{}},
			&EnqueueRequestForKunInstallation{Logger: logger, Client: r.Client},
		).
		Watches(
			&source.Kind{Type: &core.Pod{}},
			&EnqueueRequestForPod{Logger: logger},
		).
		Complete(r)
}

func DeployServer(c client.Client, ctx context.Context, install *kunapi.KunInstallation) error {
	if err := CreateServiceAccount(
		c, ctx, KunApi, install.Namespace, ServerServiceAccountTemplate,
	); err != nil {
		return err
	}

	ServerImage = path.Join(install.Spec.Hub, fmt.Sprintf("%s:latest", KunApi))
	if err := CreateDeployment(
		c, ctx, KunApi, install.Namespace, ServerDeployTemplate,
		install.Spec.Server.Replicas, ServerImage, &install.Spec.Server.Resources,
		[]core.EnvVar{
			{
				Name:  "CLICKHOUSE_ADDR",
				Value: install.Spec.Clickhouse.Addr,
			},
		},
	); err != nil {
		return err
	}

	if err := CreateService(
		c, ctx, KunApi, install.Namespace, ServerServiceTemplate,
		*install.Spec.Server.Port,
	); err != nil {
		return err
	}

	if install.Spec.Server.Ingress.Enabled {
		if err := CreateIngress(
			c, ctx, KunApi, install.Namespace, ServerIngressTemplate,
			install.Spec.Server.Ingress.Host, install.Spec.Server.Ingress.Path, *install.Spec.Server.Port,
		); err != nil {
			return err
		}
	}
	return nil
}

func UndeployServer(c client.Client, ctx context.Context, namespace string) error {
	if err := DeleteServiceAccount(c, ctx, KunApi, namespace); err != nil {
		return err
	}

	if err := DeleteDeployment(c, ctx, KunApi, namespace); err != nil {
		return err
	}

	if err := DeleteService(c, ctx, KunApi, namespace); err != nil {
		return err
	}

	if err := DeleteIngress(c, ctx, KunApi, namespace); err != nil {
		return err
	}
	return nil
}

func DeployUI(c client.Client, ctx context.Context, install *kunapi.KunInstallation) error {
	if !install.Spec.UI.Enabled {
		return nil
	}

	if err := CreateServiceAccount(
		c, ctx, KunUI, install.Namespace, UIServiceAccountTemplate,
	); err != nil {
		return err
	}

	UIImage = path.Join(install.Spec.Hub, fmt.Sprintf("%s:latest", KunUI))
	if err := CreateDeployment(
		c, ctx, KunUI, install.Namespace, UIDeployTemplate,
		install.Spec.UI.Replicas, UIImage, &install.Spec.UI.Resources, nil,
	); err != nil {
		return err
	}

	if err := CreateService(
		c, ctx, KunUI, install.Namespace, UIServiceTemplate,
		*install.Spec.UI.Port,
	); err != nil {
		return err
	}

	if install.Spec.UI.Ingress.Enabled {
		if err := CreateIngress(
			c, ctx, KunUI, install.Namespace, UIIngressTemplate,
			install.Spec.UI.Ingress.Host, install.Spec.UI.Ingress.Path, *install.Spec.UI.Port,
		); err != nil {
			return err
		}
	}
	return nil
}

func UndeployUI(c client.Client, ctx context.Context, namespace string) error {
	if err := DeleteServiceAccount(c, ctx, KunUI, namespace); err != nil {
		return err
	}

	if err := DeleteDeployment(c, ctx, KunUI, namespace); err != nil {
		return err
	}

	if err := DeleteService(c, ctx, KunUI, namespace); err != nil {
		return err
	}

	if err := DeleteIngress(c, ctx, KunUI, namespace); err != nil {
		return err
	}
	return nil
}

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
	serviceAccount, err := GetServiceAccount(c, ctx, name, namespace)
	if err != nil {
		return err
	} else if serviceAccount == nil {
		decoder := serializer.NewCodecFactory(scheme.Scheme).UniversalDecoder()

		var newServiceAccount core.ServiceAccount
		if err := runtime.DecodeInto(decoder, []byte(template), &newServiceAccount); err != nil {
			return err
		}
		newServiceAccount.Name = name
		newServiceAccount.Namespace = namespace

		if err := c.Create(ctx, &newServiceAccount); err != nil {
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
	deploy, err := GetDeployment(c, ctx, name, namespace)
	if err != nil {
		return err
	} else if deploy == nil {
		decoder := serializer.NewCodecFactory(scheme.Scheme).UniversalDecoder()

		var newDeploy apps.Deployment
		if err := runtime.DecodeInto(decoder, []byte(template), &newDeploy); err != nil {
			return err
		}

		newDeploy.Name = name
		newDeploy.Namespace = namespace
		newDeploy.Spec.Replicas = replicas

		container := &newDeploy.Spec.Template.Spec.Containers[0]
		container.Image = image
		container.Resources = *resources

		if env != nil {
			container.Env = env
		}

		if err := c.Create(ctx, &newDeploy); err != nil {
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
	service, err := GetService(c, ctx, name, namespace)
	if err != nil {
		return err
	} else if service == nil {
		decoder := serializer.NewCodecFactory(scheme.Scheme).UniversalDecoder()

		var newService core.Service
		if err := runtime.DecodeInto(decoder, []byte(template), &newService); err != nil {
			return err
		}
		newService.Name = name
		newService.Namespace = namespace
		newService.Spec.Ports[0].Port = port

		if err := c.Create(ctx, &newService); err != nil {
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
	ingress, err := GetIngress(c, ctx, name, namespace)
	if err != nil {
		return err
	} else if ingress == nil {
		decoder := serializer.NewCodecFactory(scheme.Scheme).UniversalDecoder()

		var newIngress networking.Ingress
		if err := runtime.DecodeInto(decoder, []byte(template), &newIngress); err != nil {
			return err
		}
		newIngress.Name = name
		newIngress.Namespace = namespace

		pathTypePrefix := networking.PathTypePrefix
		newIngress.Spec.Rules = []networking.IngressRule{
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
		if err := c.Create(ctx, &newIngress); err != nil {
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
