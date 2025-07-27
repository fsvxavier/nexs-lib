// Package unmarshaling provides automatic response unmarshaling based on Content-Type.
package unmarshaling

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"mime"
	"strings"

	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
)

// Unmarshaler handles automatic response unmarshaling based on Content-Type.
type Unmarshaler struct {
	strategy interfaces.UnmarshalStrategy
}

// NewUnmarshaler creates a new unmarshaler with the specified strategy.
func NewUnmarshaler(strategy interfaces.UnmarshalStrategy) *Unmarshaler {
	return &Unmarshaler{
		strategy: strategy,
	}
}

// UnmarshalResponse automatically unmarshals response based on Content-Type.
func (u *Unmarshaler) UnmarshalResponse(resp *interfaces.Response, v interface{}) error {
	if resp == nil || resp.Body == nil {
		return fmt.Errorf("response or response body is nil")
	}

	if len(resp.Body) == 0 {
		return fmt.Errorf("response body is empty")
	}

	// Determine unmarshal strategy
	strategy := u.determineStrategy(resp)

	switch strategy {
	case interfaces.UnmarshalJSON:
		return u.unmarshalJSON(resp.Body, v)
	case interfaces.UnmarshalXML:
		return u.unmarshalXML(resp.Body, v)
	case interfaces.UnmarshalNone:
		return u.unmarshalRaw(resp.Body, v)
	default:
		return u.unmarshalAuto(resp, v)
	}
}

// determineStrategy determines the unmarshaling strategy based on response Content-Type.
func (u *Unmarshaler) determineStrategy(resp *interfaces.Response) interfaces.UnmarshalStrategy {
	if u.strategy != interfaces.UnmarshalAuto {
		return u.strategy
	}

	contentType := resp.ContentType
	if contentType == "" && resp.Headers != nil {
		contentType = resp.Headers["Content-Type"]
	}

	if contentType == "" {
		return interfaces.UnmarshalJSON // Default to JSON
	}

	// Parse media type
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return interfaces.UnmarshalJSON // Default to JSON on parse error
	}

	// Determine strategy based on media type
	switch {
	case strings.HasSuffix(mediaType, "/json") || strings.HasSuffix(mediaType, "+json"):
		return interfaces.UnmarshalJSON
	case strings.HasSuffix(mediaType, "/xml") || strings.HasSuffix(mediaType, "+xml"):
		return interfaces.UnmarshalXML
	case strings.HasPrefix(mediaType, "text/"):
		return interfaces.UnmarshalNone
	case strings.HasPrefix(mediaType, "application/octet-stream"):
		return interfaces.UnmarshalNone
	default:
		return interfaces.UnmarshalJSON // Default to JSON for unknown types
	}
}

// unmarshalJSON unmarshals JSON response.
func (u *Unmarshaler) unmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// unmarshalXML unmarshals XML response.
func (u *Unmarshaler) unmarshalXML(data []byte, v interface{}) error {
	return xml.Unmarshal(data, v)
}

// unmarshalRaw handles raw data unmarshaling.
func (u *Unmarshaler) unmarshalRaw(data []byte, v interface{}) error {
	switch target := v.(type) {
	case *[]byte:
		*target = data
		return nil
	case *string:
		*target = string(data)
		return nil
	default:
		return fmt.Errorf("unsupported type for raw unmarshaling: %T", v)
	}
}

// unmarshalAuto attempts automatic unmarshaling based on content analysis.
func (u *Unmarshaler) unmarshalAuto(resp *interfaces.Response, v interface{}) error {
	data := resp.Body

	// Try to detect format from content
	trimmed := strings.TrimSpace(string(data))

	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		return u.unmarshalJSON(data, v)
	}

	if strings.HasPrefix(trimmed, "<") {
		return u.unmarshalXML(data, v)
	}

	// Default to raw
	return u.unmarshalRaw(data, v)
}

// GetSupportedContentTypes returns the list of supported content types.
func GetSupportedContentTypes() []string {
	return []string{
		"application/json",
		"application/xml",
		"text/xml",
		"text/plain",
		"application/octet-stream",
		"text/html",
	}
}
