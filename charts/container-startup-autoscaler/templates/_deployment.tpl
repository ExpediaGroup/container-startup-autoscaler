{{- define "csa.deployment.replicas" -}}
{{- if .Values.pod.leaderElectionEnabled -}}
2
{{- else -}}
1
{{- end }}
{{- end }}
