{{/* vim: set filetype=mustache: */}}
{{/*
Kubernetes container arguments
*/}}
{{- define "audittail.args" -}}
- '-f'
- '/app-audit/audit.log'
{{- end -}}

{{- define "audittail.initargs" -}}
- 'init'
{{ include "audittail.args" .}}
{{- end -}}
