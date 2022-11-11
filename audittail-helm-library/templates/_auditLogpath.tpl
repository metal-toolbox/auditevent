{{/* vim: set filetype=mustache: */}}
{{/*
audittail log path
*/}}

{{- define "audittail.auditLogPath" -}}
{{- printf "/app-audit/audit.log" -}}
{{- end -}}
