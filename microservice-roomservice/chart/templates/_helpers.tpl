{{/*
Expand the name of the chart.
*/}}
{{- define "microservice-roomservice.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "microservice-roomservice.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "microservice-roomservice.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "microservice-roomservice.labels" -}}
helm.sh/chart: {{ include "microservice-roomservice.chart" . }}
{{ include "microservice-roomservice.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "microservice-roomservice.selectorLabels" -}}
app.kubernetes.io/name: {{ include "microservice-roomservice.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Environment variables
*/}}
{{- define "microservice-roomservice.envVars" -}}
# Server configuration
- name: PORT
  value: "8080"
- name: BASE_PATH
  value: {{ .Values.basePath | default "/" | quote }}

# Database configuration
- name: DB_HOST
  value: {{ .Values.database.host | quote }}
- name: DB_PORT
  value: {{ .Values.database.port | quote }}
- name: DB_USER
  value: {{ .Values.database.username | quote }}
- name: DB_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ .Values.database.passwordSecret.name }}
      key: {{ .Values.database.passwordSecret.key }}
- name: DB_NAME
  value: {{ .Values.database.name | quote }}
- name: DB_SSLMODE
  value: "disable"

# Redis configuration
- name: REDIS_ADDR
  value: {{ .Values.redis.addr | quote }}
- name: REDIS_KITCHEN_REQUESTS_STREAM
  value: {{ .Values.redis.kitchenRequestsStream | quote }}
- name: REDIS_KITCHEN_RESPONSES_STREAM
  value: {{ .Values.redis.kitchenResponsesStream | quote }}
- name: REDIS_CONSUMER_GROUP
  value: {{ .Values.redis.consumerGroup | quote }}
- name: REDIS_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ .Values.redis.passwordSecret.name }}
      key: {{ .Values.redis.passwordSecret.key }}

# Reservation service configuration
- name: RESERVATION_SERVICE_URL
  value: "http://{{ .Values.reservations.service.name }}.{{ .Values.reservations.service.namespace }}.svc.cluster.local/api/reservations"

# Billing service configuration
- name: BILLING_SERVICE_URL
  value: "http://{{ .Values.billing.service.name }}.{{ .Values.billing.service.namespace }}.svc.cluster.local/api/billing"
- name: BILLING_ERROR_RATE
  value: {{ .Values.billing.errorRate | quote }}

# OpenTelemetry configuration
- name: OTEL_ENABLED
  value: {{ .Values.opentelemetry.enabled | quote }}
- name: OTEL_EXPORTER_OTLP_ENDPOINT
  value: {{ .Values.opentelemetry.endpoint | quote }}
- name: OTEL_SERVICE_NAME
  value: {{ .Values.opentelemetry.serviceName | quote }}
- name: OTEL_SERVICE_VERSION
  value: {{ .Values.opentelemetry.serviceVersion | quote }}
- name: OTEL_ENVIRONMENT
  value: {{ .Values.opentelemetry.environment | quote }}
- name: USER
  value: "otel-user"
{{- end }}