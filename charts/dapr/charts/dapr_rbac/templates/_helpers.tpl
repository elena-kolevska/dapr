{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "dapr_rbac.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "dapr_rbac.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "dapr_rbac.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Add common rules for sentry
*/}}
{{- define "dapr_rbac.sentry_rules" }}
  - apiGroups: ["dapr.io"]
    resources: ["configurations"]
    verbs: ["list", "get", "watch"]
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["list", "get", "watch"]
{{- end}}

{{/*
Add common rules for operator
*/}}
{{- define "dapr_rbac.operator_rules" }}
  - apiGroups: ["apps"]
    resources: ["deployments", "deployments/finalizers"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["apps"]
    resources: ["deployments/finalizers"]
    verbs: ["update"]
  - apiGroups: ["apps"]
    resources: ["statefulsets", "statefulsets/finalizers"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["apps"]
    resources: ["statefulsets/finalizers"]
    verbs: ["update"]
  - apiGroups: [""]
    resources: ["pods"]
{{-   if .Values.global.operator.watchdogCanPatchPodLabels }}
    verbs: ["get", "list", "delete", "watch", "patch"]
{{-   else }}
    verbs: ["get", "list", "delete", "watch"]
{{-   end }}
  - apiGroups: [""]
    resources: ["services","services/finalizers"]
    verbs: ["get", "list", "watch", "update", "create"]
  - apiGroups: [""]
    resources: ["services"]
    verbs: ["delete"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["dapr.io"]
    resources: ["components", "configurations", "subscriptions", "resiliencies", "httpendpoints"]
    verbs: [ "get", "list", "watch"]
{{-   if .Values.global.argoRolloutServiceReconciler.enabled }}
  - apiGroups: ["argoproj.io"]
    resources: ["rollouts"]
    verbs: ["get", "list", "watch", "delete"]
  - apiGroups: ["argoproj.io"]
    resources: ["rollouts/finalizers"]
    verbs: ["update"]
{{-   end }}
{{- end }}