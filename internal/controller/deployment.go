package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *PIDScalerReconciler) GetDeployment(ctx context.Context, namespaceName string, deploymentName string) (*appsv1.Deployment, bool) {
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: namespaceName}, deployment)
	if err != nil {
		r.Log.Error(err, "Failed to get deployment")
		return nil, false
	}
	return deployment, true
}

func (r *PIDScalerReconciler) ScaleReplicas(ctx context.Context, namespace string, deployment string, replicas int32) error {
	dep, found := r.GetDeployment(ctx, namespace, deployment)
	if found == false {
		return fmt.Errorf("No deployment %s found in ns %s", deployment, namespace)
	}
	if dep.Spec.Replicas != nil && *dep.Spec.Replicas != replicas {
		dep.Spec.Replicas = &replicas
		err := r.Update(ctx, dep)
		if err != nil {
			r.Log.Error(err, "Failed to scale deployment", "deployment", deployment, "namespace", namespace, "replicas", replicas)
			return err
		}
		r.Log.Info("Deployment scaled", "deployment", deployment, "namespace", namespace, "replicas", replicas)
	}
	return nil
}
