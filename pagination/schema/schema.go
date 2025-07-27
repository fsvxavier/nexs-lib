// Package schema provides JSON schemas for pagination validation.
package schema

// PaginationJSONSchema contains the JSON schema for pagination parameters
const PaginationJSONSchema = `{
	"$schema": "https://json-schema.org/draft/2020-12/schema",
	"$ref": "#/$defs/Pagination",
	"$defs": {
		"Pagination": {
			"properties": {
				"page": {
					"type": "number",
					"minimum": 1,
					"description": "Page number, must be greater than 0"
				},
				"limit": {
					"type": "number",
					"minLength": 1,
                    "maxLength": 3,
					"minimum": 1,
					"maximum": 150,
					"description": "Number of records per page"
				},
				"sort": {
					"type": "string",
					"minLength": 1,
					"maxLength": 100,
					"description": "Field name to sort by"
				},
				"order": {
					"type": "string",
					"minLength": 0,
					"maxLength": 4,
					"enum": ["", "asc", "desc", "ASC", "DESC"],
					"description": "Sort order direction"
				}
			},
			"additionalProperties": false,
			"type": "object",
			"required": []
		}
	}
}`

// GetPaginationSchema returns the pagination JSON schema
func GetPaginationSchema() string {
	return PaginationJSONSchema
}
