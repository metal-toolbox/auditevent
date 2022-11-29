{{/* vim: set filetype=mustache: */}}
{{/*
Kubernetes Container Security Context
*/}}
{{- define "audittail.securityContext" -}}
allowPrivilegeEscalation: false
runAsNonRoot: true
{{- end -}}
