{{- if .Values.fail }}
echo "OH"
echo "IT FAILED!"
exit 1
{{- end }}