package constraints

import "testing"

// TestISBNConstraint tests isbnConstraint.Validate() for valid ISBN-10 or ISBN-13.
func TestISBNConstraint(t *testing.T) {
	runSimpleConstraintTests(t, isbnConstraint{}, []simpleTestCase{
		// Valid ISBN-10 cases
		{"valid ISBN-10 with dashes", "0-306-40615-2", false},
		{"valid ISBN-10 no dashes", "0306406152", false},
		{"valid ISBN-10 with X checksum", "0-19-853453-1", false},
		// Valid ISBN-13 cases
		{"valid ISBN-13 with dashes", "978-0-306-40615-7", false},
		{"valid ISBN-13 no dashes", "9780306406157", false},
		// Empty string - should skip validation
		{"empty string", "", false},
		// Invalid cases
		{"invalid ISBN-10 bad checksum", "0-306-40615-3", true},
		{"invalid ISBN-13 bad checksum", "978-0-306-40615-8", true},
		{"too short", "123", true},
		{"too long", "978-0-306-40615-7-9999", true},
		{"letters in middle", "978-abc-40615-7", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestISBN10Constraint tests isbn10Constraint.Validate() for valid 10-digit ISBN.
func TestISBN10Constraint(t *testing.T) {
	runSimpleConstraintTests(t, isbn10Constraint{}, []simpleTestCase{
		// Valid ISBN-10 cases
		{"valid ISBN-10 with dashes", "0-306-40615-2", false},
		{"valid ISBN-10 no dashes", "0306406152", false},
		{"valid ISBN-10 oxford", "0-19-853453-1", false},
		{"valid ISBN-10 with X checksum", "0-8044-2957-X", false},
		{"valid ISBN-10 X checksum no dashes", "080442957X", false},
		// Empty string - should skip validation
		{"empty string", "", false},
		// Invalid cases - ISBN-13 format should fail
		{"invalid - ISBN-13 format", "978-0-306-40615-7", true},
		{"invalid - ISBN-13 no dashes", "9780306406157", true},
		// Invalid cases - bad checksum
		{"invalid checksum", "0-306-40615-3", true},
		{"invalid checksum no dashes", "0306406153", true},
		// Invalid cases - format issues
		{"too short", "12345", true},
		{"too long", "0-306-40615-2-0", true},
		{"letters in middle", "0-abc-40615-2", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestISBN13Constraint tests isbn13Constraint.Validate() for valid 13-digit ISBN (EAN).
func TestISBN13Constraint(t *testing.T) {
	runSimpleConstraintTests(t, isbn13Constraint{}, []simpleTestCase{
		// Valid ISBN-13 cases
		{"valid ISBN-13 with dashes", "978-0-306-40615-7", false},
		{"valid ISBN-13 no dashes", "9780306406157", false},
		{"valid ISBN-13 alternative prefix", "979-10-90636-07-1", false},
		// Empty string - should skip validation
		{"empty string", "", false},
		// Invalid cases - ISBN-10 format should fail
		{"invalid - ISBN-10 format", "0-306-40615-2", true},
		{"invalid - ISBN-10 no dashes", "0306406152", true},
		// Invalid cases - bad checksum
		{"invalid checksum", "978-0-306-40615-8", true},
		{"invalid checksum no dashes", "9780306406158", true},
		// Invalid cases - format issues
		{"too short", "978-0-306", true},
		{"too long", "978-0-306-40615-7-999", true},
		{"letters in middle", "978-abc-40615-7", true},
		{"does not start with 978 or 979", "123-0-306-40615-7", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestISSNConstraint tests issnConstraint.Validate() for valid 8-digit ISSN.
func TestISSNConstraint(t *testing.T) {
	runSimpleConstraintTests(t, issnConstraint{}, []simpleTestCase{
		// Valid ISSN cases
		{"valid ISSN with dash", "0378-5955", false},
		{"valid ISSN no dash", "03785955", false},
		{"valid ISSN alternative", "2049-3630", false},
		{"valid ISSN with X checksum", "0317-8471", false},
		// Empty string - should skip validation
		{"empty string", "", false},
		// Invalid cases - bad checksum
		{"invalid checksum", "0378-5956", true},
		{"invalid checksum no dash", "03785956", true},
		// Invalid cases - format issues
		{"too short", "1234", true},
		{"too long", "0378-5955-9", true},
		{"letters in middle", "03a8-5955", true},
		{"wrong dash position", "03785-955", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestSSNConstraint tests ssnConstraint.Validate() for valid U.S. SSN format.
func TestSSNConstraint(t *testing.T) {
	runSimpleConstraintTests(t, ssnConstraint{}, []simpleTestCase{
		// Valid SSN cases
		{"valid SSN standard", "123-45-6789", false},
		{"valid SSN all zeros", "000-00-0000", false},
		{"valid SSN all nines", "999-99-9999", false},
		// Empty string - should skip validation
		{"empty string", "", false},
		// Invalid cases - wrong format
		{"missing dashes", "123456789", true},
		{"wrong dash position", "12345-6789", true},
		{"wrong dash position 2", "123-456-789", true},
		{"extra dash", "123-45-67-89", true},
		// Invalid cases - wrong length
		{"too short", "123-45-678", true},
		{"too long", "123-45-67890", true},
		// Invalid cases - non-digits
		{"letters in area", "abc-45-6789", true},
		{"letters in group", "123-ab-6789", true},
		{"letters in serial", "123-45-abcd", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestEINConstraint tests einConstraint.Validate() for valid U.S. EIN format.
func TestEINConstraint(t *testing.T) {
	runSimpleConstraintTests(t, einConstraint{}, []simpleTestCase{
		// Valid EIN cases
		{"valid EIN standard", "12-3456789", false},
		{"valid EIN all zeros", "00-0000000", false},
		{"valid EIN all nines", "99-9999999", false},
		// Empty string - should skip validation
		{"empty string", "", false},
		// Invalid cases - wrong format
		{"missing dash", "123456789", true},
		{"wrong dash position", "123-456789", true},
		{"wrong dash position 2", "1-23456789", true},
		{"extra dash", "12-345-6789", true},
		// Invalid cases - wrong length
		{"too short", "12-345678", true},
		{"too long", "12-34567890", true},
		// Invalid cases - non-digits
		{"letters in prefix", "ab-3456789", true},
		{"letters in suffix", "12-abcdefg", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestE164Constraint tests e164Constraint.Validate() for valid E.164 phone format.
func TestE164Constraint(t *testing.T) {
	runSimpleConstraintTests(t, e164Constraint{}, []simpleTestCase{
		// Valid E.164 cases
		{"valid E.164 US number", "+14155552671", false},
		{"valid E.164 UK number", "+442071838750", false},
		{"valid E.164 short", "+1", false},
		{"valid E.164 minimum", "+12", false},
		{"valid E.164 maximum 15 digits", "+123456789012345", false},
		// Empty string - should skip validation
		{"empty string", "", false},
		// Invalid cases - missing plus
		{"missing plus sign", "14155552671", true},
		// Invalid cases - starts with zero after plus
		{"starts with zero", "+0123456789", true},
		// Invalid cases - double plus
		{"double plus", "++14155552671", true},
		// Invalid cases - wrong characters
		{"contains letters", "+1415abc2671", true},
		{"contains dash", "+1-415-555-2671", true},
		{"contains space", "+1 415 555 2671", true},
		{"contains parens", "+1(415)5552671", true},
		// Invalid cases - too long
		{"too long 16 digits", "+1234567890123456", true},
		// Invalid cases - only plus
		{"only plus sign", "+", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}
