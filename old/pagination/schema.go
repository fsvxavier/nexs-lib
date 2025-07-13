package pagination

var paginationSchema = `{
	"$schema": "https://json-schema.org/draft/2020-12/schema",
	"$ref": "#/$defs/Pagination",
	"$defs": {
		"Pagination": {
			"properties": {
				"page": {
					"type": "number",
					"minimum": 0
				},
				"limit": {
					"type": "number",
					"minLength": 1, 
                    "maxLength": 3, 
                    "minimum": 1
				},
				"sort": {
					"type": "string"
				},
				"order": {
					"type": "string",
					"minLength": 0, 
					"maxLength": 4, 
					"enum": ["", "asc", "desc", "ASC", "DESC"]
				}
			},
			"additionalProperties": false,
			"type": "object",
			"required": []
		}
	}
}`
