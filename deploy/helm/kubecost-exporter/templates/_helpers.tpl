{{/*
Expand the name of the chart.
*/}}
{{- define "kubecost-exporter.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name. We truncate at 63 characters because
some Kubernetes name fields are limited to this amount by the DNS naming spec.
If the release name contains the chart name, it will be used as a full name.
*/}}
{{- define "kubecost-exporter.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "kubecost-exporter.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kubecost-exporter.labels" -}}
helm.sh/chart: {{ include "kubecost-exporter.chart" . }}
{{ include "kubecost-exporter.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kubecost-exporter.selectorLabels" -}}
name: {{ include "kubecost-exporter.name" . }}
app.kubernetes.io/name: {{ include "kubecost-exporter.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Container image
*/}}
{{- define "kubecost-exporter.image" -}}
{{- with .Values.image }}
{{- if .registry }}
{{- printf "%s/%s:%s" .registry .repository .tag}}
{{- else }}
{{- printf "%s:%s" .repository .tag}}
{{- end }}
{{- end }}
{{- end }}
