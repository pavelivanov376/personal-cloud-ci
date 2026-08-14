{{/*
rollDate annotation — changes on every deploy so `helm upgrade` rolls the pod.
start.sh sets .Values.rollDate; falls back to install time if unset.
Use inside a Deployment's pod template metadata:

    template:
      metadata:
        annotations:
          {{- include "cicloud.rollAnnotation" . | nindent 10 }}
*/}}
{{- define "cicloud.rollAnnotation" -}}
cicloud.sap.com/rollDate: {{ .Values.rollDate | default now | quote }}
{{- end -}}
