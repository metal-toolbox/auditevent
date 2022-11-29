{{/* vim: set filetype=mustache: */}}
{{/*
Kubernetes Volumes
*/}}
{{- define "audittail.volume" -}}
- name: {{ template "audittail.volumeName" }}
  emptyDir: {}
{{- end -}}
