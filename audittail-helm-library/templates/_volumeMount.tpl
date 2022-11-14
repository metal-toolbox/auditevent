{{/* vim: set filetype=mustache: */}}
{{/*
Kubernetes Volume Mounts
*/}}
{{- define "audittail.volumeMount" -}}
- mountPath: {{ template "audittail.mountPath" }}
  name: {{ template "audittail.volumeName" }}
{{- end -}}
