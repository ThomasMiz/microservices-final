{{- if .Values.env }}
{{- range .Values.env }}
        - name: {{ .name }}
          value: {{ .value | quote }}
{{- end }}
{{- end }}