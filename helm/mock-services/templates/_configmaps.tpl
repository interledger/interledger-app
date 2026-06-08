{{- define "mock-services.mockpti.configMap" -}}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "common.fullname" . }}-{{ .Values.mockpti.configMaps.config.name }}
  namespace: {{ .Release.Namespace }}
  {{- include "common.metadata" (list .) | nindent 2 }}
data:
  LOG_LEVEL: {{ .Values.mockpti.config.log_level | quote }}
  MOCKPTI_REDIS_URL: {{ .Values.mockpti.config.redis_url | quote }}
  MOCKPTI_REDIS_DB: {{ .Values.mockpti.config.redis_db | quote }}
  MOCKPTI_CLIENT_ID: {{ .Values.mockpti.config.client_id | quote }}
  MOCKPTI_WEBHOOK_URL: {{ .Values.mockpti.config.webhook_url | quote }}
  WEBHOOK_MIN_DELAY_SEC: {{ .Values.mockpti.config.webhook_min_delay_sec | quote }}
{{- end }}

{{- define "mock-services.mockgatehub.configMap" -}}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "common.fullname" . }}-{{ .Values.mockgatehub.configMaps.config.name }}
  namespace: {{ .Release.Namespace }}
  {{- include "common.metadata" (list .) | nindent 2 }}
data:
  LOG_LEVEL: {{ .Values.mockgatehub.config.log_level | quote }}
  MOCKGATEHUB_REDIS_URL: {{ .Values.mockgatehub.config.redis_url | quote }}
  MOCKGATEHUB_REDIS_DB: {{ .Values.mockgatehub.config.redis_db | quote }}
  WEBHOOK_URL: {{ .Values.mockgatehub.config.webhook_url | quote }}
  WEBHOOK_MIN_DELAY_SEC: {{ .Values.mockgatehub.config.webhook_min_delay_sec | quote }}
  MOCKGATEHUB_ENFORCE_AUTHENTICATION: {{ .Values.mockgatehub.config.enforce_authentication | quote }}
  DEFAULT_ORGANIZATION_ID: {{ .Values.mockgatehub.config.default_organization_id | quote }}
  MOCKGATEHUB_PUBLIC_BASE_URL: {{ .Values.mockgatehub.config.public_base_url | quote }}
{{- end }}

{{- define "mock-services.mockxago.configMap" -}}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "common.fullname" . }}-{{ .Values.mockxago.configMaps.config.name }}
  namespace: {{ .Release.Namespace }}
  {{- include "common.metadata" (list .) | nindent 2 }}
data:
  LOG_LEVEL: {{ .Values.mockxago.config.log_level | quote }}
  MOCKXAGO_REDIS_URL: {{ .Values.mockxago.config.redis_url | quote }}
  MOCKXAGO_REDIS_DB: {{ .Values.mockxago.config.redis_db | quote }}
  WEBHOOK_URL: {{ .Values.mockxago.config.webhook_url | quote }}
  WEBHOOK_MIN_DELAY_SEC: {{ .Values.mockxago.config.webhook_min_delay_sec | quote }}
  MOCKXAGO_ENFORCE_AUTHENTICATION: {{ .Values.mockxago.config.enforce_authentication | quote }}
  XAGO_MOCK_TEST_MODE: {{ .Values.mockxago.config.test_mode | quote }}
  PERSONA_WEBHOOK_URL: {{ .Values.mockxago.config.persona_webhook_url | quote }}
{{- end }}
