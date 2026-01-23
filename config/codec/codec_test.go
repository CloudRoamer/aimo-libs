package codec

import "testing"

func TestJSONCodec_EncodeDecode(t *testing.T) {
	input := map[string]any{
		"name":  "app",
		"count": 2,
	}

	data, err := JSON.Encode(input)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	var decoded map[string]any
	if err := JSON.Decode(data, &decoded); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if decoded["name"] != "app" {
		t.Errorf("Decoded name = %v, want app", decoded["name"])
	}
	if decoded["count"] != float64(2) {
		t.Errorf("Decoded count = %v, want 2", decoded["count"])
	}
}

func TestJSONCodec_DecodeInvalid(t *testing.T) {
	var decoded map[string]any
	if err := JSON.Decode([]byte("{invalid"), &decoded); err == nil {
		t.Fatal("Decode() should fail for invalid JSON")
	}
}

func TestYAMLCodec_EncodeDecode(t *testing.T) {
	input := map[string]any{
		"name":  "app",
		"count": 2,
	}

	data, err := YAML.Encode(input)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	var decoded map[string]any
	if err := YAML.Decode(data, &decoded); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if decoded["name"] != "app" {
		t.Errorf("Decoded name = %v, want app", decoded["name"])
	}
	if decoded["count"] != 2 {
		t.Errorf("Decoded count = %v, want 2", decoded["count"])
	}
}

func TestYAMLCodec_DecodeInvalid(t *testing.T) {
	var decoded map[string]any
	if err := YAML.Decode([]byte("key: [invalid"), &decoded); err == nil {
		t.Fatal("Decode() should fail for invalid YAML")
	}
}
