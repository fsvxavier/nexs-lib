package paginate

// Schema contém a definição do esquema JSON para validação de parâmetros de paginação
var Schema = `{
	"$schema": "https://json-schema.org/draft/2020-12/schema",
	"$ref": "#/$defs/Page",
	"$defs": {
		"Page": {
			"properties": {
				"page": {
					"type": "number",
					"minimum": 1
				},
				"limit": {
					"type": "number",
					"minimum": 1,
					"maximum": 150
				},
				"sort": {
					"type": "string"
				},
				"order": {
					"type": "string",
					"enum": ["", "asc", "desc", "ASC", "DESC"]
				}
			},
			"additionalProperties": false,
			"type": "object",
			"required": []
		}
	}
}`
