package pedantigo

import (
	"encoding/json"
	"testing"
)

// maskedValue is the expected masked output for secret types.
const maskedValue = "**********"

// TestSecretStr_Value tests that Value() returns the actual secret.
func TestSecretStr_Value(t *testing.T) {
	tests := []struct {
		name   string
		secret string
	}{
		{"simple password", "mysecretpassword"},
		{"api key", "sk-1234567890abcdef"},
		{"empty string", ""},
		{"with special chars", "p@$$w0rd!#$%"},
		{"unicode", "秘密のパスワード"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSecretStr(tt.secret)
			if got := s.Value(); got != tt.secret {
				t.Errorf("SecretStr.Value() = %q, want %q", got, tt.secret)
			}
		})
	}
}

// TestSecretStr_String tests that String() returns masked value.
func TestSecretStr_String(t *testing.T) {
	tests := []struct {
		name   string
		secret string
	}{
		{"simple password", "mysecretpassword"},
		{"api key", "sk-1234567890abcdef"},
		{"empty string", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSecretStr(tt.secret)
			if got := s.String(); got != maskedValue {
				t.Errorf("SecretStr.String() = %q, want %q", got, maskedValue)
			}
		})
	}
}

// TestSecretStr_MarshalJSON tests that MarshalJSON returns masked value.
func TestSecretStr_MarshalJSON(t *testing.T) {
	s := NewSecretStr("mysecretpassword")

	got, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("SecretStr.MarshalJSON() error = %v", err)
	}

	want := `"` + maskedValue + `"`
	if string(got) != want {
		t.Errorf("SecretStr.MarshalJSON() = %s, want %s", got, want)
	}
}

// TestSecretStr_UnmarshalJSON tests that UnmarshalJSON stores actual value.
func TestSecretStr_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    string
		wantErr bool
	}{
		{"simple password", `"mysecretpassword"`, "mysecretpassword", false},
		{"api key", `"sk-1234567890abcdef"`, "sk-1234567890abcdef", false},
		{"empty string", `""`, "", false},
		{"with special chars", `"p@$$w0rd!#$%"`, "p@$$w0rd!#$%", false},
		{"invalid json", `not json`, "", true},
		{"null value", `null`, "", false}, // JSON null unmarshals to empty string
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s SecretStr
			err := json.Unmarshal([]byte(tt.json), &s)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecretStr.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && s.Value() != tt.want {
				t.Errorf("SecretStr.Value() after unmarshal = %q, want %q", s.Value(), tt.want)
			}
		})
	}
}

// TestSecretStr_InStruct tests SecretStr behavior in a struct.
func TestSecretStr_InStruct(t *testing.T) {
	type Config struct {
		APIKey SecretStr `json:"api_key"`
		Name   string    `json:"name"`
	}

	// Unmarshal JSON into struct
	jsonInput := `{"api_key": "secret123", "name": "test"}`
	var config Config
	if err := json.Unmarshal([]byte(jsonInput), &config); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify value is preserved
	if config.APIKey.Value() != "secret123" {
		t.Errorf("APIKey.Value() = %q, want %q", config.APIKey.Value(), "secret123")
	}

	// Marshal back to JSON - should be masked
	output, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	want := `{"api_key":"` + maskedValue + `","name":"test"}`
	if string(output) != want {
		t.Errorf("Marshal output = %s, want %s", output, want)
	}
}

// TestSecretStr_Stringer tests fmt.Stringer interface.
func TestSecretStr_Stringer(t *testing.T) {
	s := NewSecretStr("mysecret")

	// Using String() should show masked value
	if got := s.String(); got != maskedValue {
		t.Errorf("SecretStr.String() = %q, want %q", got, maskedValue)
	}
}

// TestSecretBytes_Value tests that Value() returns the actual secret bytes.
func TestSecretBytes_Value(t *testing.T) {
	tests := []struct {
		name   string
		secret []byte
	}{
		{"simple bytes", []byte{0x01, 0x02, 0x03}},
		{"encryption key", []byte("32byteencryptionkey1234567890ab")},
		{"empty bytes", []byte{}},
		{"nil bytes", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSecretBytes(tt.secret)
			got := s.Value()
			if len(got) != len(tt.secret) {
				t.Errorf("SecretBytes.Value() len = %d, want %d", len(got), len(tt.secret))
				return
			}
			for i := range got {
				if got[i] != tt.secret[i] {
					t.Errorf("SecretBytes.Value()[%d] = %d, want %d", i, got[i], tt.secret[i])
				}
			}
		})
	}
}

// TestSecretBytes_String tests that String() returns masked value.
func TestSecretBytes_String(t *testing.T) {
	s := NewSecretBytes([]byte{0x01, 0x02, 0x03})
	if got := s.String(); got != maskedValue {
		t.Errorf("SecretBytes.String() = %q, want %q", got, maskedValue)
	}
}

// TestSecretBytes_MarshalJSON tests that MarshalJSON returns masked value.
func TestSecretBytes_MarshalJSON(t *testing.T) {
	s := NewSecretBytes([]byte{0x01, 0x02, 0x03})

	got, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("SecretBytes.MarshalJSON() error = %v", err)
	}

	want := `"` + maskedValue + `"`
	if string(got) != want {
		t.Errorf("SecretBytes.MarshalJSON() = %s, want %s", got, want)
	}
}

// TestSecretBytes_UnmarshalJSON tests that UnmarshalJSON decodes base64 and stores actual value.
func TestSecretBytes_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    []byte
		wantErr bool
	}{
		{"valid base64 hello", `"SGVsbG8="`, []byte("Hello"), false},
		{"valid base64 world", `"V29ybGQ="`, []byte("World"), false},
		{"valid base64 empty", `""`, []byte{}, false},
		{"invalid base64", `"not valid base64!"`, nil, true},
		{"invalid json", `not json`, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s SecretBytes
			err := json.Unmarshal([]byte(tt.json), &s)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecretBytes.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				got := s.Value()
				if len(got) != len(tt.want) {
					t.Errorf("SecretBytes.Value() len = %d, want %d", len(got), len(tt.want))
					return
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("SecretBytes.Value()[%d] = %d, want %d", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}
