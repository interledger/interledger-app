{{/*
Renders the backend config.yaml ConfigMap by serialising .Values.backend.config verbatim.
The resulting YAML file is mounted into backend pods at /etc/backend/config.yaml.
Secret values should be expressed as configa templates, e.g.:
  {{ secret "backend-secrets" "myKey" }}
which configa resolves against the Kubernetes Secrets API at pod startup.
*/}}
{{- define "interledger-app.backend.configMap" -}}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "common.fullname" . }}-backend-config
  namespace: {{ .Release.Namespace }}
  {{- include "common.metadata" (list .) | nindent 2 }}
data:
  config.yaml: |
    {{- .Values.backend.config | toYaml | nindent 4 }}
{{- end }}

{{/*
Renders the backend migration config.yaml ConfigMap from .Values.backend.migration.config.
Mounted at /etc/backend/config.yaml in the migration init-job.
*/}}
{{- define "interledger-app.backend.migration.configMap" -}}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "common.fullname" . }}-backend-migration-config
  namespace: {{ .Release.Namespace }}
  {{- include "common.metadata" (list .) | nindent 2 }}
  annotations:
    "helm.sh/hook": pre-install,pre-upgrade
    "helm.sh/hook-weight": "-5"
    "helm.sh/hook-delete-policy": before-hook-creation,hook-succeeded
data:
  config.yaml: |
    {{- .Values.backend.migration.config | toYaml | nindent 4 }}
{{- end }}

{{/*
Renders the frontend (protea) config.yaml ConfigMap by serialising
.Values.frontend.config verbatim, the same way the backend ConfigMap does.
The resulting YAML file is mounted into frontend pods at /etc/frontend/config.yaml.
Secret values should be expressed as configa templates, e.g.:
  {{ secret "frontend-secrets" "myKey" }}
which configa resolves against the Kubernetes Secrets API at pod startup.
backend.grpc_url/backend.http_url default to the in-cluster backend Service
DNS names when left empty, mirroring the old per-key ConfigMap's behaviour.
*/}}
{{- define "interledger-app.frontend.configMap" -}}
{{- $cfg := deepCopy .Values.frontend.config -}}
{{- $backend := deepCopy $cfg.backend -}}
{{- $_ := set $backend "grpc_url" (default (printf "http://%s-backend-service-grpc:8443" (include "common.fullname" .)) $cfg.backend.grpc_url) -}}
{{- $_ := set $backend "http_url" (default (printf "http://%s-backend-service-http:8080" (include "common.fullname" .)) $cfg.backend.http_url) -}}
{{- $_ := set $cfg "backend" $backend -}}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "common.fullname" . }}-frontend-config
  namespace: {{ .Release.Namespace }}
  {{- include "common.metadata" (list .) | nindent 2 }}
data:
  config.yaml: |
    {{- $cfg | toYaml | nindent 4 }}
{{- end }}

{{/*
Renders the admin (botanist) config.yaml ConfigMap by serialising
.Values.admin.config verbatim, the same way the backend/frontend ConfigMaps do.
The resulting YAML file is mounted into admin pods at /etc/admin/config.yaml.
Secret values should be expressed as configa templates, e.g.:
  {{ secret "admin-secrets" "myKey" }}
which configa resolves against the Kubernetes Secrets API at pod startup.
backend_grpc_url defaults to the in-cluster backend Service's gRPC target
(host:port, no scheme — this is a raw gRPC dial target, not an HTTP URL)
when left empty, mirroring the old per-key ConfigMap's behaviour.
*/}}
{{- define "interledger-app.admin.configMap" -}}
{{- $cfg := deepCopy .Values.admin.config -}}
{{- $_ := set $cfg "backend_grpc_url" (default (printf "%s-backend-service-grpc:8448" (include "common.fullname" .)) $cfg.backend_grpc_url) -}}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "common.fullname" . }}-admin-config
  namespace: {{ .Release.Namespace }}
  {{- include "common.metadata" (list .) | nindent 2 }}
data:
  config.yaml: |
    {{- $cfg | toYaml | nindent 4 }}
{{- end }}
