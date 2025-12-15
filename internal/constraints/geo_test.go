package constraints

import "testing"

// TestLatitudeConstraint tests latitudeConstraint.Validate() for valid latitude values (-90 to +90).
func TestLatitudeConstraint(t *testing.T) {
	runSimpleConstraintTests(t, latitudeConstraint{}, []simpleTestCase{
		// Valid latitudes
		{"valid zero", float64(0), false},
		{"valid positive", float64(45.5), false},
		{"valid max", float64(90), false},
		{"valid min", float64(-90), false},
		{"valid positive decimal", float64(51.5074), false},
		{"valid negative decimal", float64(-33.8688), false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid latitudes - out of range
		{"invalid over max", float64(91), true},
		{"invalid under min", float64(-91), true},
		{"invalid way over max", float64(180), true},
		{"invalid way under min", float64(-180), true},
		// Integer types (should work via toFloat64 conversion)
		{"valid int zero", int(0), false},
		{"valid int max", int(90), false},
		{"valid int min", int(-90), false},
		{"invalid int over max", int(91), true},
		// Nil pointer - should skip validation
		{"nil pointer float64", (*float64)(nil), false},
		{"nil pointer int", (*int)(nil), false},
		// Invalid types
		{"invalid type - string", "45.5", true},
		{"invalid type - bool", true, true},
		{"invalid type - slice", []float64{45.5}, true},
	})
}

// TestLongitudeConstraint tests longitudeConstraint.Validate() for valid longitude values (-180 to +180).
func TestLongitudeConstraint(t *testing.T) {
	runSimpleConstraintTests(t, longitudeConstraint{}, []simpleTestCase{
		// Valid longitudes
		{"valid zero", float64(0), false},
		{"valid positive", float64(90.5), false},
		{"valid max", float64(180), false},
		{"valid min", float64(-180), false},
		{"valid positive decimal", float64(0.1278), false},
		{"valid negative decimal", float64(-122.4194), false},
		{"valid 45 degrees", float64(45), false},
		{"valid -45 degrees", float64(-45), false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid longitudes - out of range
		{"invalid over max", float64(181), true},
		{"invalid under min", float64(-181), true},
		{"invalid way over max", float64(360), true},
		{"invalid way under min", float64(-360), true},
		// Integer types (should work via toFloat64 conversion)
		{"valid int zero", int(0), false},
		{"valid int max", int(180), false},
		{"valid int min", int(-180), false},
		{"invalid int over max", int(181), true},
		// Nil pointer - should skip validation
		{"nil pointer float64", (*float64)(nil), false},
		{"nil pointer int", (*int)(nil), false},
		// Invalid types
		{"invalid type - string", "90.5", true},
		{"invalid type - bool", true, true},
		{"invalid type - slice", []float64{90.5}, true},
	})
}
