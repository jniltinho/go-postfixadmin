package handlers

import "testing"

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     string
	}{
		{name: "too short", password: "Ab1@234", want: "Password must be at least 8 characters long"},
		{name: "missing uppercase", password: "ab1@2345", want: "Password must contain at least one uppercase letter"},
		{name: "missing lowercase", password: "AB1@2345", want: "Password must contain at least one lowercase letter"},
		{name: "missing number", password: "Abc@defg", want: "Password must contain at least one number"},
		{name: "missing special", password: "Abcdef12", want: "Password must contain at least one special character"},
		{name: "valid", password: "Abcdef1@", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidatePassword(tt.password); got != tt.want {
				t.Fatalf("ValidatePassword() = %q, want %q", got, tt.want)
			}
		})
	}
}
