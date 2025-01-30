package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	KafkaLag = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kafka_lag",
			Help: "Kafka lag per namespaced name",
		},
		[]string{"namespaced_name", "topic", "group"},
	)
	ReferenceSignal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "reference_signal",
			Help: "Reference signal per namespaced name",
		},
		[]string{"namespaced_name", "topic", "group"},
	)
	PidKp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pid_kp",
			Help: "Kp value of PID controller per namespaced name",
		},
		[]string{"namespaced_name"},
	)
	PidKi = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pid_ki",
			Help: "Ki value of PID controller per namespaced name",
		},
		[]string{"namespaced_name"},
	)
	PidKd = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pid_kd",
			Help: "Kd value of PID controller per namespaced name",
		},
		[]string{"namespaced_name"},
	)
	MinOutput = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pid_min_output",
			Help: "Minimum output of PID controller per namespaced name",
		},
		[]string{"namespaced_name", "namespace", "deployment"},
	)
	MaxOutput = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pid_max_output",
			Help: "Maximum output of PID controller per namespaced name",
		},
		[]string{"namespaced_name", "namespace", "deployment"},
	)
	PidOutput = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pid_actual_output",
			Help: "PID actual output of PID controller per namespaced name",
		},
		[]string{"namespaced_name", "namespace", "deployment"},
	)
	Replicas = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "replicas",
			Help: "Desired replicas per namespaced name",
		},
		[]string{"namespaced_name", "namespace", "deployment"},
	)
	CRDUpdateErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "crd_update_errors",
			Help: "Number of errors updating CRD",
		},
		[]string{"namespaced_name"},
	)
	CRDFetchErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "crd_fetch_errors",
			Help: "Number of errors fetching CRD",
		},
		[]string{"namespaced_name"},
	)
)
