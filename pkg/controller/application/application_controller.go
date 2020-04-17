package application

import (
	"context"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	appV1Alpha1 "github.com/soxat/operator-sdk-testing/pkg/apis/app/v1alpha1"
	appsV1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_application")

// Add creates a new Application Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileApplication{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("application-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Application
	err = c.Watch(&source.Kind{Type: &appV1Alpha1.Application{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Deployment and requeue the owner Application
	err = c.Watch(&source.Kind{Type: &appsV1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appV1Alpha1.Application{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileApplication implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileApplication{}

// ReconcileApplication reconciles a Application object
type ReconcileApplication struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Application object and makes changes based on the state read
// and what is in the Application.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileApplication) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request", request)
	reqLogger.Info("Reconciling Application")

	// Fetch the Application instance
	applicationInstance := &appV1Alpha1.Application{}
	err := r.client.Get(context.TODO(), request.NamespacedName, applicationInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not foundDeployment, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("Application resource not foundDeployment. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Check if this Deployment already exists
	foundDeployment := &appsV1.Deployment{}
	searchDeploymentOpts := types.NamespacedName{
		Name:      applicationInstance.Name,
		Namespace: applicationInstance.Namespace,
	}
	err = r.client.Get(context.TODO(), searchDeploymentOpts, foundDeployment)

	if err != nil {
		if errors.IsNotFound(err) {
			// Define a new Deployment object
			deployment := newDeploymentForApplication(applicationInstance)

			// Set Application applicationInstance as the owner and controller
			err := controllerutil.SetControllerReference(applicationInstance, deployment, r.scheme)
			if err != nil {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating a new Deployment",
				"Deployment.Namespace", deployment.Namespace,
				"Deployment.Name", deployment.Name)

			err = r.client.Create(context.TODO(), deployment)
			if err != nil {
				reqLogger.Error(err, "Failed to create new Deployment",
					"Deployment.Namespace", deployment.Namespace,
					"Deployment.Name", deployment.Name)
				return reconcile.Result{}, err
			}

			// Deployment created successfully - don't requeue
			return reconcile.Result{}, nil
		}

		reqLogger.Error(err, "Failed to get Deployment")
		return reconcile.Result{}, err
	}

	// Ensure the deployment size is the same as the spec
	replicas := applicationInstance.Spec.Replicas
	if *foundDeployment.Spec.Replicas != *replicas {
		foundDeployment.Spec.Replicas = replicas
		err = r.client.Update(context.TODO(), foundDeployment)
		if err != nil {
			reqLogger.Error(err, "Failed to update Deployment",
				"Deployment.Namespace", foundDeployment.Namespace,
				"Deployment.Name", foundDeployment.Name)
			return reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}

	// Update the Application status with the pod names
	// List the pods for this application's deployment
	podList := &coreV1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(applicationInstance.Namespace),
		client.MatchingLabels(labelsForApplication(applicationInstance.Name)),
	}

	err = r.client.List(context.TODO(), podList, listOpts...)
	if err != nil {
		reqLogger.Error(err, "Failed to list pods",
			"Deployment.Namespace", applicationInstance.Namespace,
			"Deployment.Name", applicationInstance.Name)
		return reconcile.Result{}, err
	}

	podNames := getPodNames(podList.Items)

	if !reflect.DeepEqual(podNames, applicationInstance.Status.Pods) {
		applicationInstance.Status.Pods = podNames
		err := r.client.Status().Update(context.TODO(), applicationInstance)
		if err != nil {
			reqLogger.Error(err, "Failed to update Application status")
			return reconcile.Result{}, err
		}
	}

	// Update the Application status with the deployment replicas
	deploymentReplicas := foundDeployment.Status.Replicas
	if applicationInstance.Status.Replicas != deploymentReplicas {
		applicationInstance.Status.Replicas = deploymentReplicas
		err := r.client.Status().Update(context.TODO(), applicationInstance)
		if err != nil {
			reqLogger.Error(err, "Failed to update Application status")
			return reconcile.Result{}, err
		}
	}

	reqLogger.Info("Success!")

	return reconcile.Result{}, nil
}

// newDeploymentForApplication returns a application deployment
func newDeploymentForApplication(application *appV1Alpha1.Application) *appsV1.Deployment {
	labels := labelsForApplication(application.Name)
	containers := buildContainersForApplication(application)

	return &appsV1.Deployment{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      application.Name,
			Namespace: application.Namespace,
		},
		Spec: appsV1.DeploymentSpec{
			Replicas: application.Spec.Replicas,
			Selector: &metaV1.LabelSelector{
				MatchLabels: labels,
			},
			Template: coreV1.PodTemplateSpec{
				ObjectMeta: metaV1.ObjectMeta{
					Labels: labels,
				},
				Spec: coreV1.PodSpec{
					Containers: containers,
				},
			},
		},
	}
}

func buildContainersForApplication(application *appV1Alpha1.Application) []coreV1.Container {
	var containers []coreV1.Container

	for _, appContainer := range application.Spec.Containers {
		container := buildContainerForApplicationContainer(appContainer)
		containers = append(containers, container)
	}

	return containers
}

func buildContainerForApplicationContainer(applicationContainer appV1Alpha1.ApplicationContainer) coreV1.Container {
	var ports []coreV1.ContainerPort

	for _, appContainerPort := range applicationContainer.Ports {
		port := coreV1.ContainerPort{
			Name:          appContainerPort.Name,
			ContainerPort: appContainerPort.ContainerPort,
			HostPort:      appContainerPort.HostPort,
		}
		ports = append(ports, port)
	}

	return coreV1.Container{
		Name:  applicationContainer.Name,
		Image: applicationContainer.Image,
		Ports: ports,
	}
}

// labelsForApplication returns the labels for selecting the resources
// belonging to the given application CR name.
func labelsForApplication(name string) map[string]string {
	return map[string]string{"app": "application", "application_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []coreV1.Pod) []string {
	var podNames []string

	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}

	return podNames
}
