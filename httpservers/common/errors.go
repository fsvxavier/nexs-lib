package common

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// APIError represents a standardized API error response
type APIError struct {
	InnerError InnerError `json:"error,omitempty"`
	StatusCode int        `json:"status_code,omitempty"`
}

// InnerError contains detailed information about the error
type InnerError struct {
	ID           string        `json:"id,omitempty"`
	Code         string        `json:"code,omitempty"`
	Description  string        `json:"description,omitempty"`
	ErrorDetails []ErrorDetail `json:"error_details,omitempty"`
}

// ErrorDetail contains attribute-specific error details
type ErrorDetail struct {
	Attribute string   `json:"attribute,omitempty"`
	Messages  []string `json:"messages,omitempty"`
}

// Error implements the error interface
func (ae *APIError) Error() string {
	if len(ae.InnerError.ErrorDetails) > 0 {
		return fmt.Sprintf("id=%s,code=%s,description=%s,error_details=%v",
			ae.InnerError.ID, ae.InnerError.Code, ae.InnerError.Description, ae.InnerError.ErrorDetails)
	}
	return fmt.Sprintf("id=%s,code=%s,description=%s",
		ae.InnerError.ID, ae.InnerError.Code, ae.InnerError.Description)
}

// MakeErrorCode creates a formatted error code
func MakeErrorCode(baseCode, detailCode string) string {
	return fmt.Sprintf("%s-%s", baseCode, detailCode)
}

// NewAPIError creates a new APIError
func NewAPIError(statusCode int, code, description string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		InnerError: InnerError{
			ID:          uuid.NewString(),
			Code:        code,
			Description: description,
		},
	}
}

// SetID sets the error ID
func (ae *APIError) SetID(id string) error {
	_, err := uuid.Parse(id)
	if err == nil {
		ae.InnerError.ID = id
	}
	return err
}

// AddErrorDetail adds detailed error information
func (ae *APIError) AddErrorDetail(attribute string, messages ...string) *APIError {
	if ae.InnerError.ErrorDetails == nil {
		ae.InnerError.ErrorDetails = make([]ErrorDetail, 0)
	}

	ed := ErrorDetail{
		Attribute: attribute,
		Messages:  messages,
	}

	ae.InnerError.ErrorDetails = append(ae.InnerError.ErrorDetails, ed)

	return ae
}

// JSONMap converts the error to a map for JSON marshaling
func (ae *APIError) JSONMap() (map[string]any, error) {
	var m map[string]any

	b, err := json.Marshal(&ae)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(b, &m)

	return m, nil
}

// Bytes returns the JSON representation as bytes
func (ae *APIError) Bytes() []byte {
	b, _ := json.Marshal(ae)
	return b
}
