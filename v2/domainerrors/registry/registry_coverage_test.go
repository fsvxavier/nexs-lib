package registry

import (
	"fmt"
	"testing"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

func TestRegistryBasicOperations(t *testing.T) {
	t.Run("test get registered code", func(t *testing.T) {
		registry := NewErrorCodeRegistry()

		code := interfaces.ErrorCodeInfo{
			Code:       "GET_TEST",
			Message:    "Get test error",
			Type:       string(types.ErrorTypeValidation),
			StatusCode: 400,
			Severity:   interfaces.SeverityMedium,
			Category:   interfaces.CategoryBusiness,
		}

		err := registry.Register(code)
		if err != nil {
			t.Errorf("Register failed: %v", err)
		}

		retrieved, exists := registry.Get("GET_TEST")
		if !exists {
			t.Error("Get should return true for existing code")
		}

		if retrieved.Code != code.Code {
			t.Errorf("Expected code %s, got %s", code.Code, retrieved.Code)
		}
		if retrieved.Message != code.Message {
			t.Errorf("Expected message %s, got %s", code.Message, retrieved.Message)
		}
	})

	t.Run("test get non-existent code", func(t *testing.T) {
		registry := NewErrorCodeRegistry()

		_, exists := registry.Get("NON_EXISTENT")
		if exists {
			t.Error("Get should return false for non-existent code")
		}
	})

	t.Run("test list all codes", func(t *testing.T) {
		registry := NewErrorCodeRegistry()

		codes := []interfaces.ErrorCodeInfo{
			{Code: "LIST_001", Message: "List test 1", Type: string(types.ErrorTypeValidation), StatusCode: 400},
			{Code: "LIST_002", Message: "List test 2", Type: string(types.ErrorTypeInternal), StatusCode: 500},
			{Code: "LIST_003", Message: "List test 3", Type: string(types.ErrorTypeNetwork), StatusCode: 502},
		}

		for _, code := range codes {
			err := registry.Register(code)
			if err != nil {
				t.Errorf("Register failed: %v", err)
			}
		}

		allCodes := registry.List()
		// Deve incluir os códigos padrão + os que registramos
		if len(allCodes) < len(codes) {
			t.Errorf("Expected at least %d codes, got %d", len(codes), len(allCodes))
		}

		// Verificar se nossos códigos estão na lista
		foundCodes := 0
		for _, registeredCode := range allCodes {
			for _, expectedCode := range codes {
				if registeredCode.Code == expectedCode.Code {
					foundCodes++
					break
				}
			}
		}

		if foundCodes != len(codes) {
			t.Errorf("Expected to find %d registered codes, found %d", len(codes), foundCodes)
		}
	})

	t.Run("test create error from registry", func(t *testing.T) {
		registry := NewErrorCodeRegistry()

		code := interfaces.ErrorCodeInfo{
			Code:       "CREATE_TEST",
			Message:    "Create test error: %s",
			Type:       string(types.ErrorTypeValidation),
			StatusCode: 400,
			Severity:   interfaces.SeverityMedium,
			Category:   interfaces.CategoryBusiness,
			Retryable:  false,
			Temporary:  false,
		}

		err := registry.Register(code)
		if err != nil {
			t.Errorf("Register failed: %v", err)
		}

		// Testar CreateError com argumentos
		domainErr, err := registry.CreateError("CREATE_TEST", "custom message")
		if err != nil {
			t.Errorf("CreateError failed: %v", err)
		}

		if domainErr == nil {
			t.Error("CreateError should return a domain error")
		}

		if domainErr.Error() == "" {
			t.Error("Domain error should have an error message")
		}

		// Testar CreateError sem argumentos
		domainErr2, err := registry.CreateError("CREATE_TEST")
		if err != nil {
			t.Errorf("CreateError without args failed: %v", err)
		}

		if domainErr2 == nil {
			t.Error("CreateError without args should return a domain error")
		}
	})

	t.Run("test create error with non-existent code", func(t *testing.T) {
		registry := NewErrorCodeRegistry()

		_, err := registry.CreateError("NON_EXISTENT")
		if err == nil {
			t.Error("CreateError should fail for non-existent code")
		}
	})
}

func TestRegistryExtendedOperations(t *testing.T) {
	t.Run("test concrete registry methods", func(t *testing.T) {
		// Cast para o tipo concreto para acessar métodos estendidos
		concreteRegistry := NewErrorCodeRegistry().(*ErrorCodeRegistry)

		// Test RegisterMultiple
		codes := []interfaces.ErrorCodeInfo{
			{Code: "MULTI_001", Message: "Multi test 1", Type: string(types.ErrorTypeValidation), StatusCode: 400},
			{Code: "MULTI_002", Message: "Multi test 2", Type: string(types.ErrorTypeInternal), StatusCode: 500},
		}

		err := concreteRegistry.RegisterMultiple(codes)
		if err != nil {
			t.Errorf("RegisterMultiple failed: %v", err)
		}

		// Verificar se foram registrados
		for _, code := range codes {
			if !concreteRegistry.Exists(code.Code) {
				t.Errorf("Code %s should exist after RegisterMultiple", code.Code)
			}
		}

		// Test ListByType
		validationCodes := concreteRegistry.ListByType(string(types.ErrorTypeValidation))
		t.Logf("Found %d validation codes", len(validationCodes))

		// Test ListBySeverity
		mediumCodes := concreteRegistry.ListBySeverity(interfaces.SeverityMedium)
		t.Logf("Found %d medium severity codes", len(mediumCodes))

		// Test Count
		count := concreteRegistry.Count()
		if count == 0 {
			t.Error("Count should be greater than 0")
		}
		t.Logf("Registry has %d codes", count)

		// Test Search
		searchResults := concreteRegistry.Search("test")
		t.Logf("Search for 'test' returned %d results", len(searchResults))

		// Test Update (apenas um parâmetro)
		updated := interfaces.ErrorCodeInfo{
			Code:       "MULTI_001",
			Message:    "Updated multi test 1",
			Type:       string(types.ErrorTypeValidation),
			StatusCode: 400,
		}
		err = concreteRegistry.Update(updated)
		if err != nil {
			t.Errorf("Update failed: %v", err)
		}

		retrieved, exists := concreteRegistry.Get("MULTI_001")
		if !exists {
			t.Error("Updated code should still exist")
		}
		if retrieved.Message != "Updated multi test 1" {
			t.Errorf("Expected updated message, got: %s", retrieved.Message)
		}

		// Test Remove
		err = concreteRegistry.Remove("MULTI_002")
		if err != nil {
			t.Errorf("Remove failed: %v", err)
		}

		if concreteRegistry.Exists("MULTI_002") {
			t.Error("Removed code should not exist")
		}

		// Test Export
		exported := concreteRegistry.Export()
		if len(exported) == 0 {
			t.Error("Export should return data")
		}

		// Test Import
		newRegistry := NewErrorCodeRegistry().(*ErrorCodeRegistry)
		err = newRegistry.Import(exported, true) // Use overwrite = true
		if err != nil {
			t.Errorf("Import failed: %v", err)
		}

		// Test Clear (fazer por último)
		initialCount := concreteRegistry.Count()
		concreteRegistry.Clear()
		if concreteRegistry.Count() >= initialCount {
			t.Error("Clear should reduce the count")
		}
	})
}

func TestGlobalRegistryFunctions(t *testing.T) {
	t.Run("test global registry operations", func(t *testing.T) {
		// Test GetGlobalRegistry
		globalReg := GetGlobalRegistry()
		if globalReg == nil {
			t.Error("GetGlobalRegistry should return a registry")
		}

		// Test SetGlobalRegistry
		newReg := NewErrorCodeRegistry()
		SetGlobalRegistry(newReg)

		// Test RegisterGlobal
		err := RegisterGlobal(interfaces.ErrorCodeInfo{
			Code:       "GLOBAL_TEST",
			Message:    "Global test error",
			Type:       string(types.ErrorTypeValidation),
			StatusCode: 400,
		})
		if err != nil {
			t.Errorf("RegisterGlobal failed: %v", err)
		}

		// Test GetGlobal
		retrieved, exists := GetGlobal("GLOBAL_TEST")
		if !exists {
			t.Error("GetGlobal should return true for registered code")
		}
		if retrieved.Code != "GLOBAL_TEST" {
			t.Errorf("Expected code GLOBAL_TEST, got %s", retrieved.Code)
		}

		// Test CreateErrorGlobal
		domainErr, err := CreateErrorGlobal("GLOBAL_TEST", "Custom message")
		if err != nil {
			t.Errorf("CreateErrorGlobal failed: %v", err)
		}
		if domainErr == nil {
			t.Error("CreateErrorGlobal should return a domain error")
		}

		// Test ExistsGlobal
		if !ExistsGlobal("GLOBAL_TEST") {
			t.Error("ExistsGlobal should return true for registered code")
		}

		if ExistsGlobal("NON_EXISTENT_GLOBAL") {
			t.Error("ExistsGlobal should return false for non-existent code")
		}
	})
}

func TestRegistryEdgeCases(t *testing.T) {
	t.Run("test nil and empty values", func(t *testing.T) {
		registry := NewErrorCodeRegistry().(*ErrorCodeRegistry)

		// Test with empty code
		emptyCode := interfaces.ErrorCodeInfo{
			Code:       "",
			Message:    "Empty code test",
			Type:       string(types.ErrorTypeValidation),
			StatusCode: 400,
		}
		err := registry.Register(emptyCode)
		if err == nil {
			t.Error("Register should fail for empty code")
		}

		// Test with empty message
		emptyMessage := interfaces.ErrorCodeInfo{
			Code:       "EMPTY_MSG",
			Message:    "",
			Type:       string(types.ErrorTypeValidation),
			StatusCode: 400,
		}
		err = registry.Register(emptyMessage)
		if err == nil {
			t.Error("Register should fail for empty message")
		}

		// Test with invalid status code
		invalidStatus := interfaces.ErrorCodeInfo{
			Code:       "INVALID_STATUS",
			Message:    "Invalid status test",
			Type:       string(types.ErrorTypeValidation),
			StatusCode: 999,
		}
		err = registry.Register(invalidStatus)
		if err == nil {
			t.Error("Register should fail for invalid status code")
		}

		// Test with negative status code
		negativeStatus := interfaces.ErrorCodeInfo{
			Code:       "NEGATIVE_STATUS",
			Message:    "Negative status test",
			Type:       string(types.ErrorTypeValidation),
			StatusCode: -1,
		}
		err = registry.Register(negativeStatus)
		if err == nil {
			t.Error("Register should fail for negative status code")
		}
	})

	t.Run("test duplicate registration", func(t *testing.T) {
		registry := NewErrorCodeRegistry().(*ErrorCodeRegistry)

		code := interfaces.ErrorCodeInfo{
			Code:       "DUPLICATE_TEST",
			Message:    "Duplicate test",
			Type:       string(types.ErrorTypeValidation),
			StatusCode: 400,
		}

		// First registration should succeed
		err := registry.Register(code)
		if err != nil {
			t.Errorf("First registration failed: %v", err)
		}

		// Second registration should fail
		err = registry.Register(code)
		if err == nil {
			t.Error("Second registration should fail for duplicate code")
		}
	})

	t.Run("test search edge cases", func(t *testing.T) {
		registry := NewErrorCodeRegistry().(*ErrorCodeRegistry)

		// Search with empty string
		results := registry.Search("")
		t.Logf("Search with empty string returned %d results", len(results))

		// Search with non-existent term
		results = registry.Search("NONEXISTENT_TERM_12345")
		if len(results) != 0 {
			t.Errorf("Search for non-existent term should return 0 results, got %d", len(results))
		}

		// Search with special characters
		results = registry.Search("@#$%^&*()")
		t.Logf("Search with special characters returned %d results", len(results))
	})

	t.Run("test list by non-existent type", func(t *testing.T) {
		registry := NewErrorCodeRegistry().(*ErrorCodeRegistry)

		results := registry.ListByType("NON_EXISTENT_TYPE")
		if len(results) != 0 {
			t.Errorf("ListByType for non-existent type should return 0 results, got %d", len(results))
		}
	})
	t.Run("test list by non-existent severity", func(t *testing.T) {
		registry := NewErrorCodeRegistry().(*ErrorCodeRegistry)

		results := registry.ListBySeverity(interfaces.Severity(999)) // Invalid severity
		if len(results) != 0 {
			t.Errorf("ListBySeverity for non-existent severity should return 0 results, got %d", len(results))
		}
	})

	t.Run("test remove non-existent code", func(t *testing.T) {
		registry := NewErrorCodeRegistry().(*ErrorCodeRegistry)

		err := registry.Remove("NON_EXISTENT_CODE")
		if err == nil {
			t.Error("Remove should fail for non-existent code")
		}
	})

	t.Run("test update non-existent code", func(t *testing.T) {
		registry := NewErrorCodeRegistry().(*ErrorCodeRegistry)

		code := interfaces.ErrorCodeInfo{
			Code:       "NON_EXISTENT_UPDATE",
			Message:    "Non-existent update test",
			Type:       string(types.ErrorTypeValidation),
			StatusCode: 400,
		}

		err := registry.Update(code)
		if err == nil {
			t.Error("Update should fail for non-existent code")
		}
	})

	t.Run("test export and import edge cases", func(t *testing.T) {
		emptyRegistry := NewErrorCodeRegistry().(*ErrorCodeRegistry)
		emptyRegistry.Clear()

		// Export empty registry
		exported := emptyRegistry.Export()
		t.Logf("Exported %d codes from empty registry", len(exported))

		// Import into another registry
		newRegistry := NewErrorCodeRegistry().(*ErrorCodeRegistry)
		err := newRegistry.Import(exported, false)
		if err != nil {
			t.Errorf("Import of empty data failed: %v", err)
		} // Test import with overwrite
		code := interfaces.ErrorCodeInfo{
			Code:       "IMPORT_TEST",
			Message:    "Import test",
			Type:       string(types.ErrorTypeValidation),
			StatusCode: 400,
		}
		newRegistry.Register(code)

		// Export and re-import to test import functionality
		exportedData := newRegistry.Export()
		anotherRegistry := NewErrorCodeRegistry().(*ErrorCodeRegistry)

		// Try to import without overwrite
		err = anotherRegistry.Import(exportedData, false)
		if err == nil {
			t.Error("Import without overwrite should fail for some existing codes")
		}

		// Try to import with overwrite
		err = anotherRegistry.Import(exportedData, true)
		if err != nil {
			t.Errorf("Import with overwrite should succeed: %v", err)
		}
	})
}

func TestRegistryStressTest(t *testing.T) {
	t.Run("test many registrations", func(t *testing.T) {
		registry := NewErrorCodeRegistry().(*ErrorCodeRegistry)

		// Register many codes
		for i := 0; i < 100; i++ {
			code := interfaces.ErrorCodeInfo{
				Code:       fmt.Sprintf("STRESS_%03d", i),
				Message:    fmt.Sprintf("Stress test %d", i),
				Type:       string(types.ErrorTypeValidation),
				StatusCode: 400,
			}
			err := registry.Register(code)
			if err != nil {
				t.Errorf("Failed to register stress code %d: %v", i, err)
			}
		}

		// Verify count
		count := registry.Count()
		if count < 100 {
			t.Errorf("Expected at least 100 codes, got %d", count)
		}

		// Test bulk operations
		codes := make([]interfaces.ErrorCodeInfo, 10)
		for i := 0; i < 10; i++ {
			codes[i] = interfaces.ErrorCodeInfo{
				Code:       fmt.Sprintf("BULK_%03d", i),
				Message:    fmt.Sprintf("Bulk test %d", i),
				Type:       string(types.ErrorTypeInternal),
				StatusCode: 500,
			}
		}

		err := registry.RegisterMultiple(codes)
		if err != nil {
			t.Errorf("RegisterMultiple failed: %v", err)
		}

		// Search for bulk codes
		results := registry.Search("BULK")
		if len(results) < 10 {
			t.Errorf("Expected at least 10 bulk codes, got %d", len(results))
		}
	})
}
