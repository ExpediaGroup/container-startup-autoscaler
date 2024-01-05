{{- define "csa.pod.terminationGracePeriodSeconds" -}}
{{- if .Values.csa.gracefulShutdownTimeoutSecs -}}
{{ add .Values.csa.gracefulShutdownTimeoutSecs 5 }}
{{- else -}}
{{- /*
Must be graceful-shutdown-timeout-secs default + 5s (i.e. some margin)
*/ -}}
15
{{- end }}
{{- end }}
