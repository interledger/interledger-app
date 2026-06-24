{{- define "interledger-app.frontend.configMap" -}}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "common.fullname" . }}-{{ .Values.frontend.configMaps.server.name }}
  namespace: {{ .Release.Namespace }}
  {{- include "common.metadata" (list .) | nindent 2 }}
data:
  OTEL_EXPORTER_OTLP_ENDPOINT: {{ .Values.frontend.config.otel.endpoint | quote }}
  OTEL_SERVICE_NAME: {{ .Values.frontend.config.otel.service_name | quote }}
  KRATOS_URL: {{ .Values.frontend.config.kratos_url | quote }}
  PAYMENT_POINTER_BASE: {{ .Values.frontend.config.payment_pointer_base | quote }}
  RAFIKI_AUTH_ENDPOINT: {{ .Values.frontend.config.rafiki.auth.endpoint | quote }}
  BACKEND_GRPC_URL: {{ default (printf "http://%s-backend-service-grpc:8443" (include "common.fullname" .)) .Values.frontend.config.backend.grpc.url | quote }}
  FYNBOS_ENV: {{ .Values.common.environment.behaviour | quote }}
  LOG_LEVEL: {{ .Values.frontend.config.log_level | quote }}
  LOG_PRETTY: {{ .Values.frontend.config.log_pretty | toString | quote }}
  PTI_CLIENT_ID: {{ .Values.frontend.config.pti.client_id | quote }}
  PTI_SDK_URL: {{ .Values.frontend.config.pti.sdk_url | quote }}
  PTI_FORMS_URL: {{ .Values.frontend.config.pti.forms_url | quote }}
  BACKEND_HTTP_URL: {{ default (printf "http://%s-backend-service-http:8080" (include "common.fullname" .)) .Values.frontend.config.backend.http.url | quote }}
  TARGET_HOST: {{ .Values.frontend.config.target_host | quote }}
  SUPPORT_EMAIL: {{ .Values.frontend.config.support_email | quote }}
  PUBLIC_OP_AUTH_HOST: {{ .Values.frontend.config.public_op_auth_host | quote }}
  PERSONA_SDK_URL: {{ .Values.frontend.config.persona_sdk_url | quote }}
  MOCKXAGO_ENDPOINT: {{ .Values.frontend.config.mockxago_endpoint | quote }}
  SENTRY_ENV_LABEL: {{ .Values.common.environment.label | quote }}
{{- end }}

{{- define "interledger-app.admin.configMap" -}}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "common.fullname" . }}-{{ .Values.admin.configMaps.config.name }}
  namespace: {{ .Release.Namespace }}
  {{- include "common.metadata" (list .) | nindent 2 }}
data:
  BACKEND_GRPC_URL: {{ default (printf "http://%s-backend-service-grpc:8448" (include "common.fullname" .)) .Values.admin.config.backend.grpcUrl | quote }}
  KRATOS_ADMIN_URL: {{ .Values.admin.config.kratos.adminUrl | quote }}
  PAYMENT_POINTER_BASE: {{ .Values.admin.config.payment_pointer_base | quote }}
{{- end }}
