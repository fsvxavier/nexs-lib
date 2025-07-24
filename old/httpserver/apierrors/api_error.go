package apierrors

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type DockApiError struct {
	InnerError DockApiInnerError `json:"error,omitempty"`
	StatusCode int               `json:"status_code,omitempty"`
}

type DockApiInnerError struct {
	Id           string               `json:"id,omitempty"`
	Code         string               `json:"code,omitempty"`
	Description  string               `json:"description,omitempty"`
	ErrorDetails []DockApiErrorDetail `json:"error_details,omitempty"`
}

type DockApiErrorDetail struct {
	Attribute string   `json:"attribute,omitempty"`
	Messages  []string `json:"messages,omitempty"`
}

func (dae *DockApiError) Error() string {
	if len(dae.InnerError.ErrorDetails) > 0 {
		return fmt.Sprintf("id=%s,code=%s,description=%s,error_details=%s", dae.InnerError.Id, dae.InnerError.Code, dae.InnerError.Description, dae.InnerError.ErrorDetails)
	}
	return fmt.Sprintf("id=%s,code=%s,description=%s", dae.InnerError.Id, dae.InnerError.Code, dae.InnerError.Description)
}

func MakeDockApiErrorCode(baseCode, detailCode string) string {
	return fmt.Sprintf("%s-%s", baseCode, detailCode)
}

func NewDockApiError(statusCode int, code, description string) *DockApiError {
	return &DockApiError{
		InnerError: DockApiInnerError{
			Id:          uuid.NewString(),
			Code:        code,
			Description: description,
		},
	}
}

func (dae *DockApiError) SetId(id string) error {
	_, err := uuid.Parse(id)
	if err == nil {
		dae.InnerError.Id = id
	}
	return err
}

func (dae *DockApiError) AddErrorDetail(attribute string, messages ...string) *DockApiError {
	if dae.InnerError.ErrorDetails == nil {
		dae.InnerError.ErrorDetails = make([]DockApiErrorDetail, 0)
	}

	ed := DockApiErrorDetail{
		Attribute: attribute,
		Messages:  messages,
	}

	dae.InnerError.ErrorDetails = append(dae.InnerError.ErrorDetails, ed)

	return dae
}

func (dae *DockApiError) JsonMap() (map[string]any, error) {
	var m map[string]any

	b, err := json.Marshal(&dae)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(b, &m)

	return m, nil
}

func (dae *DockApiError) Bytes() []byte {
	b, _ := json.Marshal(dae)
	return b
}
