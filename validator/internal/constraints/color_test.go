package constraints

import "testing"

// TestHexcolorConstraint tests hexcolorConstraint.Validate() for valid hex color formats.
func TestHexcolorConstraint(t *testing.T) {
	runSimpleConstraintTests(t, hexcolorConstraint{}, []simpleTestCase{
		// Valid hex colors - 3 char shorthand
		{"valid 3 char lowercase", "#fff", false},
		{"valid 3 char uppercase", "#FFF", false},
		{"valid 3 char mixed", "#fFf", false},
		{"valid 3 char with numbers", "#123", false},
		{"valid 3 char abc", "#abc", false},
		// Valid hex colors - 6 char full
		{"valid 6 char lowercase", "#ffffff", false},
		{"valid 6 char uppercase", "#FFFFFF", false},
		{"valid 6 char mixed", "#FfFfFf", false},
		{"valid 6 char black", "#000000", false},
		{"valid 6 char red", "#ff0000", false},
		{"valid 6 char green", "#00ff00", false},
		{"valid 6 char blue", "#0000ff", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid hex colors
		{"invalid no hash", "fff", true},
		{"invalid no hash 6 char", "ffffff", true},
		{"invalid char g", "#gggggg", true},
		{"invalid char z", "#zzzzzz", true},
		{"invalid 2 chars", "#ff", true},
		{"invalid 4 chars", "#ffff", true},
		{"invalid 5 chars", "#fffff", true},
		{"invalid 7 chars", "#fffffff", true},
		{"invalid 8 chars (rgba)", "#ffffffff", true},
		{"invalid double hash", "##ffffff", true},
		{"invalid space", "# ffffff", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestRgbConstraint tests rgbConstraint.Validate() for valid rgb() color formats.
func TestRgbConstraint(t *testing.T) {
	runSimpleConstraintTests(t, rgbConstraint{}, []simpleTestCase{
		// Valid rgb colors
		{"valid white", "rgb(255,255,255)", false},
		{"valid black", "rgb(0,0,0)", false},
		{"valid red", "rgb(255,0,0)", false},
		{"valid green", "rgb(0,255,0)", false},
		{"valid blue", "rgb(0,0,255)", false},
		{"valid with spaces", "rgb(0, 0, 0)", false},
		{"valid with spaces after commas", "rgb(100, 150, 200)", false},
		{"valid single digits", "rgb(1,2,3)", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid rgb colors
		{"invalid over 255 first", "rgb(256,0,0)", true},
		{"invalid over 255 second", "rgb(0,256,0)", true},
		{"invalid over 255 third", "rgb(0,0,256)", true},
		{"invalid only 2 values", "rgb(0,0)", true},
		{"invalid 4 values", "rgb(0,0,0,0)", true},
		{"invalid negative", "rgb(-1,0,0)", true},
		{"invalid no parens", "rgb 0,0,0", true},
		{"invalid missing closing paren", "rgb(0,0,0", true},
		{"invalid uppercase RGB", "RGB(0,0,0)", true},
		{"invalid decimal", "rgb(0.5,0,0)", true},
		{"invalid text", "rgb(red,green,blue)", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestRgbaConstraint tests rgbaConstraint.Validate() for valid rgba() color formats.
func TestRgbaConstraint(t *testing.T) {
	runSimpleConstraintTests(t, rgbaConstraint{}, []simpleTestCase{
		// Valid rgba colors
		{"valid white opaque", "rgba(255,255,255,1)", false},
		{"valid black transparent", "rgba(0,0,0,0)", false},
		{"valid half transparent", "rgba(255,255,255,0.5)", false},
		{"valid with spaces", "rgba(0, 0, 0, 0.5)", false},
		{"valid alpha 0.1", "rgba(100,100,100,0.1)", false},
		{"valid alpha 0.99", "rgba(100,100,100,0.99)", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid rgba colors
		{"invalid only 3 values (rgb format)", "rgba(0,0,0)", true},
		{"invalid over 255 first", "rgba(256,0,0,1)", true},
		{"invalid over 255 second", "rgba(0,256,0,1)", true},
		{"invalid over 255 third", "rgba(0,0,256,1)", true},
		{"invalid alpha over 1", "rgba(0,0,0,1.5)", true},
		{"invalid alpha negative", "rgba(0,0,0,-0.5)", true},
		{"invalid 5 values", "rgba(0,0,0,1,1)", true},
		{"invalid no parens", "rgba 0,0,0,1", true},
		{"invalid uppercase RGBA", "RGBA(0,0,0,1)", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestHslConstraint tests hslConstraint.Validate() for valid hsl() color formats.
func TestHslConstraint(t *testing.T) {
	runSimpleConstraintTests(t, hslConstraint{}, []simpleTestCase{
		// Valid hsl colors
		{"valid red", "hsl(0,100%,50%)", false},
		{"valid max hue", "hsl(360,100%,50%)", false},
		{"valid black", "hsl(0,0%,0%)", false},
		{"valid white", "hsl(0,0%,100%)", false},
		{"valid with spaces", "hsl(0, 0%, 0%)", false},
		{"valid mid values", "hsl(180,50%,50%)", false},
		{"valid decimal hue", "hsl(180.5,50%,50%)", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid hsl colors
		{"invalid hue over 360", "hsl(361,100%,50%)", true},
		{"invalid hue negative", "hsl(-1,100%,50%)", true},
		{"invalid saturation over 100", "hsl(0,101%,50%)", true},
		{"invalid lightness over 100", "hsl(0,100%,101%)", true},
		{"invalid saturation no percent", "hsl(0,100,50%)", true},
		{"invalid lightness no percent", "hsl(0,100%,50)", true},
		{"invalid only 2 values", "hsl(0,100%)", true},
		{"invalid 4 values", "hsl(0,100%,50%,1)", true},
		{"invalid no parens", "hsl 0,100%,50%", true},
		{"invalid uppercase HSL", "HSL(0,100%,50%)", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestHslaConstraint tests hslaConstraint.Validate() for valid hsla() color formats.
func TestHslaConstraint(t *testing.T) {
	runSimpleConstraintTests(t, hslaConstraint{}, []simpleTestCase{
		// Valid hsla colors
		{"valid red opaque", "hsla(0,100%,50%,1)", false},
		{"valid transparent", "hsla(0,0%,0%,0)", false},
		{"valid half transparent", "hsla(360,100%,50%,0.5)", false},
		{"valid with spaces", "hsla(0, 0%, 0%, 0.5)", false},
		{"valid alpha 0.1", "hsla(180,50%,50%,0.1)", false},
		{"valid alpha 0.99", "hsla(180,50%,50%,0.99)", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid hsla colors
		{"invalid only 3 values (hsl format)", "hsla(0,0%,0%)", true},
		{"invalid hue over 360", "hsla(361,100%,50%,1)", true},
		{"invalid saturation over 100", "hsla(0,101%,50%,1)", true},
		{"invalid lightness over 100", "hsla(0,100%,101%,1)", true},
		{"invalid alpha over 1", "hsla(0,100%,50%,1.5)", true},
		{"invalid alpha negative", "hsla(0,100%,50%,-0.5)", true},
		{"invalid 5 values", "hsla(0,100%,50%,1,1)", true},
		{"invalid no parens", "hsla 0,100%,50%,1", true},
		{"invalid uppercase HSLA", "HSLA(0,100%,50%,1)", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}
