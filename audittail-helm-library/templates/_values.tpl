{{/* vim: set filetype=mustache: */}}
{{/*
audittail log path
*/}}

{{- define "audittail.auditLogPath" -}}
{{- printf "/app-audit/audit.log" -}}
{{- end -}}


{{- define "audittail.image" -}}
{{- printf "ghcr.io/metal-toolbox/audittail:v0.1.7" -}}
{{- end -}}


{{- define "audittail.volumeName" -}}
{{- printf "audit-logs" -}}
{{- end -}}


{{- define "audittail.mountPath" -}}
{{- printf "/app-audit" -}}
{{- end -}}


{{- define "audittail.initContainerName" -}}
{{- printf "init-audi-logs" -}}
{{- end -}}

{{- define "audittail.sidecarContainerName" -}}
{{- printf "audit-logger" -}}
{{- end -}}
