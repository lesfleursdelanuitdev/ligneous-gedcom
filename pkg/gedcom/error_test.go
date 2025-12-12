package gedcom

import (
	"testing"
)

func TestGedcomError_Error(t *testing.T) {
	tests := []struct {
		name       string
		error      *GedcomError
		wantPrefix string
	}{
		{
			name: "error with line number",
			error: &GedcomError{
				Severity:   SeverityWarning,
				Message:    "Test error",
				LineNumber: 42,
				Context:    "Test",
			},
			wantPrefix: "warning: Test error (Line 42)",
		},
		{
			name: "error without line number",
			error: &GedcomError{
				Severity:   SeveritySevere,
				Message:    "Severe error",
				LineNumber: 0,
				Context:    "Test",
			},
			wantPrefix: "severe: Severe error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.error.Error()
			if got != tt.wantPrefix {
				t.Errorf("Error() = %q, want %q", got, tt.wantPrefix)
			}
		})
	}
}

func TestErrorManager_AddError(t *testing.T) {
	em := NewErrorManager()

	em.AddError(SeverityWarning, "Test warning", 1, "Test")
	em.AddError(SeveritySevere, "Test severe", 2, "Test")

	if !em.HasErrors() {
		t.Error("Expected errors to be present")
	}

	if !em.HasSevereErrors() {
		t.Error("Expected severe errors to be present")
	}

	errors := em.Errors()
	if len(errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errors))
	}

	if errors[0].Severity != SeverityWarning {
		t.Errorf("Expected first error to be warning, got %v", errors[0].Severity)
	}

	if errors[1].Severity != SeveritySevere {
		t.Errorf("Expected second error to be severe, got %v", errors[1].Severity)
	}
}

func TestErrorManager_HasErrors(t *testing.T) {
	em := NewErrorManager()

	if em.HasErrors() {
		t.Error("Expected no errors initially")
	}

	em.AddError(SeverityWarning, "Test", 1, "Test")
	if !em.HasErrors() {
		t.Error("Expected errors after adding one")
	}
}

func TestErrorManager_HasSevereErrors(t *testing.T) {
	em := NewErrorManager()

	if em.HasSevereErrors() {
		t.Error("Expected no severe errors initially")
	}

	em.AddError(SeverityWarning, "Test", 1, "Test")
	if em.HasSevereErrors() {
		t.Error("Expected no severe errors with only warning")
	}

	em.AddError(SeveritySevere, "Test", 2, "Test")
	if !em.HasSevereErrors() {
		t.Error("Expected severe errors after adding one")
	}
}

func TestErrorManager_GetErrorsBySeverity(t *testing.T) {
	em := NewErrorManager()

	em.AddError(SeverityWarning, "Warning 1", 1, "Test")
	em.AddError(SeveritySevere, "Severe 1", 2, "Test")
	em.AddError(SeverityWarning, "Warning 2", 3, "Test")
	em.AddError(SeveritySevere, "Severe 2", 4, "Test")

	warnings := em.GetErrorsBySeverity(SeverityWarning)
	if len(warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d", len(warnings))
	}

	severe := em.GetErrorsBySeverity(SeveritySevere)
	if len(severe) != 2 {
		t.Errorf("Expected 2 severe errors, got %d", len(severe))
	}
}

func TestErrorManager_GetErrorSummary(t *testing.T) {
	em := NewErrorManager()

	em.AddError(SeverityWarning, "Warning 1", 1, "Test")
	em.AddError(SeveritySevere, "Severe 1", 2, "Test")
	em.AddError(SeverityWarning, "Warning 2", 3, "Test")

	summary := em.GetErrorSummary()
	if summary[SeverityWarning] != 2 {
		t.Errorf("Expected 2 warnings, got %d", summary[SeverityWarning])
	}
	if summary[SeveritySevere] != 1 {
		t.Errorf("Expected 1 severe error, got %d", summary[SeveritySevere])
	}
}

func TestErrorManager_Clear(t *testing.T) {
	em := NewErrorManager()

	em.AddError(SeverityWarning, "Test", 1, "Test")
	em.Clear()

	if em.HasErrors() {
		t.Error("Expected no errors after Clear()")
	}

	if em.Count() != 0 {
		t.Errorf("Expected 0 errors after Clear(), got %d", em.Count())
	}
}

func TestErrorManager_Count(t *testing.T) {
	em := NewErrorManager()

	if em.Count() != 0 {
		t.Errorf("Expected 0 errors initially, got %d", em.Count())
	}

	em.AddError(SeverityWarning, "Test 1", 1, "Test")
	em.AddError(SeverityWarning, "Test 2", 2, "Test")
	em.AddError(SeveritySevere, "Test 3", 3, "Test")

	if em.Count() != 3 {
		t.Errorf("Expected 3 errors, got %d", em.Count())
	}
}

