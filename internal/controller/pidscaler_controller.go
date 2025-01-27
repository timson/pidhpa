/*
Copyright 2024.

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

package controller

import (
	"context"
	"sync"

	"github.com/timson/pidhpa-operator/internal/storage"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	pidscalerv1 "github.com/timson/pidhpa-operator/api/v1"
	"github.com/timson/pidhpa-operator/internal/metrics"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PIDScalerReconciler reconciles a PIDScaler object
type PIDScalerReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	Log             logr.Logger
	Storage         *storage.PIDScalerStateStorage
	OperatorContext context.Context
	wg              *sync.WaitGroup
	m               sync.Mutex
}

func (r *PIDScalerReconciler) GetCRD(ctx context.Context, namespacedName client.ObjectKey) (pidscalerv1.PIDScaler, error) {
	var pidScaler pidscalerv1.PIDScaler
	err := r.Client.Get(ctx, namespacedName, &pidScaler)
	if err != nil {
		metrics.CRDFetchErrors.WithLabelValues(namespacedName.String()).Inc()
	}
	return pidScaler, err
}

func (r *PIDScalerReconciler) updateStatus(ctx context.Context, namespacedName client.ObjectKey, status string, message string) error {
	r.m.Lock()
	defer r.m.Unlock()

	pidScaler, err := r.GetCRD(ctx, namespacedName)
	if err != nil {
		metrics.CRDFetchErrors.WithLabelValues(namespacedName.String()).Inc()
		return err
	}

	pidScaler.Status = pidscalerv1.PIDScalerStatus{
		Status:     status,
		Message:    message,
		UpdateTime: metav1.Now(),
	}
	if err = r.Update(ctx, &pidScaler); err != nil {
		metrics.CRDUpdateErrors.WithLabelValues(namespacedName.String()).Inc()
		return err
	}

	return nil
}

func (r *PIDScalerReconciler) updateDesiredReplicas(ctx context.Context, namespacedName client.ObjectKey, replicas int32) error {
	r.m.Lock()
	defer r.m.Unlock()
	pidScaler, err := r.GetCRD(ctx, namespacedName)
	if err != nil {
		metrics.CRDFetchErrors.WithLabelValues(namespacedName.String()).Inc()
		return err
	}
	pidScaler.Spec.Target.DesiredReplicas = &replicas
	pidScaler.Spec.Target.UpdateTime = metav1.Now()
	if err = r.Update(ctx, &pidScaler); err != nil {
		metrics.CRDUpdateErrors.WithLabelValues(namespacedName.String()).Inc()
	}
	return err
}

// +kubebuilder:rbac:groups=pidscaler.ts,resources=pidscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pidscaler.ts,resources=pidscalers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=pidscaler.ts,resources=pidscalers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PIDScalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("Start reconcile")

	pidScalerCRD, err := r.GetCRD(ctx, req.NamespacedName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("PIDScaler CRD not found. Ignoring since it must be deleted.")
			if err = r.updateStatus(ctx, req.NamespacedName, pidscalerv1.StatusTerminating, "Resource is being deleted"); err != nil {
				return ctrl.Result{}, err
			}
			r.StopWorker(req.NamespacedName)
			return ctrl.Result{}, nil
		}
		r.Log.Error(err, "Failed to get PIDScaler CRD")
		if statusErr := r.updateStatus(ctx, req.NamespacedName, pidscalerv1.StatusUnknown, err.Error()); statusErr != nil {
			return ctrl.Result{}, statusErr
		}
		return ctrl.Result{}, err
	}

	if err = r.updateStatus(ctx, req.NamespacedName, pidscalerv1.StatusInProgress, "Reconciliation in progress"); err != nil {
		return ctrl.Result{}, err
	}

	pidScaler := storage.NewPIDScalerState(&pidScalerCRD)

	existingPIDScaler, exists := r.Storage.Get(req.NamespacedName.String())
	if !exists {
		r.StartWorker(r.OperatorContext, req.NamespacedName, pidScaler)
	} else {
		r.UpdateWorker(existingPIDScaler, pidScaler)
	}
	if pidScalerCRD.Spec.Target.DesiredReplicas != nil {
		err = r.ScaleReplicas(ctx, pidScaler.TargetSettings.Namespace, pidScaler.TargetSettings.Deployment, *pidScalerCRD.Spec.Target.DesiredReplicas)
		if err != nil {
			r.Log.Error(err, "Failed to scale replicas")
			if statusErr := r.updateStatus(ctx, req.NamespacedName, pidscalerv1.StatusFailed, "Failed to scale deployment: "+err.Error()); statusErr != nil {
				return ctrl.Result{}, statusErr
			}
		}
	}

	if err = r.updateStatus(ctx, req.NamespacedName, pidscalerv1.StatusDeployed, "Reconciliation completed successfully"); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PIDScalerReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	r.Log = log.Log.WithName("controller").WithName("PIDScaler")
	r.Storage = storage.NewPIDScalerStorage()
	r.wg = &sync.WaitGroup{}
	r.OperatorContext = ctx

	return ctrl.NewControllerManagedBy(mgr).
		For(&pidscalerv1.PIDScaler{}).
		WithEventFilter(predicate.Funcs{
			// Ignore updates that only change the status field
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldObject := e.ObjectOld.(*pidscalerv1.PIDScaler)
				newObject := e.ObjectNew.(*pidscalerv1.PIDScaler)
				return !equality.Semantic.DeepEqual(oldObject.Spec, newObject.Spec)
			},
		}).
		Complete(r)
}
