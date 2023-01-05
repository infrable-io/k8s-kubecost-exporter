# Deploying k8s-kubecost-exporter Locally via Minikube

k8s-kubecost-exporter works in conjunction with [Kubecost](https://www.kubecost.com).

For deploying Kubecost locally via Minikube, see the following documentation:
* [Deploying Kubecost Locally via Minikube](docs/deploying-kubecost-locally-via-minikube.md)

**Build the Docker image**:

```bash
# To point your shell to minikube's docker-daemon, run:
$ eval $(minikube -p minikube docker-env)
$ docker build -t kubecost-exporter:latest .
```

**Create a namespace for k8s-kubecost-exporter**:

```bash
$ kubectl create namespace kubecost-exporter
```

**Install the Helm chart**:

```bash
$ helm install kubecost-exporter deploy/helm/kubecost-exporter --namespace kubecost-exporter
```

Cost allocation metrics are exposed on port 9090 of the kubecost-exporter Kubernetes Service (`/metrics`).
