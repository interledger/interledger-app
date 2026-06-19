{{- define "mock-services.mockpti.configMap" -}}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "common.fullname" . }}-{{ .Values.mockpti.configMaps.config.name }}
  namespace: {{ .Release.Namespace }}
  {{- include "common.metadata" (list .) | nindent 2 }}
data:
  config.yaml: |
    {{- .Values.mockpti.config | toYaml | nindent 4 }}
{{- end }}

{{- define "mock-services.mockgatehub.configMap" -}}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "common.fullname" . }}-{{ .Values.mockgatehub.configMaps.config.name }}
  namespace: {{ .Release.Namespace }}
  {{- include "common.metadata" (list .) | nindent 2 }}
data:
  config.yaml: |
    {{- .Values.mockgatehub.config | toYaml | nindent 4 }}
{{- end }}

{{- define "mock-services.mockxago.configMap" -}}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "common.fullname" . }}-{{ .Values.mockxago.configMaps.config.name }}
  namespace: {{ .Release.Namespace }}
  {{- include "common.metadata" (list .) | nindent 2 }}
data:
  config.yaml: |
    {{- .Values.mockxago.config | toYaml | nindent 4 }}
{{- end }}
