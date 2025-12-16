// Package constraints provides validation constraint types and builders for pedantigo.
package constraints

import (
	"fmt"
	"reflect"

	"github.com/SmrutAI/Pedantigo/internal/isocodes"
)

// ISO code constraint name constants.
const (
	CISO3166Alpha2   = "iso3166_alpha2"    // ISO 3166-1 alpha-2 country code
	CISO3166Alpha2EU = "iso3166_alpha2_eu" // ISO 3166-1 alpha-2 EU country code
	CISO3166Alpha3   = "iso3166_alpha3"    // ISO 3166-1 alpha-3 country code
	CISO3166Alpha3EU = "iso3166_alpha3_eu" // ISO 3166-1 alpha-3 EU country code
	CISO3166Numeric  = "iso3166_numeric"   // ISO 3166-1 numeric country code
	CISO31662        = "iso3166_2"         // ISO 3166-2 subdivision code
	CISO4217         = "iso4217"           // ISO 4217 currency code
	CISO4217Numeric  = "iso4217_numeric"   // ISO 4217 numeric currency code
	CPostcode        = "postcode"          // Postal code with country parameter
	CBCP47           = "bcp47"             // BCP 47 language tag
)

// ISO code constraint types.
type (
	// iso3166Alpha2Constraint validates ISO 3166-1 alpha-2 country codes (e.g., "US", "GB").
	iso3166Alpha2Constraint struct{}

	// iso3166Alpha2EUConstraint validates EU ISO 3166-1 alpha-2 country codes.
	iso3166Alpha2EUConstraint struct{}

	// iso3166Alpha3Constraint validates ISO 3166-1 alpha-3 country codes (e.g., "USA", "GBR").
	iso3166Alpha3Constraint struct{}

	// iso3166Alpha3EUConstraint validates EU ISO 3166-1 alpha-3 country codes.
	iso3166Alpha3EUConstraint struct{}

	// iso3166NumericConstraint validates ISO 3166-1 numeric country codes.
	iso3166NumericConstraint struct{}

	// iso31662Constraint validates ISO 3166-2 subdivision codes (e.g., "US-CA", "GB-ENG").
	iso31662Constraint struct{}

	// iso4217Constraint validates ISO 4217 currency codes (e.g., "USD", "EUR").
	iso4217Constraint struct{}

	// iso4217NumericConstraint validates ISO 4217 numeric currency codes.
	iso4217NumericConstraint struct{}

	// postcodeConstraint validates postal codes for a specific country.
	// Uses ISO 3166-1 alpha-2 country code as parameter.
	postcodeConstraint struct {
		countryCode string
	}

	// bcp47Constraint validates BCP 47 language tags (e.g., "en", "en-US", "zh-Hans-CN").
	bcp47Constraint struct{}
)

// Validate checks if the value is a valid ISO 3166-1 alpha-2 country code.
func (c iso3166Alpha2Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("iso3166_alpha2 constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	if !isocodes.IsISO3166Alpha2(str) {
		return NewConstraintError(CodeInvalidCountryCode, "must be a valid ISO 3166-1 alpha-2 country code")
	}
	return nil
}

// Validate checks if the value is a valid EU ISO 3166-1 alpha-2 country code.
func (c iso3166Alpha2EUConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil
	}
	if err != nil {
		return fmt.Errorf("iso3166_alpha2_eu constraint %w", err)
	}

	if str == "" {
		return nil
	}

	if !isocodes.IsISO3166Alpha2EU(str) {
		return NewConstraintError(CodeInvalidCountryCode, "must be a valid EU ISO 3166-1 alpha-2 country code")
	}
	return nil
}

// Validate checks if the value is a valid ISO 3166-1 alpha-3 country code.
func (c iso3166Alpha3Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil
	}
	if err != nil {
		return fmt.Errorf("iso3166_alpha3 constraint %w", err)
	}

	if str == "" {
		return nil
	}

	if !isocodes.IsISO3166Alpha3(str) {
		return NewConstraintError(CodeInvalidCountryCode, "must be a valid ISO 3166-1 alpha-3 country code")
	}
	return nil
}

// Validate checks if the value is a valid EU ISO 3166-1 alpha-3 country code.
func (c iso3166Alpha3EUConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil
	}
	if err != nil {
		return fmt.Errorf("iso3166_alpha3_eu constraint %w", err)
	}

	if str == "" {
		return nil
	}

	if !isocodes.IsISO3166Alpha3EU(str) {
		return NewConstraintError(CodeInvalidCountryCode, "must be a valid EU ISO 3166-1 alpha-3 country code")
	}
	return nil
}

// Validate checks if the value is a valid ISO 3166-1 numeric country code.
func (c iso3166NumericConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // skip validation for nil/invalid values
	}

	// Must be an integer type
	var code int
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		code = int(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u := v.Uint()
		if u > 999 { // ISO 3166-1 numeric codes are 1-999
			return NewConstraintError(CodeInvalidCountryCode, "must be a valid ISO 3166-1 numeric country code")
		}
		code = int(u) //nolint:gosec // bounds checked above
	default:
		return fmt.Errorf("iso3166_numeric constraint requires integer value")
	}

	if !isocodes.IsISO3166Numeric(code) {
		return NewConstraintError(CodeInvalidCountryCode, "must be a valid ISO 3166-1 numeric country code")
	}
	return nil
}

// Validate checks if the value is a valid ISO 3166-2 subdivision code.
func (c iso31662Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil
	}
	if err != nil {
		return fmt.Errorf("iso3166_2 constraint %w", err)
	}

	if str == "" {
		return nil
	}

	if !isocodes.IsISO31662(str) {
		return NewConstraintError(CodeInvalidSubdivision, "must be a valid ISO 3166-2 subdivision code")
	}
	return nil
}

// Validate checks if the value is a valid ISO 4217 currency code.
func (c iso4217Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil
	}
	if err != nil {
		return fmt.Errorf("iso4217 constraint %w", err)
	}

	if str == "" {
		return nil
	}

	if !isocodes.IsISO4217(str) {
		return NewConstraintError(CodeInvalidCurrencyCode, "must be a valid ISO 4217 currency code")
	}
	return nil
}

// Validate checks if the value is a valid ISO 4217 numeric currency code.
func (c iso4217NumericConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil
	}

	var code int
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		code = int(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u := v.Uint()
		if u > 999 { // ISO 4217 numeric codes are 1-999
			return NewConstraintError(CodeInvalidCurrencyCode, "must be a valid ISO 4217 numeric currency code")
		}
		code = int(u) //nolint:gosec // bounds checked above
	default:
		return fmt.Errorf("iso4217_numeric constraint requires integer value")
	}

	if !isocodes.IsISO4217Numeric(code) {
		return NewConstraintError(CodeInvalidCurrencyCode, "must be a valid ISO 4217 numeric currency code")
	}
	return nil
}

// Validate checks if the value is a valid postal code for the configured country.
func (c postcodeConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil
	}
	if err != nil {
		return fmt.Errorf("postcode constraint %w", err)
	}

	if str == "" {
		return nil
	}

	// Check if country is supported
	if !isocodes.HasPostcodePattern(c.countryCode) {
		return NewConstraintError(CodeInvalidPostalCode, fmt.Sprintf("postal code validation not supported for country %s", c.countryCode))
	}

	if !isocodes.IsPostcode(str, c.countryCode) {
		return NewConstraintError(CodeInvalidPostalCode, fmt.Sprintf("must be a valid postal code for %s", c.countryCode))
	}
	return nil
}

// Validate checks if the value is a valid BCP 47 language tag.
func (c bcp47Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil
	}
	if err != nil {
		return fmt.Errorf("bcp47 constraint %w", err)
	}

	if str == "" {
		return nil
	}

	if !isocodes.IsBCP47LanguageTag(str) {
		return NewConstraintError(CodeInvalidLanguageTag, "must be a valid BCP 47 language tag")
	}
	return nil
}

// appendISOConstraint appends ISO code constraints based on constraint name.
func appendISOConstraint(result []Constraint, name, value string) []Constraint {
	switch name {
	case CISO3166Alpha2:
		return append(result, iso3166Alpha2Constraint{})
	case CISO3166Alpha2EU:
		return append(result, iso3166Alpha2EUConstraint{})
	case CISO3166Alpha3:
		return append(result, iso3166Alpha3Constraint{})
	case CISO3166Alpha3EU:
		return append(result, iso3166Alpha3EUConstraint{})
	case CISO3166Numeric:
		return append(result, iso3166NumericConstraint{})
	case CISO31662:
		return append(result, iso31662Constraint{})
	case CISO4217:
		return append(result, iso4217Constraint{})
	case CISO4217Numeric:
		return append(result, iso4217NumericConstraint{})
	case CPostcode:
		return append(result, postcodeConstraint{countryCode: value})
	case CBCP47:
		return append(result, bcp47Constraint{})
	}
	return result
}
