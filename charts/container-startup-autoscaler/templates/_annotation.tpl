{{ define "csa.annotation.deployment" }}
{{- end }}

{{ define "csa.annotation.pod" }}
{{- if .Values.pod.extraAnnotations }}
annotations: {{- toYaml .Values.pod.extraAnnotations | nindent 2 }}
{{- end }}
{{- end }}

{{ define "csa.annotation.serviceaccount" }}
{{- end}}

{{ define "csa.annotation.clusterrole" }}
{{- end }}

{{ define "csa.annotation.clusterrolebinding" }}
{{- end }}
