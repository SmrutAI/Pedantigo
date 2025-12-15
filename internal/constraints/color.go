// Package constraints provides validation constraint types and builders for pedantigo.
package constraints

import (
	"fmt"
	"regexp"
	"strconv"
)

// Color format constraint types.
type (
	hexcolorConstraint struct{} // hexcolor: validates hex color #RGB or #RRGGBB
	rgbConstraint      struct{} // rgb: validates rgb(R,G,B) format
	rgbaConstraint     struct{} // rgba: validates rgba(R,G,B,A) format
	hslConstraint      struct{} // hsl: validates hsl(H,S%,L%) format
	hslaConstraint     struct{} // hsla: validates hsla(H,S%,L%,A) format
)

// Pre-compiled regex patterns for color validation.
var (
	hexcolorRegex = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)
	rgbRegex      = regexp.MustCompile(`^rgb\(\s*(\d{1,3})\s*,\s*(\d{1,3})\s*,\s*(\d{1,3})\s*\)$`)
	rgbaRegex     = regexp.MustCompile(`^rgba\(\s*(\d{1,3})\s*,\s*(\d{1,3})\s*,\s*(\d{1,3})\s*,\s*(0|1|0?\.\d+)\s*\)$`)
	hslRegex      = regexp.MustCompile(`^hsl\(\s*(\d+(?:\.\d+)?)\s*,\s*(\d{1,3})%\s*,\s*(\d{1,3})%\s*\)$`)
	hslaRegex     = regexp.MustCompile(`^hsla\(\s*(\d+(?:\.\d+)?)\s*,\s*(\d{1,3})%\s*,\s*(\d{1,3})%\s*,\s*(0|1|0?\.\d+)\s*\)$`)
)

// Validate checks if the value is a valid hex color (#RGB or #RRGGBB).
func (c hexcolorConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("hexcolor constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	if !hexcolorRegex.MatchString(str) {
		return fmt.Errorf("must be a valid hex color (#RGB or #RRGGBB)")
	}

	return nil
}

// Validate checks if the value is a valid rgb(R,G,B) format.
func (c rgbConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("rgb constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	matches := rgbRegex.FindStringSubmatch(str)
	if matches == nil {
		return fmt.Errorf("must be a valid rgb(R,G,B) color")
	}

	// Validate R, G, B values are 0-255
	for i := 1; i <= 3; i++ {
		val, _ := strconv.Atoi(matches[i])
		if val > 255 {
			return fmt.Errorf("must be a valid rgb(R,G,B) color")
		}
	}

	return nil
}

// Validate checks if the value is a valid rgba(R,G,B,A) format.
func (c rgbaConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("rgba constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	matches := rgbaRegex.FindStringSubmatch(str)
	if matches == nil {
		return fmt.Errorf("must be a valid rgba(R,G,B,A) color")
	}

	// Validate R, G, B values are 0-255
	for i := 1; i <= 3; i++ {
		val, _ := strconv.Atoi(matches[i])
		if val > 255 {
			return fmt.Errorf("must be a valid rgba(R,G,B,A) color")
		}
	}

	// Validate alpha is 0-1 (already constrained by regex pattern)
	alpha, _ := strconv.ParseFloat(matches[4], 64)
	if alpha < 0 || alpha > 1 {
		return fmt.Errorf("must be a valid rgba(R,G,B,A) color")
	}

	return nil
}

// Validate checks if the value is a valid hsl(H,S%,L%) format.
func (c hslConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("hsl constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	matches := hslRegex.FindStringSubmatch(str)
	if matches == nil {
		return fmt.Errorf("must be a valid hsl(H,S%%,L%%) color")
	}

	// Validate H is 0-360
	hue, _ := strconv.ParseFloat(matches[1], 64)
	if hue < 0 || hue > 360 {
		return fmt.Errorf("must be a valid hsl(H,S%%,L%%) color")
	}

	// Validate S is 0-100
	saturation, _ := strconv.Atoi(matches[2])
	if saturation > 100 {
		return fmt.Errorf("must be a valid hsl(H,S%%,L%%) color")
	}

	// Validate L is 0-100
	lightness, _ := strconv.Atoi(matches[3])
	if lightness > 100 {
		return fmt.Errorf("must be a valid hsl(H,S%%,L%%) color")
	}

	return nil
}

// Validate checks if the value is a valid hsla(H,S%,L%,A) format.
func (c hslaConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("hsla constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	matches := hslaRegex.FindStringSubmatch(str)
	if matches == nil {
		return fmt.Errorf("must be a valid hsla(H,S%%,L%%,A) color")
	}

	// Validate H is 0-360
	hue, _ := strconv.ParseFloat(matches[1], 64)
	if hue < 0 || hue > 360 {
		return fmt.Errorf("must be a valid hsla(H,S%%,L%%,A) color")
	}

	// Validate S is 0-100
	saturation, _ := strconv.Atoi(matches[2])
	if saturation > 100 {
		return fmt.Errorf("must be a valid hsla(H,S%%,L%%,A) color")
	}

	// Validate L is 0-100
	lightness, _ := strconv.Atoi(matches[3])
	if lightness > 100 {
		return fmt.Errorf("must be a valid hsla(H,S%%,L%%,A) color")
	}

	// Validate alpha is 0-1 (already constrained by regex pattern)
	alpha, _ := strconv.ParseFloat(matches[4], 64)
	if alpha < 0 || alpha > 1 {
		return fmt.Errorf("must be a valid hsla(H,S%%,L%%,A) color")
	}

	return nil
}
