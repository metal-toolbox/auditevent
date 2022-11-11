{{/* vim: set filetype=mustache: */}}
{{/*
Kubernetes Init Container
*/}}
{{- define "audittail.sidecarContainer" -}}
  - image: ghcr.io/metal-toolbox/audittail:v0.1.7
  name: audit-logger
  args:
    - '-f'
    - '/app-audit/audit.log'
  volumeMounts:
    - mountPath: /app-audit
      name: audit-logs
      readonly: true
{{- end -}}
