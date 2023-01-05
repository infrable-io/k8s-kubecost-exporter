# K8s Kubecost Exporter

[![MIT License](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/infrable-io/k8s-opencost-exporter/blob/master/LICENSE)
[![Maintained by Infrable](https://img.shields.io/badge/Maintained%20by-Infrable-000000)](https://infrable.io)
[![GitHub Actions - Go](https://github.com/infrable-io/k8s-opencost/actions/workflows/go.yml/badge.svg)](https://github.com/infrable-io/k8s-opencost/actions/workflows/go.yml)
[![GitHub Actions - Helm](https://github.com/infrable-io/k8s-opencost/actions/workflows/helm.yml/badge.svg)](https://github.com/infrable-io/k8s-opencost/actions/workflows/helm.yml)

k8s-kubecost-exporter is a Kubernetes application that exposes cost allocation metrics retrieved from Kubecost.

Cost allocation metrics are retrieved from the [Kubecost Allocation API](https://docs.kubecost.com/apis/apis/allocation) and made available via a metrics HTTP endpoint (`/metrics`). Applications that can extract custom metrics from OpenMetrics endpoints (Prometheus, Datadog, New Relic etc.) can be configured to scrape this endpoint.

<p align="center">
  <img src="assets/architecture.svg">
</p>

For an architecture overview of Kubecost, see the following documentation:
* [Kubecost Core Architecture Overview](https://docs.kubecost.com/architecture/architecture)

## Development

k8s-kubecost-exporter works in conjunction with [Kubecost](https://www.kubecost.com).

For deploying k8s-kubecost-exporter locally via Minikube, see the following documentation:
* [Deploying k8s-kubecost-exporter Locally via Minikube](docs/deploying-k8s-kubecost-exporter-locally-via-minikube.md)

For deploying Kubecost locally via Minikube, see the following documentation:
* [Deploying Kubecost Locally via Minikube](docs/deploying-kubecost-locally-via-minikube.md)

Optionally, you can install New Relic's infrastructure monitoring agent via Helm using the following documentation:
* [Install the Kubernetes integration using Helm](https://docs.newrelic.com/docs/kubernetes-pixie/kubernetes-integration/installation/install-kubernetes-integration-using-helm)

See the [Standard Go Project Layout](https://github.com/golang-standards/project-layout) GitHub repository for a canonical layout of Go applications upon which this repository is based.

The official container image (`Dockerfile`) contains only the binary executable, so for debugging purposes, a Dockerfile (`Dockerfile.debug`), is provided based on Alpine Linux.

### Unit Testing

Unit testing is facilitate by [gomock](https://github.com/golang/mock), a mocking framework for Go.

Mocked interfaces are generated using the following command:

```bash
$ mockgen -package main -source client.go -destination client_mock.go
```

## Testing

To test the Go code, run the following:

```bash
$ go test
```

To test the Helm chart, run the following:

```bash
$ helm lint deploy/helm/kubecost-exporter
$ helm package deploy/helm/kubecost-exporter
$ kubectl create namespace kubecost-exporter
$ helm install kubecost-exporter kubecost-exporter-<version>.tgz --namespace kubecost-exporter \
  --set image.registry="" \
  --set image.repository="kubecost-exporter" \
  --set image.tag="latest" \
  --wait
$ helm test kubecost-exporter --namespace kubecost-exporter
```

**NOTE**: The container image can be built locally and made available to Minikube by running the following:

```bash
$ eval $(minikube -p minikube docker-env)
$ docker build -t kubecost-exporter:latest .
```
