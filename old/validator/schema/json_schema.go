package schema

import (
	"github.com/dock-tech/isis-golang-lib/domainerrors"
	"github.com/dock-tech/isis-golang-lib/validator/schema/checks"
	"github.com/xeipuuv/gojsonschema"
)

func init() {
	gojsonschema.FormatCheckers.Add("date_time", checks.DateTimeChecker{})
	gojsonschema.FormatCheckers.Add("text_match", checks.TextMatch{})
	gojsonschema.FormatCheckers.Add("text_match_with_number", checks.TextMatchWithNumber{})
	gojsonschema.FormatCheckers.Add("strong_name", checks.StrongNameFormat{})
	gojsonschema.FormatCheckers.Add("json_number", checks.JsonNumber{})
	gojsonschema.FormatCheckers.Add("iso_8601_date", checks.Iso8601Date{})
	gojsonschema.FormatCheckers.Add("decimal_by_factor_of_8", checks.NewDecimalByFactor8())
	gojsonschema.FormatCheckers.Add("decimal", checks.NewDecimal())
}

func AddCustomFormat(formatName string, regex string) {
	gojsonschema.FormatCheckers.Add(formatName, checks.NewTextMatchCustom(regex))
}

func Validate(loader interface{}, schemaLoader string) error {
	jsonLoader := gojsonschema.NewGoLoader(loader)
	schema := gojsonschema.NewStringLoader(schemaLoader)

	result, err := gojsonschema.Validate(schema, jsonLoader)
	if err != nil {
		return err
	}

	if result.Valid() {
		return nil
	}

	dae := domainerrors.InvalidSchemaError{}
	dae.Details = make(map[string][]string)

	for _, err := range result.Errors() {
		field := err.Field()
		if field == "(root)" {
			property, found := err.Details()["property"]
			if found {
				field = property.(string)
			}
		}

		dae.Details[field] = []string{defaultErrors[err.Type()]}
	}

	return &dae
}
