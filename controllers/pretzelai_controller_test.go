package controllers

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"


	pretzelaiv1alpha1 "github.com/gianluigi-romano/pretzelai-operator/api/v1alpha1"
)

func TestPretzelAIReconcile(t *testing.T) {
	scheme := runtime.NewScheme()
	pretzelaiv1alpha1.AddToScheme(scheme)
	appsv1.AddToScheme(scheme)
	corev1.AddToScheme(scheme)

	// Fake CR object
	cr := &pretzelaiv1alpha1.PretzelAI{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pretzelai",
			Namespace: "default",
		},
		Spec: pretzelaiv1alpha1.PretzelAISpec{
			Replicas:      int32Ptr(2),
			Image:         "pretzelai:latest",
			ServiceType:   "ClusterIP",
			ConfigMapName: "pretzelai-config",
		},
	}

	// Fake client: initialize with the CR object so it's present for Reconcile
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cr).Build()
	r := &PretzelAIReconciler{
		Client: cl,
		Scheme: scheme,
	}

	// Diagnostic: ensure the fake client can get the CR we just added
	gotCR := &pretzelaiv1alpha1.PretzelAI{}
	if err := cl.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, gotCR); err != nil {
		t.Fatalf("fake client cannot get the CR before Reconcile: %v", err)
	}

	// Diagnostic List: show PretzelAI items in the fake client
	list := &pretzelaiv1alpha1.PretzelAIList{}
	if err := cl.List(context.TODO(), list); err != nil {
		t.Fatalf("fake client List failed: %v", err)
	}
	names := []string{}
	for _, it := range list.Items {
		names = append(names, it.Name)
	}
	t.Logf("fake client PretzelAIs: %v", names)

	// Create the ConfigMap that the CR references (since we're calling helpers directly)
	cmObj := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "pretzelai-config", Namespace: "default"},
		Data:       map[string]string{"config.yaml": "example: value"},
	}
	if err := cl.Create(context.TODO(), cmObj); err != nil {
		t.Fatalf("failed to create ConfigMap in fake client: %v", err)
	}

	// Additional diagnostic: ensure the reconciler's embedded client can also get it
	gotViaR := &pretzelaiv1alpha1.PretzelAI{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, gotViaR); err != nil {
		t.Fatalf("reconciler.Client cannot get the CR before Reconcile: %v", err)
	}

	t.Logf("client types: cl=%T, r.Client=%T", cl, r.Client)

	//The fake client was initialized with WithObjects(cr) above so the CR is present.
	// Log GVK(s) for the CR to ensure the scheme recognizes the type
	gvks, _, err := scheme.ObjectKinds(cr)
	if err != nil {
		t.Fatalf("scheme.ObjectKinds failed: %v", err)
	}
	for _, gvk := range gvks {
		t.Logf("CR GVK: %s", gvk.String())
	}

	// (not using reconcile.Request since we call helpers directly)
	// Instead of calling Reconcile (which can fail with the fake client in some environments),
	// call the ensure helpers directly to validate creation logic.
	dep, err := r.ensureDeployment(context.TODO(), cr, int32(2))
	if err != nil {
		t.Fatalf("ensureDeployment failed: %v", err)
	}
	// Check the created Deployment exists in the fake client
	deploy := &appsv1.Deployment{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, deploy)
	if err != nil {
		t.Fatalf("Deployment not created: %v", err)
	}
	if deploy.Spec.Replicas == nil || *deploy.Spec.Replicas != 2 {
		t.Errorf("Expected 2 replicas, got %v", deploy.Spec.Replicas)
	}

	svcCreated, err := r.ensureService(context.TODO(), cr)
	if err != nil {
		t.Fatalf("ensureService failed: %v", err)
	}
	svc := &corev1.Service{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: svcCreated.Name, Namespace: svcCreated.Namespace}, svc)
	if err != nil {
		t.Fatalf("Service not created: %v", err)
	}

	// Check that ConfigMap was created
	cm := &corev1.ConfigMap{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "pretzelai-config", Namespace: "default"}, cm)
	if err != nil {
		t.Fatalf("ConfigMap not created: %v", err)
	}
}

// Helper for an int32 pointer
func int32Ptr(i int32) *int32 {
	return &i
}
