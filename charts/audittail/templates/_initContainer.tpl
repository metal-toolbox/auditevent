{{/* vim: set filetype=mustache: */}}
{{/*
Kubernetes Init Container
*/}}
{{- define "audittail.initContainer" -}}
  - image: {{ include "audittail.image" .}}
  name: {{ template "audittail.initContainerName" }}
  args: {{ include "audittail.initargs" . | nindent 4}}
  securityContext: {{ include "audittail.securityContext" . | nindent 4}}
  volumeMounts: {{ include "audittail.volumeMount" . | nindent 4}}
{{- end -}}
