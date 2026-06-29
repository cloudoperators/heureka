side_cars:
    - name: linkerd-proxy
      service: linkerd
      support_group: containers
{{- if .Values.scanner.excluded_pods }}
excluded_pods:
{{- range .Values.scanner.excluded_pods }}
    - {{ . }}
{{- end }}
{{- end }}
