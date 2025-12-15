package constraints

import "testing"

// TestHtmlConstraint tests htmlConstraint.Validate() for strings containing HTML tags.
func TestHtmlConstraint(t *testing.T) {
	runSimpleConstraintTests(t, htmlConstraint{}, []simpleTestCase{
		// Valid HTML (contains tags)
		{"valid div tag", "<div>", false},
		{"valid p tag", "<p>", false},
		{"valid span with text", "<span>text</span>", false},
		{"valid self-closing br", "<br/>", false},
		{"valid self-closing br space", "<br />", false},
		{"valid img tag", "<img src=\"test.jpg\">", false},
		{"valid a tag with href", "<a href=\"http://example.com\">link</a>", false},
		{"valid nested tags", "<div><p>nested</p></div>", false},
		{"valid with attributes", "<div class=\"test\" id=\"main\">content</div>", false},
		{"valid doctype", "<!DOCTYPE html>", false},
		{"valid comment", "<!-- comment -->", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid HTML (no tags)
		{"invalid no tags", "no tags", true},
		{"invalid empty angle brackets", "<>", true},
		{"invalid just text", "plain text without tags", true},
		{"invalid just ampersand entity", "&nbsp;", true},
		{"invalid escaped lt", "&lt;div&gt;", true},
		{"invalid unclosed tag", "<div", true},
		{"invalid only less than", "<", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestCronConstraint tests cronConstraint.Validate() for valid cron expressions (5 fields).
func TestCronConstraint(t *testing.T) {
	runSimpleConstraintTests(t, cronConstraint{}, []simpleTestCase{
		// Valid cron expressions (5 fields: minute hour day month weekday)
		{"valid all wildcards", "* * * * *", false},
		{"valid midnight daily", "0 0 * * *", false},
		{"valid every 5 mins", "*/5 * * * *", false},
		{"valid specific time", "30 9 * * *", false},
		{"valid ranges", "0-30 * * * *", false},
		{"valid lists", "0,15,30,45 * * * *", false},
		{"valid step", "0 */2 * * *", false},
		{"valid complex", "*/15 9-17 * * 1-5", false},
		{"valid specific day", "0 0 1 * *", false},
		{"valid specific month", "0 0 1 1 *", false},
		{"valid weekday specific", "0 0 * * 0", false},
		{"valid weekday sunday", "0 0 * * SUN", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid cron expressions
		{"invalid only 3 fields", "* * *", true},
		{"invalid only 4 fields", "* * * *", true},
		{"invalid 6 fields", "* * * * * *", true},
		{"invalid text", "invalid", true},
		{"invalid empty", "     ", true},
		{"invalid minute > 59", "60 * * * *", true},
		{"invalid hour > 23", "* 24 * * *", true},
		{"invalid day > 31", "* * 32 * *", true},
		{"invalid month > 12", "* * * 13 *", true},
		{"invalid weekday > 7", "* * * * 8", true},
		{"invalid letters in wrong place", "a b c d e", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestSemverConstraint tests semverConstraint.Validate() for valid semantic versions (X.Y.Z).
func TestSemverConstraint(t *testing.T) {
	runSimpleConstraintTests(t, semverConstraint{}, []simpleTestCase{
		// Valid semver
		{"valid basic", "1.0.0", false},
		{"valid with prerelease", "1.2.3-beta", false},
		{"valid with prerelease dot", "1.2.3-beta.1", false},
		{"valid with build", "1.0.0+build", false},
		{"valid with build metadata", "1.0.0+build.123", false},
		{"valid with prerelease and build", "1.0.0-alpha+build", false},
		{"valid with prerelease dot and build", "1.0.0-alpha.1+build.123", false},
		{"valid zero major", "0.0.0", false},
		{"valid zero minor", "1.0.3", false},
		{"valid large numbers", "100.200.300", false},
		{"valid rc prerelease", "1.0.0-rc.1", false},
		{"valid alpha prerelease", "2.0.0-alpha", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid semver
		{"invalid only major minor", "1.0", true},
		{"invalid with v prefix", "v1.0.0", true},
		{"invalid four parts", "1.0.0.0", true},
		{"invalid single number", "1", true},
		{"invalid text only", "version", true},
		{"invalid leading v lowercase", "v1.2.3", true},
		{"invalid leading V uppercase", "V1.2.3", true},
		{"invalid with spaces", "1.0.0 beta", true},
		{"invalid negative major", "-1.0.0", true},
		{"invalid leading zeros major", "01.0.0", true},
		{"invalid leading zeros minor", "1.00.0", true},
		{"invalid leading zeros patch", "1.0.00", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestUlidConstraint tests ulidConstraint.Validate() for valid ULID format (26 char Crockford base32).
func TestUlidConstraint(t *testing.T) {
	runSimpleConstraintTests(t, ulidConstraint{}, []simpleTestCase{
		// Valid ULIDs (26 characters, Crockford base32)
		{"valid ulid uppercase", "01ARZ3NDEKTSV4RRFFQ69G5FAV", false},
		{"valid ulid lowercase", "01arz3ndektsv4rrffq69g5fav", false},
		{"valid ulid mixed case", "01Arz3NdeKtsV4rrFfq69G5fAv", false},
		{"valid ulid all zeros", "00000000000000000000000000", false},
		{"valid ulid max timestamp", "7ZZZZZZZZZZZZZZZZZZZZZZZZZ", false},
		{"valid ulid example 2", "01BX5ZZKBKACTAV9WEVGEMMVRY", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid ULIDs
		{"invalid too short", "too-short", true},
		{"invalid 25 chars", "01ARZ3NDEKTSV4RRFFQ69G5FA", true},
		{"invalid 27 chars", "01ARZ3NDEKTSV4RRFFQ69G5FAVX", true},
		{"invalid with hyphen", "01ARZ3ND-EKTSV4RR-FFQ69G5F", true},
		{"invalid with spaces", "01ARZ3NDEKTSV4RR FFQ69G5F", true},
		{"invalid chars I", "01ARZ3NDIKTSV4RRFFQ69G5FAV", true}, // I is not in Crockford base32
		{"invalid chars L", "01ARZ3NDLKTSV4RRFFQ69G5FAV", true}, // L is not in Crockford base32
		{"invalid chars O", "01ARZ3NDOKTSV4RRFFQ69G5FAV", true}, // O is not in Crockford base32
		{"invalid chars U", "01ARZ3NDUKTSV4RRFFQ69G5FAV", true}, // U is not in Crockford base32
		{"invalid special char", "01ARZ3NDEKTSV4RRFFQ69G5FA!", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}
