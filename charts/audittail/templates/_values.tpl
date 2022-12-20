{{/* vim: set filetype=mustache: */}}
{{/*
audittail log path
*/}}

{{- define "audittail.auditLogPath" -}}
{{- printf "/app-audit/audit.log" -}}
{{- end -}}


{{- define "audittail.image" -}}
{{- if .Values.auditailImage -}}
{{- .Values.auditailImage -}}
{{- else -}}
{{- printf "ghcr.io/metal-toolbox/audittail:v0.5.1" -}}
{{- end -}}
{{- end -}}


{{- define "audittail.volumeName" -}}
{{- printf "audit-logs" -}}
{{- end -}}


{{- define "audittail.mountPath" -}}
{{- printf "/app-audit" -}}
{{- end -}}


{{- define "audittail.initContainerName" -}}
{{- printf "init-audit-logs" -}}
{{- end -}}

{{- define "audittail.sidecarContainerName" -}}
{{- printf "audit-logger" -}}
{{- end -}}
