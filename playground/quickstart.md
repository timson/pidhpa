# PID HPA Operator Playground - Quick Start Guide

## Overview
This guide provides step-by-step instructions to set up a Kubernetes (K8s) playground for testing the PID HPA operator. The setup includes:
- A `kind` K8s cluster
- Kafka with a `test` topic containing 24 partitions
- Consumer and producer deployments
- PID HPA operator with our CRD
- VictoriaMetrics monitoring stack with Grafana

## Prerequisites
Ensure you have the following tools installed:
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
- [Kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Helm](https://helm.sh/docs/intro/install/)
- [Python3](https://www.python.org/downloads/) (for ingestion script)

## Steps to Set Up the Playground

### 1. Create Kind Kubernetes Cluster
```sh
kind create cluster --name pidhpa --config kind-cluster.yaml
```

### 2. Install Kafka
```sh
kubectl create namespace kafka
kubectl create -f 'https://strimzi.io/install/latest?namespace=kafka' -n kafka
kubectl apply -f kafka-topic.yaml -n kafka
```

wait until operator is ready:
```sh
kubectl apply -f https://strimzi.io/examples/latest/kafka/kraft/kafka-single-node.yaml -n kafka
```


#### 3. Update Kafka Topic to 24 Partitions
```sh
kubectl exec pod/my-cluster-dual-role-0 -n kafka -it -- \
    /opt/kafka/bin/kafka-topics.sh --alter --topic test --partitions 24 --bootstrap-server 0.0.0.0:9092
```

#### 4. Verify Kafka Topic Configuration
```sh
kubectl exec pod/my-cluster-dual-role-0 -n kafka -it -- \
    /opt/kafka/bin/kafka-topics.sh --describe --topic test --bootstrap-server 0.0.0.0:9092
```

### 5. Deploy Kafka Producer and Consumer
```sh
kubectl apply -f producer-deployment.yaml -n default
kubectl apply -f consumer-deployment.yaml -n default
```
> **Note:** One consumer pod processes approximately 10 messages per second.

### 6. Install PID HPA Operator

From the project root directory:

```sh
helm install pidhpa-operator ./helm --namespace default
```

### 7. Apply Custom Resource Definition (CRD)
```sh
kubectl apply -f pidscaler.yaml -n default
```

### 8. Install VictoriaMetrics Monitoring Stack

Install VictoriaMetrics monitoring stack using Helm:

```sh
helm repo add vm https://victoriametrics.github.io/helm-charts/
helm repo update
helm install vmks vm/victoria-metrics-k8s-stack -f vmstack-values.yaml -n vm --create-namespace
```

### 9. Install Static Scraper for VictoriaMetrics
```sh
kubectl apply -f static-scrape.yaml -n vm
```

## Data Ingestion and Metrics Access

### 10. Start Kafka Data Ingestion
```sh
kubectl port-forward deployment.apps/kafka-producer 5000:500
python3 ingest.py
```
> **Note:** The ingestion script sends approximately 50 messages per second.

### 11. Access Grafana for Monitoring

#### Get Grafana Admin Password
```sh
kubectl get secret --namespace vm vmks-grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo
```

#### Port-Forward Grafana
```sh
kubectl port-forward service/vmks-grafana 3000:80 -n vm
```

Now, open [http://localhost:3000](http://localhost:3000) in your browser and log in to Grafana using the retrieved password.

### 12. Import the Dashboard
You can import the `pid-hpa-test-dashboard.json` file into Grafana for monitoring.

---

## Cleanup
To delete the playground environment:
```sh
kind delete cluster --name pidhpa
```

---
This guide ensures a quick and reproducible setup for testing the PID HPA operator with Kafka and VictoriaMetrics in a local Kubernetes environment.
