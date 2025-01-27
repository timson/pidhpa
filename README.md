# pidhpa
A Kubernetes operator that uses a PID controller to manage Horizontal Pod Autoscaling (HPA), offering a new way to handle scaling based on specific metrics.

## Description
pidhpa is a project that experiments with using a PID (Proportional-Integral-Derivative) controller for scaling applications in Kubernetes.
Instead of relying on standard metrics like CPU or memory, it focuses on Kafka consumer lag, which shows how much data is waiting to be processed.

You can set the Kafka topic and consumer group to monitor, and the PID controller will adjust the number of pods to keep the lag at a target level.
Users can customize the PID settings, scaling frequency, cooldown time, and the minimum/maximum number of pods.

### Why Use a PID Controller?
PID controllers are great for making quick and stable adjustments. They can:

- **React to Changes Quickly**: Adjust scaling based on real-time lag.
- **Avoid Overreacting**: Balance short-term changes with long-term trends.
- **Save Resources**: Scale efficiently to match workload needs.

This project explores how PID controllers can improve scaling in Kubernetes, especially for apps with dynamic workloads.

## Getting Started

### Prerequisites
- go version v1.22.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/pidhpa:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```


> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

## Configure Your Own CRD
To set up your own PIDScaler Custom Resource Definition (CRD), you can use the following example configuration. Each field is explained below:

```yaml
apiVersion: pidscaler.ts/v1
kind: PIDScaler
metadata:
  labels:
    app.kubernetes.io/name: pidhpa
    app.kubernetes.io/managed-by: kustomize
  name: pidscaler-sample
spec:
  target:
    deployment: nginx-deployment
    namespace: default
    min_replicas: 1
    max_replicas: 10
  pid:
    kp: "0.0001"
    ki: "0.001"
    kd: "0"
    reference_signal: 10
  kafka:
    brokers:
      - "localhost:50000"
    topic: glogger
    group: group1
    use_sasl: false
  interval: 5
  cooldown_timeout: 30
```

### Field Explanations
#### `target`
- **deployment**: The name of the deployment to scale.
- **namespace**: The namespace where the deployment is located.
- **min_replicas**: Minimum number of pods allowed.
- **max_replicas**: Maximum number of pods allowed.

#### `pid`
- **kp**: Proportional gain. Controls the reaction to the current error.
- **ki**: Integral gain. Corrects past errors by accounting for accumulated lag.
- **kd**: Derivative gain. Reacts to the rate of change of lag.
- **reference_signal**: The desired target lag value to maintain (e.g., 10).

#### `kafka`
- **brokers**: A list of Kafka broker addresses (e.g., "localhost:9092").
- **topic**: The Kafka topic to monitor.
- **group**: The Kafka consumer group to track lag for.
- **use_sasl**: Whether to enable SASL authentication (optional).
- **sasl_mechanism**, **username**, **password**: SASL authentication settings (if enabled).

#### General Settings
- **interval**: The time (in seconds) between scaling checks.
- **cooldown_timeout**: The minimum time (in seconds) between scaling actions.

### Create Instances of Your Solution
You can apply the sample configuration:

```sh
kubectl apply -k config/samples/
```

> **NOTE**: Ensure the samples have valid default values for testing.

### To Uninstall
To delete the CRD instances from the cluster:

```sh
kubectl delete -k config/samples/
```



**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Simulation
The project includes a Python simulation notebook located in the `simulation` folder. This notebook allows you to simulate workloads and test your PID coefficients before applying them to your Kubernetes cluster.

### Features of the Simulation
- Implements a Python version of the PID controller to mimic its behavior in Kubernetes.
- Simulates dynamic ingestion rates and pod scaling.
- Provides visualizations of queue size, pod count, and ingestion rates over time.

### How to Use the Simulation
1. Navigate to the `simulation` folder and open the Jupyter notebook.
2. Adjust the PID coefficients (`kp`, `ki`, `kd`) and other parameters such as `min_out`, `max_out`, and `reference_signal` to suit your needs.
3. Run the simulation to observe how the PID controller handles varying workloads.
4. Analyze the plots generated to understand the impact of your settings.

### Example Workflow
The simulation demonstrates:
- A workload with dynamically changing ingestion rates.
- PID-controlled scaling to maintain a target queue size.
- Visualization of metrics like queue size, pod count, and ingestion rates.

This tool helps fine-tune your PID settings in a controlled environment, making it easier to achieve stable and efficient scaling in production.

### Example Output
The simulation produces the following plots:
- **Queue Size Over Time**: Shows how the queue size fluctuates and how the PID controller reacts to maintain the target.
- **Pod Count Over Time**: Displays the number of pods adjusted by the PID controller.
- **Ingestion Rate Over Time**: Visualizes the changing workload ingestion rates.


## License

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
