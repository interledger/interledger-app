{{/*
Returns a {create: false, name: "<name>"} dict for the serviceAccount argument of
common.deployment. Using create: false tells common not to try to create the SA itself —
we handle that in the serviceaccount.*.yaml templates. The name is:
  - serviceAccount.name if explicitly set
  - <fullname>-<svcName> when serviceAccount.create is true and name is empty
  - ""  when serviceAccount.create is false and name is empty (pod uses namespace default)

Usage: include "mock-services.saSpec" (list $top $svcValues "svcName") | fromYaml
*/}}
{{- define "mock-services.saSpec" -}}
{{- $top := first . -}}
{{- $svcValues := index . 1 -}}
{{- $svcName := index . 2 -}}
{{- $sa := $svcValues.serviceAccount | default (dict) -}}
{{- $name := "" -}}
{{- if $sa.create -}}
  {{- $name = $sa.name | default (printf "%s-%s" (include "common.fullname" $top) $svcName) -}}
{{- else -}}
  {{- $name = $sa.name | default "" -}}
{{- end -}}
create: false
name: {{ $name | quote }}
{{- end }}

{{/*
Renders ServiceAccount, Role, and RoleBinding for a mock service.

  serviceAccount.create: true   — renders the ServiceAccount
  secretAccess: [...]           — additionally renders a Role (get on named secrets)
                                  and a RoleBinding attaching it to the SA

The Role only grants get on the explicitly listed secret names — no wildcards.

Usage: include "mock-services.rbac" (list $top $svcValues "svcName")
*/}}
{{- define "mock-services.rbac" -}}
{{- $top := first . -}}
{{- $svcValues := index . 1 -}}
{{- $svcName := index . 2 -}}
{{- $sa := $svcValues.serviceAccount | default (dict) -}}
{{- if $sa.create -}}
{{- $secretAccess := $svcValues.secretAccess | default (list) -}}
{{- $saName := $sa.name | default (printf "%s-%s" (include "common.fullname" $top) $svcName) -}}
{{- $roleName := printf "%s-%s-secret-reader" (include "common.fullname" $top) $svcName -}}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ $saName }}
  namespace: {{ $top.Release.Namespace }}
  {{- include "common.metadata" (list $top) | nindent 2 }}
{{- if gt (len $secretAccess) 0 }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{ $roleName }}
  namespace: {{ $top.Release.Namespace }}
  {{- include "common.metadata" (list $top) | nindent 2 }}
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    resourceNames:
      {{- range $secretAccess }}
      - {{ . | quote }}
      {{- end }}
    verbs: ["get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{ $roleName }}
  namespace: {{ $top.Release.Namespace }}
  {{- include "common.metadata" (list $top) | nindent 2 }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: {{ $roleName }}
subjects:
  - kind: ServiceAccount
    name: {{ $saName }}
    namespace: {{ $top.Release.Namespace }}
{{- end }}
{{- end }}
{{- end }}
