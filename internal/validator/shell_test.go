package validator

import "testing"

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr bool
	}{
		{"empty", "", true},
		{"whitespace", "   ", true},
		{"safe", "echo hello", false},
		{"dangerous rm", "rm -rf /", true},
		{"fork bomb", ":(){ :|:& };:", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Validate(tt.script); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
