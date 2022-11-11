{{/* vim: set filetype=mustache: */}}
{{/*
Kubernetes Volumes
*/}}
{{- define "audittail.volume" -}}
- name: audit-logs
  emptyDir: {}
{{- end -}}
