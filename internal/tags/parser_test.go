package tags

import (
	"reflect"
	"testing"
)

func TestParseTag_Required(t *testing.T) {
	tag := reflect.StructTag(`pedantigo:"required"`)
	constraints := ParseTag(tag)

	if constraints == nil {
		t.Fatal("expected constraints map, got nil")
	}

	if _, ok := constraints["required"]; !ok {
		t.Error("expected 'required' constraint")
	}
}

func TestParseTag_Email(t *testing.T) {
	tag := reflect.StructTag(`pedantigo:"required,email"`)
	constraints := ParseTag(tag)

	if constraints == nil {
		t.Fatal("expected constraints map, got nil")
	}

	if _, ok := constraints["required"]; !ok {
		t.Error("expected 'required' constraint")
	}

	if _, ok := constraints["email"]; !ok {
		t.Error("expected 'email' constraint")
	}
}

func TestParseTag_MinMax(t *testing.T) {
	tag := reflect.StructTag(`pedantigo:"min=18,max=120"`)
	constraints := ParseTag(tag)

	if constraints == nil {
		t.Fatal("expected constraints map, got nil")
	}

	if val, ok := constraints["min"]; !ok || val != "18" {
		t.Errorf("expected min=18, got min=%v", val)
	}

	if val, ok := constraints["max"]; !ok || val != "120" {
		t.Errorf("expected max=120, got max=%v", val)
	}
}

func TestParseTag_Default(t *testing.T) {
	tag := reflect.StructTag(`pedantigo:"default=active"`)
	constraints := ParseTag(tag)

	if constraints == nil {
		t.Fatal("expected constraints map, got nil")
	}

	if val, ok := constraints["default"]; !ok || val != "active" {
		t.Errorf("expected default=active, got default=%v", val)
	}
}

func TestParseTag_NoValidateTag(t *testing.T) {
	tag := reflect.StructTag(`json:"email"`)
	constraints := ParseTag(tag)

	if len(constraints) > 0 {
		t.Errorf("expected empty constraints for tag without validate, got %v", constraints)
	}
}
