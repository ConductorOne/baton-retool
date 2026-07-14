package connector

import "testing"

func TestParseObjectID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantID  int64
		wantErr bool
	}{
		{"empty string", "", 0, true},
		{"single char prefix only", "u", 0, true},
		{"valid user id", "u12345", 12345, false},
		{"invalid non-numeric suffix", "uabc", 0, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseObjectID(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("parseObjectID(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
			}
			if !tc.wantErr && got != tc.wantID {
				t.Fatalf("parseObjectID(%q) = %d, want %d", tc.input, got, tc.wantID)
			}
		})
	}
}
