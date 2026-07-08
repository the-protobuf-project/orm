# Connection URL for the {{.Database}} database, pre-populated from the proto
# datasource url (regenerated with the schema; git-ignored). Edit locally if
# your credentials differ. {{.Database}}.config.ts reads it via env("{{.EnvVar}}").
{{.EnvVar}}="{{.URLExample}}"
