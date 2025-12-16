// Package constraints provides validation constraint types and builders for pedantigo.
package constraints

import (
	"reflect"
)

// Geographic coordinate constraint types.
type (
	latitudeConstraint  struct{} // latitude: validates float -90 to +90 (WGS 84)
	longitudeConstraint struct{} // longitude: validates float -180 to +180 (WGS 84)
)

// Validate checks if the value is a valid latitude (-90 to +90).
func (c latitudeConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // nil/invalid values should skip validation
	}

	// Skip empty strings (handled by required constraint)
	if v.Kind() == reflect.String && v.String() == "" {
		return nil
	}

	num, err := extractNumericValue(v)
	if err != nil {
		return NewConstraintError(CodeInvalidType, "latitude constraint requires numeric value")
	}

	if num < -90 || num > 90 {
		return NewConstraintError(CodeInvalidLatitude, "must be a valid latitude (-90 to 90)")
	}
	return nil
}

// Validate checks if the value is a valid longitude (-180 to +180).
func (c longitudeConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // nil/invalid values should skip validation
	}

	// Skip empty strings (handled by required constraint)
	if v.Kind() == reflect.String && v.String() == "" {
		return nil
	}

	num, err := extractNumericValue(v)
	if err != nil {
		return NewConstraintError(CodeInvalidType, "longitude constraint requires numeric value")
	}

	if num < -180 || num > 180 {
		return NewConstraintError(CodeInvalidLongitude, "must be a valid longitude (-180 to 180)")
	}
	return nil
}
