{{/* vim: set filetype=mustache: */}}
{{/*
Kubernetes Init Container
*/}}
{{- define "audittail.sidecarContainer" -}}
  - image: {{ include "audittail.image" .}}
  name: {{ template "audittail.sidecarContainerName" }}
  args: {{ include "audittail.args" .| nindent 4}}
  securityContext: {{ include "audittail.securityContext" . | nindent 4}}
  volumeMounts: {{ include "audittail.volumeMount" . | nindent 4}}
      readOnly: true
{{- end -}}
