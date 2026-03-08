package cli

import "testing"

func TestExitCodeDescription(t *testing.T) {
	tests := []struct {
		code int
		want string
	}{
		{ExitSuccess, "all validations passed"},
		{ExitValidation, "one or more validation failures"},
		{ExitSchemaError, "schema file could not be parsed or is invalid"},
		{99, "unknown exit code"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := ExitCodeDescription(tt.code)
			if got != tt.want {
				t.Errorf("ExitCodeDescription(%d) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestExitCodeValues(t *testing.T) {
	// Ensure exit codes are exactly as specified
	if ExitSuccess != 0 {
		t.Errorf("ExitSuccess = %d, want 0", ExitSuccess)
	}
	if ExitValidation != 1 {
		t.Errorf("ExitValidation = %d, want 1", ExitValidation)
	}
	if ExitSchemaError != 2 {
		t.Errorf("ExitSchemaError = %d, want 2", ExitSchemaError)
	}
}
