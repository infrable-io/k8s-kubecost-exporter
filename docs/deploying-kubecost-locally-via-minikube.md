# Deploying Kubecost Locally via Minikube

**Install Minikube via Homebrew**:

```bash
$ brew install minikube
```

**Start Minikube**:

```bash
$ minikube start
```

**NOTE**: Ensure Docker is running.

**Create a namespace for Kubecost**:

```bash
$ kubectl create namespace kubecost
```

**Add the Kubecost Helm chart repository**:

```bash
$ helm repo add kubecost https://kubecost.github.io/cost-analyzer
```

**Install the Helm chart**:
```bash
helm upgrade --install kubecost kubecost/cost-analyzer --namespace kubecost --create-namespace
```

When Pods are ready, you can enable port forwarding with the following command:
```bash
$ kubectl port-forward --namespace kubecost deployment/kubecost-cost-analyzer 9090
```

Next, navigate to http://localhost:9090 in a web browser.

If you are having issues, view the [Troubleshooting Guide](http://docs.kubecost.com/troubleshoot-install).
