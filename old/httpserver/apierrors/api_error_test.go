package apierrors

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDockApiError_JsonMap(t *testing.T) {
	expectedCode := "test-code"
	expectedDescription := "test-description"
	dae := NewDockApiError(200, expectedCode, expectedDescription)
	expectedId := dae.InnerError.Id
	jm, _ := dae.JsonMap()

	if errorNode, ok := jm["error"]; ok {
		if errMap, ok := errorNode.(map[string]any); ok {
			if id, ok := errMap["id"]; ok {
				if idS, ok := id.(string); ok {
					if idS != expectedId {
						t.Errorf("expected json node 'error->id' to be %s, got %s", expectedId, idS)
					}
				} else {
					t.Error("json node 'error->id' is not a string type")
				}
			} else {
				t.Error("json missing attribute 'error->status'")
			}

			if code, ok := errMap["code"]; ok {
				if codeS, ok := code.(string); ok {
					if codeS != expectedCode {
						t.Errorf("expected json node 'error->code' to be %s, got %s", expectedCode, codeS)
					}
				} else {
					t.Error("json node 'error->code' is not a string type")
				}
			} else {
				t.Error("json missing attribute 'error->code'")
			}

			if description, ok := errMap["code"]; ok {
				if descriptionS, ok := description.(string); ok {
					if descriptionS != expectedCode {
						t.Errorf("expected json node 'error->description' to be %s, got %s", expectedDescription, descriptionS)
					}
				} else {
					t.Error("json node 'error->description' is not a string type")
				}
			} else {
				t.Error("json missing attribute 'error->description'")
			}

		} else {
			t.Error("json node 'error' is not an object type")
		}
	} else {
		t.Error("json missing attribute 'error'")
	}
}

func TestMakeDockApiErrorCode(t *testing.T) {
	matchRegex := "^[A-Z]{3,5}-[A-Z0-9-_]{3,10}$"
	c := MakeDockApiErrorCode("TEST", "TEST")
	r, _ := regexp.Compile(matchRegex)
	if !r.MatchString(c) {
		t.Errorf("Expected MakeApiErrorCode to return a code matching regex %s Got %s", matchRegex, c)
	}
}

func TestDockApiError_Error(t *testing.T) {
	err := &DockApiError{
		InnerError: DockApiInnerError{
			Id:          "123",
			Code:        "test_code",
			Description: "test description",
		},
	}
	assert.Equal(t, "id=123,code=test_code,description=test description", err.Error())

	err.InnerError.ErrorDetails = []DockApiErrorDetail{
		{
			Attribute: "attr",
			Messages:  []string{"msg1", "msg2"},
		},
	}
	assert.Contains(t, err.Error(), "error_details")
}

func TestMakeDockApiErrorCode2(t *testing.T) {
	code := MakeDockApiErrorCode("base", "detail")
	assert.Equal(t, "base-detail", code)
}

func TestNewDockApiError(t *testing.T) {
	err := NewDockApiError(400, "test_code", "test description")
	assert.Equal(t, "test_code", err.InnerError.Code)
	assert.Equal(t, "test description", err.InnerError.Description)
}

func TestDockApiError_SetId(t *testing.T) {
	err := &DockApiError{}
	assert.NotNil(t, err.SetId("invalid-uuid"))

	validUUID := "123e4567-e89b-12d3-a456-426614174000"
	assert.Nil(t, err.SetId(validUUID))
	assert.Equal(t, validUUID, err.InnerError.Id)
}

func TestDockApiError_AddErrorDetail(t *testing.T) {
	err := &DockApiError{}
	err.AddErrorDetail("attr", "msg1", "msg2")
	assert.Len(t, err.InnerError.ErrorDetails, 1)
	assert.Equal(t, "attr", err.InnerError.ErrorDetails[0].Attribute)
	assert.Equal(t, []string{"msg1", "msg2"}, err.InnerError.ErrorDetails[0].Messages)
}

func TestDockApiError_JsonMap2(t *testing.T) {
	err := &DockApiError{
		InnerError: DockApiInnerError{
			Id:          "123",
			Code:        "test_code",
			Description: "test description",
		},
	}
	jsonMap, _ := err.JsonMap()
	assert.Equal(t, "123", jsonMap["error"].(map[string]interface{})["id"])
	assert.Equal(t, "test_code", jsonMap["error"].(map[string]interface{})["code"])
	assert.Equal(t, "test description", jsonMap["error"].(map[string]interface{})["description"])
}

func TestDockApiError_Bytes(t *testing.T) {
	err := &DockApiError{
		InnerError: DockApiInnerError{
			Id:          "123",
			Code:        "test_code",
			Description: "test description",
		},
	}
	bytes := err.Bytes()
	assert.Contains(t, string(bytes), `"id":"123"`)
	assert.Contains(t, string(bytes), `"code":"test_code"`)
	assert.Contains(t, string(bytes), `"description":"test description"`)
}
