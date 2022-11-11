{{/* vim: set filetype=mustache: */}}
{{/*
Kubernetes Volume Mounts
*/}}
{{- define "audittail.volumeMount" -}}
- mountPath: /app-audit
  name: audit-logs
{{- end -}}
