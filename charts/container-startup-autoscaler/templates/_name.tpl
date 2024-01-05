{{- define "csa.name.app" -}}
{{ .Chart.Name }}
{{- end }}

{{- define "csa.name.release" -}}
{{ .Release.Name }}
{{- end }}

{{- define "csa.name.namespace" -}}
{{ .Release.Namespace }}
{{- end }}
