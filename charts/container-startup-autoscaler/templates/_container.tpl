{{- define "csa.container.tag" -}}
{{- if .Values.container.tag }}
{{- .Values.container.tag}}
{{- else }}
{{- .Chart.AppVersion }}
{{- end }}
{{- end }}

{{- define "csa.container.imageTag" -}}
{{ .Values.container.image }}:{{ include "csa.container.tag" . }}
{{- end }}

{{ define "csa.container.args" }}
args:
  - --leader-election-enabled
  - "{{ .Values.pod.leaderElectionEnabled }}"
  - --leader-election-resource-namespace
  - "{{ include "csa.name.namespace" . }}"
  {{- if .Values.csa.cacheSyncPeriodMins }}
  - --cache-sync-period-mins
  - "{{ .Values.csa.cacheSyncPeriodMins }}"
  {{- end }}
  {{- if .Values.csa.gracefulShutdownTimeoutSecs }}
  - --graceful-shutdown-timeout-secs
  - "{{ .Values.csa.gracefulShutdownTimeoutSecs }}"
  {{- end }}
  {{- if .Values.csa.requeueDurationSecs }}
  - --requeue-duration-secs
  - "{{ .Values.csa.requeueDurationSecs }}"
  {{- end }}
  {{- if .Values.csa.maxConcurrentReconciles }}
  - --max-concurrent-reconciles
  - "{{ .Values.csa.maxConcurrentReconciles }}"
  {{- end }}
  {{- if .Values.csa.standardRetryAttempts }}
  - --standard-retry-attempts
  - "{{ .Values.csa.standardRetryAttempts }}"
  {{- end }}
  {{- if .Values.csa.standardRetryDelaySecs }}
  - --standard-retry-delay-secs
  - "{{ .Values.csa.standardRetryDelaySecs }}"
  {{- end }}
  {{- if .Values.csa.scaleWhenUnknownResources }}
  - --scale-when-unknown-resources
  - "{{ .Values.csa.scaleWhenUnknownResources }}"
  {{- end }}
  {{- if .Values.csa.logV }}
  - --log-v
  - "{{ .Values.csa.logV }}"
  {{- end }}
  {{- if .Values.csa.logAddCaller }}
  - --log-add-caller
  - "{{ .Values.csa.logAddCaller }}"
  {{- end }}
{{- end }}
