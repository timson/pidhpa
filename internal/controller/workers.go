package controller

import (
	"context"
	"errors"
	"github.com/timson/pidhpa-operator/internal/kafka"
	"github.com/timson/pidhpa-operator/internal/metrics"
	"github.com/timson/pidhpa-operator/internal/pid"
	"github.com/timson/pidhpa-operator/internal/storage"
	"github.com/twmb/franz-go/pkg/kadm"
	"math"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

func (r *PIDScalerReconciler) WaitForAllWorkers() {
	r.wg.Wait()
}

func (r *PIDScalerReconciler) StartWorker(ctx context.Context, namespacedName client.ObjectKey, pidScaler *storage.PIDScalerState) {
	r.Storage.AddOrUpdate(namespacedName.String(), pidScaler)
	r.wg.Add(1)
	go r.Worker(ctx, namespacedName, pidScaler)
}

func updateMetrics(nsName string, lag float64, output float64, ps *storage.PIDScalerState) {
	metrics.KafkaLag.WithLabelValues(nsName, ps.KafkaSettings.Topic, ps.KafkaSettings.Group).Set(lag)
	metrics.ReferenceSignal.WithLabelValues(nsName, ps.KafkaSettings.Topic,
		ps.KafkaSettings.Group).Set(float64(ps.PidSettings.ReferenceSignal))
	metrics.MinOutput.WithLabelValues(nsName, ps.TargetSettings.Namespace,
		ps.TargetSettings.Deployment).Set(float64(ps.TargetSettings.MinReplicas))
	metrics.MaxOutput.WithLabelValues(nsName, ps.TargetSettings.Namespace,
		ps.TargetSettings.Deployment).Set(float64(ps.TargetSettings.MaxReplicas))
	metrics.PidKp.WithLabelValues(nsName).Set(ps.PidSettings.GetKp())
	metrics.PidKi.WithLabelValues(nsName).Set(ps.PidSettings.GetKi())
	metrics.PidKd.WithLabelValues(nsName).Set(ps.PidSettings.GetKd())
	metrics.PidOutput.WithLabelValues(nsName, ps.TargetSettings.Namespace,
		ps.TargetSettings.Deployment).Set(output)
}

func (r *PIDScalerReconciler) Worker(ctx context.Context, namespacedName client.ObjectKey, initialPIDScaler *storage.PIDScalerState) {
	var lastScale time.Time
	var pidScaler *storage.PIDScalerState
	var kafkaAdminClient *kadm.Client
	var pidController *pid.PID
	var err error
	pidScaler = initialPIDScaler

	r.Log.Info("Start worker", "name", namespacedName.String())
	defer r.wg.Done()

	for {
		select {
		case <-ctx.Done():
			r.Log.Info("Context done, exit from worker", "name", namespacedName.String())
			return

		case changes, ok := <-pidScaler.ControlCh:
			if !ok {
				r.Log.Info("Control channel closed", "name", namespacedName.String())
				return
			}
			r.Log.Info("Got update", "name", namespacedName.String(), "mask", changes)
			pidScaler, ok = r.Storage.Get(namespacedName.String())
			if !ok {
				r.Log.Error(errors.New("PIDScaler not found"), "Weird case, pidScaler not found after update event received", "name", namespacedName.String())
				return
			}
			// check if we need to reset PID controller and Kafka client
			if (changes&storage.PidSettingsMask != 0 || changes&storage.TargetSettingsMask != 0) && pidController != nil {
				r.Log.Info("Updating PID controller", "name", namespacedName.String(), "Kp", pidScaler.PidSettings.GetKp(), "Ki",
					pidScaler.PidSettings.GetKi(), "Kd", pidScaler.PidSettings.GetKd(), "minReplicas", pidScaler.TargetSettings.MinReplicas,
					"maxReplicas", pidScaler.TargetSettings.MaxReplicas)
				pidController.UpdateConfig(
					pidScaler.PidSettings.GetKp(), pidScaler.PidSettings.GetKi(), pidScaler.PidSettings.GetKd(),
					float64(pidScaler.TargetSettings.MinReplicas), float64(pidScaler.TargetSettings.MaxReplicas), true)
			}
			if changes&storage.KafkaSettingsMask != 0 {
				r.Log.Info("Updating Kafka client", "name", namespacedName.String(), "brokers", pidScaler.KafkaSettings.Brokers, "topic", pidScaler.KafkaSettings.Topic,
					"group", pidScaler.KafkaSettings.Group)
				kafkaAdminClient = nil
			}
		default:
			if pidController == nil {
				pidController = pid.NewPID(
					pidScaler.PidSettings.GetKp(), pidScaler.PidSettings.GetKi(), pidScaler.PidSettings.GetKd(),
					float64(pidScaler.TargetSettings.MinReplicas), float64(pidScaler.TargetSettings.MaxReplicas), true)
			}

			if kafkaAdminClient == nil {
				kafkaAdminClient, err = kafka.NewKafkaClient(
					pidScaler.KafkaSettings.Brokers,
					pidScaler.KafkaSettings.UseSASL,
					pidScaler.KafkaSettings.SASLMechanism,
					pidScaler.KafkaSettings.Username,
					pidScaler.KafkaSettings.Password,
				)
				if err != nil {
					r.Log.Error(err, "Failed to create Kafka client")
					time.Sleep(time.Duration(pidScaler.Interval) * time.Second)
					continue
				}
			}

			lag, err := kafka.GetKafkaLag(ctx, kafkaAdminClient, pidScaler.KafkaSettings.Group, pidScaler.KafkaSettings.Topic)
			if err != nil {
				if !errors.Is(err, kafka.ErrConsumerGroupNotStable) {
					r.Log.Error(err, "Failed to read Kafka lag", "name", namespacedName.String())
				}
			} else {
				now := time.Now()
				output := pidController.Update(float64(pidScaler.PidSettings.ReferenceSignal), float64(lag), now)
				// update metrics
				updateMetrics(namespacedName.String(), float64(lag), output, pidScaler)

				if now.Sub(lastScale) > (time.Duration(pidScaler.CooldownTimeout) * time.Second) {
					lastScale = now
					roundedOutput := math.Round(output)
					replicas := int32(roundedOutput)
					metrics.Replicas.WithLabelValues(namespacedName.String(), pidScaler.TargetSettings.Namespace,
						pidScaler.TargetSettings.Deployment).Set(roundedOutput)
					pidScalerCRD, err := r.GetCRD(ctx, namespacedName)
					if err != nil {
						r.Log.Error(err, "Failed to get PIDScaler")
					} else {
						dep, found := r.GetDeployment(ctx, pidScaler.TargetSettings.Namespace, pidScaler.TargetSettings.Deployment)
						if pidScalerCRD.Spec.Target.DesiredReplicas != &replicas || (found == true && dep.Spec.Replicas != nil && *dep.Spec.Replicas != replicas) {
							err = r.updateDesiredReplicas(ctx, namespacedName, replicas)
							if err != nil {
								r.Log.Error(err, "Failed to update PIDScaler desired replicas", "name", namespacedName.String())
							}
						}
					}
				}
			}
			time.Sleep(time.Duration(pidScaler.Interval) * time.Second)
		}
	}
}

func (r *PIDScalerReconciler) StopWorker(namespacedName client.ObjectKey) {
	r.Storage.Delete(namespacedName.String())
}

func (r *PIDScalerReconciler) UpdateWorker(existingPIDScaler *storage.PIDScalerState, pidScaler *storage.PIDScalerState) {
	differenceMask := existingPIDScaler.GetDifferenceMask(pidScaler)
	if differenceMask != 0 {
		existingPIDScaler.ControlCh <- differenceMask // trigger update
	}
}
