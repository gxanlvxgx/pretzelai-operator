/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
*/

package controllers

import (
	"context"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	pretzelaiv1alpha1 "github.com/gianluigi-romano/pretzelai-operator/api/v1alpha1"
)

// PretzelAIReconciler reconciles a PretzelAI object
type PretzelAIReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// RBAC Permissions
//+kubebuilder:rbac:groups=pretzelai.pretzelai.local,resources=pretzelais,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pretzelai.pretzelai.local,resources=pretzelais/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pretzelai.pretzelai.local,resources=pretzelais/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PretzelAIReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// 1. Fetch the PretzelAI instance
	var pretzelAI pretzelaiv1alpha1.PretzelAI
	if err := r.Get(ctx, req.NamespacedName, &pretzelAI); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 2. Handle Finalizer (Cleanup logic)
	const pretzelAIFinalizer = "pretzelai.finalizers.pretzelai.local"
	if !pretzelAI.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is being deleted
		if containsString(pretzelAI.ObjectMeta.Finalizers, pretzelAIFinalizer) {

			// Remove the finalizer to allow deletion
			pretzelAI.ObjectMeta.Finalizers = removeString(pretzelAI.ObjectMeta.Finalizers, pretzelAIFinalizer)
			if err := r.Update(ctx, &pretzelAI); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if not present
	if !containsString(pretzelAI.ObjectMeta.Finalizers, pretzelAIFinalizer) {
		pretzelAI.ObjectMeta.Finalizers = append(pretzelAI.ObjectMeta.Finalizers, pretzelAIFinalizer)
		if err := r.Update(ctx, &pretzelAI); err != nil {
			return ctrl.Result{}, err
		}
	}

	// 3. Ensure ConfigMap exists if specified in the CR
	if pretzelAI.Spec.ConfigMapName != "" {
		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pretzelAI.Spec.ConfigMapName,
				Namespace: req.Namespace,
			},
			Data: map[string]string{
				"config.yaml": "example: value",
			},
		}
		if err := ctrl.SetControllerReference(&pretzelAI, configMap, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		var existing corev1.ConfigMap
		err := r.Get(ctx, client.ObjectKey{Name: configMap.Name, Namespace: configMap.Namespace}, &existing)
		if err != nil {
			if client.IgnoreNotFound(err) != nil {
				return ctrl.Result{}, err
			}
			if err := r.Create(ctx, configMap); err != nil {
				return ctrl.Result{}, err
			}
		} else if !reflect.DeepEqual(configMap.Data, existing.Data) {
			existing.Data = configMap.Data
			if err := r.Update(ctx, &existing); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	// 4. Ensure Deployment and Service exist and match the desired state
	replicas := int32(1)
	if pretzelAI.Spec.Replicas != nil {
		replicas = *pretzelAI.Spec.Replicas
	}

	dep, err := r.ensureDeployment(ctx, &pretzelAI, replicas)
	if err != nil {
		return ctrl.Result{}, err
	}

	svc, err := r.ensureService(ctx, &pretzelAI)
	if err != nil {
		return ctrl.Result{}, err
	}

	// 5. Update CR Status
	if pretzelAI.Status.ReadyReplicas != dep.Status.ReadyReplicas {
		pretzelAI.Status.ReadyReplicas = dep.Status.ReadyReplicas
	}
	if pretzelAI.Status.AvailableReplicas != dep.Status.AvailableReplicas {
		pretzelAI.Status.AvailableReplicas = dep.Status.AvailableReplicas
	}
	
	if svc != nil {
		pretzelAI.Status.ServiceStatus = string(svc.Spec.Type)
	}

	if err := r.Status().Update(ctx, &pretzelAI); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Successfully reconciled PretzelAI", "name", pretzelAI.Name)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PretzelAIReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pretzelaiv1alpha1.PretzelAI{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}

// ==============================================================================
// Helper Functions
// ==============================================================================

// ensureDeployment creates or updates a Deployment for the PretzelAI CR.
func (r *PretzelAIReconciler) ensureDeployment(ctx context.Context, pretzelAI *pretzelaiv1alpha1.PretzelAI, replicas int32) (*appsv1.Deployment, error) {
	// PretzelAI (Jupyter based) typically runs on port 8888
	const containerPort = 8888

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pretzelAI.Name,
			Namespace: pretzelAI.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": pretzelAI.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": pretzelAI.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "pretzelai",
						Image: "pretzelai:local", // Using local image for development
						// PullIfNotPresent is important for local testing with Kind/Minikube
						ImagePullPolicy: corev1.PullIfNotPresent,
						Ports: []corev1.ContainerPort{{
							ContainerPort: containerPort,
						}},
					}},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(pretzelAI, dep, r.Scheme); err != nil {
		return nil, err
	}

	var existing appsv1.Deployment
	err := r.Get(ctx, client.ObjectKey{Name: dep.Name, Namespace: dep.Namespace}, &existing)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return nil, err
		}
		// Create if it doesn't exist
		if err := r.Create(ctx, dep); err != nil {
			return nil, err
		}
		return dep, nil
	}

	// Check for Drift: Update only if Replicas or Image changed to avoid infinite loops
	needsUpdate := false

	if existing.Spec.Replicas == nil || dep.Spec.Replicas == nil || *existing.Spec.Replicas != *dep.Spec.Replicas {
		needsUpdate = true
	}
	// Check container image 
	if len(existing.Spec.Template.Spec.Containers) > 0 &&
		existing.Spec.Template.Spec.Containers[0].Image != dep.Spec.Template.Spec.Containers[0].Image {
		needsUpdate = true
	}

	if needsUpdate {
		// Update the existing object with desired specs
		existing.Spec.Replicas = dep.Spec.Replicas
		existing.Spec.Template.Spec.Containers[0].Image = dep.Spec.Template.Spec.Containers[0].Image

		if err := r.Update(ctx, &existing); err != nil {
			return nil, err
		}
	}

	return &existing, nil
}

// ensureService creates or updates a Service for the PretzelAI CR.
func (r *PretzelAIReconciler) ensureService(ctx context.Context, pretzelAI *pretzelaiv1alpha1.PretzelAI) (*corev1.Service, error) {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pretzelAI.Name,
			Namespace: pretzelAI.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": pretzelAI.Name},
			Ports: []corev1.ServicePort{{
				Port:       80,
				TargetPort: intstr.FromInt(8888), // Targeting the container port 8888
			}},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	if err := ctrl.SetControllerReference(pretzelAI, svc, r.Scheme); err != nil {
		return nil, err
	}

	var existing corev1.Service
	err := r.Get(ctx, client.ObjectKey{Name: svc.Name, Namespace: svc.Namespace}, &existing)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return nil, err
		}
		if err := r.Create(ctx, svc); err != nil {
			return nil, err
		}
		return svc, nil
	}

	if !reflect.DeepEqual(svc.Spec.Ports, existing.Spec.Ports) {
		existing.Spec.Ports = svc.Spec.Ports
		if err := r.Update(ctx, &existing); err != nil {
			return nil, err
		}
	}
	return &existing, nil
}

// Helper functions for Finalizers
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) []string {
	var result []string
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}
