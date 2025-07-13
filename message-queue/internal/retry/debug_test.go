package retry

import (
	"fmt"
	"testing"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors"
)

func TestTimeoutErrorStructure(t *testing.T) {
	err := domainerrors.NewTimeoutError("operation timed out", "OP")

	fmt.Printf("Error: %v\n", err)
	fmt.Printf("Error Type: %T\n", err)

	// Convert to error interface first
	var errInterface error = err

	if domainErr, ok := errInterface.(*domainerrors.DomainError); ok {
		fmt.Printf("Is DomainError: true\n")
		fmt.Printf("DomainError Type: %v\n", domainErr.Type)
	} else {
		fmt.Printf("Is DomainError: false\n")
	}

	if timeoutErr, ok := errInterface.(*domainerrors.TimeoutError); ok {
		fmt.Printf("Is TimeoutError: true\n")
		if timeoutErr.DomainError != nil {
			fmt.Printf("TimeoutError.DomainError.Type: %v\n", timeoutErr.DomainError.Type)
		}
	} else {
		fmt.Printf("Is TimeoutError: false\n")
	}

	// Test actual function
	result := IsRetryableError(errInterface)
	fmt.Printf("IsRetryableError result: %v\n", result)
}
