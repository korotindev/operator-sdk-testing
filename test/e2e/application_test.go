package e2e

import (
	goCtx "context"
	"fmt"
	"testing"
	"time"

	"github.com/soxat/operator-sdk-testing/pkg/apis"
	operator "github.com/soxat/operator-sdk-testing/pkg/apis/app/v1alpha1"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestApplication(t *testing.T) {
	applicationList := &operator.ApplicationList{}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, applicationList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	// run subtests
	t.Run("application-group", func(t *testing.T) {
		t.Run("Cluster", ApplicationCluster)
		t.Run("Cluster2", ApplicationCluster)
	})
}

func ApplicationScaleTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetOperatorNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}
	// create application custom resource
	exampleApplication := &operator.Application{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "example-application",
			Namespace: namespace,
		},
		Spec: operator.ApplicationSpec{
			Containers: []operator.ApplicationContainer{
				{
					Name:  "nginx",
					Image: "nginx:latest",
					Ports: []operator.ApplicationContainerPort{
						{
							Name:          "default",
							HostPort:      80,
							ContainerPort: 80,
						},
					},
					CPULimit:    "test",
					MemoryLimit: "test",
				},
			},
			Replicas: pointer.Int32Ptr(2),
		},
	}

	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(goCtx.TODO(), exampleApplication, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}
	// wait for exampleApplication to reach 3 replicas
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, exampleApplication.Name, 2, retryInterval, timeout)
	if err != nil {
		return err
	}

	err = f.Client.Get(goCtx.TODO(), types.NamespacedName{Name: exampleApplication.Name, Namespace: namespace}, exampleApplication)
	if err != nil {
		return err
	}
	exampleApplication.Spec.Replicas = pointer.Int32Ptr(3)
	err = f.Client.Update(goCtx.TODO(), exampleApplication)
	if err != nil {
		return err
	}

	// wait for exampleApplication to reach 3 replicas
	return e2eutil.WaitForDeployment(t, f.KubeClient, namespace, exampleApplication.Name, 3, retryInterval, timeout)
}

func ApplicationCluster(t *testing.T) {
	t.Parallel()
	ctx := framework.NewContext(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetOperatorNamespace()
	if err != nil {
		t.Fatal(err)
	}
	// get global framework variables
	f := framework.Global
	// wait for application-operator to be ready
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "application-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	if err = ApplicationScaleTest(t, f, ctx); err != nil {
		t.Fatal(err)
	}
}
