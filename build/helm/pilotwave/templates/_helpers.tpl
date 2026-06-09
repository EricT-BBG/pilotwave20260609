{{/*
Expand the chart name.
*/}}
{{- define "pilotwave.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a release-scoped name. The default pilotwave release keeps the historic
resource names while additional releases avoid cluster-scoped name collisions.
*/}}
{{- define "pilotwave.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := include "pilotwave.name" . -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Labels shared by namespaced workload objects.
*/}}
{{- define "pilotwave.labels" -}}
app: "pilotwave"
chart: {{ .Chart.Name }}
release: {{ .Release.Name }}
heritage: {{ .Release.Service }}
app.kubernetes.io/name: {{ include "pilotwave.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" }}
{{- end -}}

{{/*
Selector labels intentionally keep the historic selectors for upgrade safety.
*/}}
{{- define "pilotwave.selectorLabels" -}}
app: "pilotwave"
chart: {{ .Chart.Name }}
release: {{ .Release.Name }}
heritage: {{ .Release.Service }}
tier: frontend
{{- end -}}

{{- define "pilotwave.serviceAccountName" -}}
{{- default (include "pilotwave.fullname" .) .Values.serviceAccount.name -}}
{{- end -}}

{{- define "pilotwave.configMapName" -}}
{{- printf "%s-config" (include "pilotwave.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "pilotwave.authSecretName" -}}
{{- default (printf "%s-auth" (include "pilotwave.fullname" .)) .Values.auth.secret.name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "pilotwave.authSecretKey" -}}
{{- default "auth-secret" .Values.auth.secret.key -}}
{{- end -}}

{{- define "pilotwave.imagePullSecretName" -}}
{{- $name := coalesce .Values.imagePullSecret.name .Values.imageCredentials.name -}}
{{- default (printf "%s-registry" (include "pilotwave.fullname" .)) $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "pilotwave.imagePullSecrets" -}}
{{- with .Values.imagePullSecret.existingSecret }}
- name: {{ . }}
{{- end }}
{{- if or .Values.imagePullSecret.create .Values.imageCredentials.create }}
- name: {{ include "pilotwave.imagePullSecretName" . }}
{{- end }}
{{- with .Values.imagePullSecrets }}
{{ toYaml . }}
{{- end }}
{{- end -}}

{{- define "pilotwave.pvcName" -}}
{{- include "pilotwave.fullname" . -}}
{{- end -}}
