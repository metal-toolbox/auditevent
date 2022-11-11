{{/* vim: set filetype=mustache: */}}
{{/*
Kubernetes Init Container
*/}}
{{- define "audittail.initContainer" -}}
  - image: ghcr.io/metal-toolbox/audittail:v0.1.7
  name: init-audi-logs
  args:
    - 'init'
    - '-f'
    - '/app-audit/audit.log'
  volumeMounts:
    - mountPath: /app-audit
      name: audit-logs
{{- end -}}
