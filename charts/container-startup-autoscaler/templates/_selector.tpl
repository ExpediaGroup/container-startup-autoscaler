{{ define "csa.selector.deployment" }}
matchLabels:
  {{ include "csa.label.kubeName" . }}
  {{ include "csa.label.kubeInstance" . }}
{{- end }}
