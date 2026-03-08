package validator

import "testing"

func TestRedactValue(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		sensitive bool
		want      string
	}{
		{
			name:      "sensitive value is redacted",
			value:     "super-secret-password",
			sensitive: true,
			want:      RedactedPlaceholder,
		},
		{
			name:      "non-sensitive value is returned",
			value:     "plain-value",
			sensitive: false,
			want:      "plain-value",
		},
		{
			name:      "non-sensitive long value is truncated",
			value:     "this-is-a-very-long-value-that-exceeds-fifty-characters-and-should-be-truncated",
			sensitive: false,
			want:      "this-is-a-very-long-value-that-exceeds-fifty-chara...",
		},
		{
			name:      "sensitive empty value is redacted",
			value:     "",
			sensitive: true,
			want:      RedactedPlaceholder,
		},
		{
			name:      "non-sensitive empty value is returned",
			value:     "",
			sensitive: false,
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactValue(tt.value, tt.sensitive)
			if got != tt.want {
				t.Errorf("RedactValue() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRedactMessage(t *testing.T) {
	tests := []struct {
		name      string
		message   string
		value     string
		sensitive bool
		want      string
	}{
		{
			name:      "sensitive value is redacted from message",
			message:   "value 'secret123' is invalid",
			value:     "secret123",
			sensitive: true,
			want:      "value '" + RedactedPlaceholder + "' is invalid",
		},
		{
			name:      "non-sensitive value is not redacted",
			message:   "value 'plaintext' is invalid",
			value:     "plaintext",
			sensitive: false,
			want:      "value 'plaintext' is invalid",
		},
		{
			name:      "multiple occurrences are all redacted",
			message:   "value 'secret' does not match expected 'secret'",
			value:     "secret",
			sensitive: true,
			want:      "value '" + RedactedPlaceholder + "' does not match expected '" + RedactedPlaceholder + "'",
		},
		{
			name:      "empty value in sensitive mode returns original message",
			message:   "some error message",
			value:     "",
			sensitive: true,
			want:      "some error message",
		},
		{
			name:      "value not in message returns original",
			message:   "an error occurred",
			value:     "notfound",
			sensitive: true,
			want:      "an error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactMessage(tt.message, tt.value, tt.sensitive)
			if got != tt.want {
				t.Errorf("RedactMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatValueForError(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		sensitive bool
		want      string
	}{
		{
			name:      "sensitive value shows generic 'value'",
			value:     "secret123",
			sensitive: true,
			want:      "value",
		},
		{
			name:      "non-sensitive value shows quoted value",
			value:     "plaintext",
			sensitive: false,
			want:      `value "plaintext"`,
		},
		{
			name:      "non-sensitive long value is truncated",
			value:     "this-is-a-very-long-value-that-exceeds-fifty-characters-and-should-be-truncated",
			sensitive: false,
			want:      `value "this-is-a-very-long-value-that-exceeds-fifty-chara..."`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatValueForError(tt.value, tt.sensitive)
			if got != tt.want {
				t.Errorf("FormatValueForError() = %q, want %q", got, tt.want)
			}
		})
	}
}
