{{ define "csa.label.deployment" }}
labels: {{- include "csa.label.core" . | nindent 2 }}
{{- end }}

{{ define "csa.label.pod" }}
labels: {{- include "csa.label.core" . | nindent 2 }}
{{- if .Values.pod.extraLabels }}
{{ toYaml .Values.pod.extraLabels | indent 2 }}
{{- end }}
{{- end }}

{{ define "csa.label.serviceaccount" }}
labels: {{- include "csa.label.core" . | nindent 2 }}
{{- end}}

{{ define "csa.label.clusterrole" }}
labels: {{- include "csa.label.core" . | nindent 2 }}
{{- end }}

{{ define "csa.label.clusterrolebinding" }}
labels: {{- include "csa.label.core" . | nindent 2 }}
{{- end }}

{{- define "csa.label.core" -}}
helm.sh/chart: "{{ .Chart.Name }}-{{ .Chart.Version }}"
app.kubernetes.io/managed-by: "{{ .Release.Service }}"
{{ include "csa.label.kubeName" . }}
{{ include "csa.label.kubeInstance" . }}
app.kubernetes.io/version: "{{ include "csa.container.tag" . }}"
{{- end }}

{{- define "csa.label.kubeName" -}}
app.kubernetes.io/name: "{{ include "csa.name.app" . }}"
{{- end }}

{{- define "csa.label.kubeInstance" -}}
app.kubernetes.io/instance: "{{ include "csa.name.release" . }}"
{{- end }}
