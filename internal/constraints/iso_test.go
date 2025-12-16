package constraints

import "testing"

// TestISO3166Alpha2Constraint tests iso3166Alpha2Constraint.Validate() for valid ISO 3166-1 alpha-2 codes.
func TestISO3166Alpha2Constraint(t *testing.T) {
	runSimpleConstraintTests(t, iso3166Alpha2Constraint{}, []simpleTestCase{
		// Valid ISO 3166-1 alpha-2 country codes
		{"valid US", "US", false},
		{"valid GB", "GB", false},
		{"valid DE", "DE", false},
		{"valid JP", "JP", false},
		{"valid FR", "FR", false},
		{"valid AU", "AU", false},
		{"valid CA", "CA", false},
		{"valid CN", "CN", false},
		{"valid IN", "IN", false},
		{"valid BR", "BR", false},
		// Empty string - should be skipped (handled by required)
		{"empty string", "", false},
		// Invalid codes
		{"invalid lowercase us", "us", true},
		{"invalid lowercase gb", "gb", true},
		{"invalid XX nonexistent", "XX", true},
		{"invalid 3 chars USA", "USA", true},
		{"invalid 1 char U", "U", true},
		{"invalid numeric 01", "01", true},
		{"invalid with space", "U S", true},
		{"invalid with hyphen", "U-S", true},
		{"invalid special chars", "U$", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
		{"invalid type - float", 12.34, true},
	})
}

// TestISO3166Alpha2EUConstraint tests iso3166Alpha2EUConstraint.Validate() for valid EU country codes.
func TestISO3166Alpha2EUConstraint(t *testing.T) {
	runSimpleConstraintTests(t, iso3166Alpha2EUConstraint{}, []simpleTestCase{
		// Valid EU ISO 3166-1 alpha-2 country codes (27 current EU members)
		{"valid DE Germany", "DE", false},
		{"valid FR France", "FR", false},
		{"valid IT Italy", "IT", false},
		{"valid ES Spain", "ES", false},
		{"valid NL Netherlands", "NL", false},
		{"valid BE Belgium", "BE", false},
		{"valid AT Austria", "AT", false},
		{"valid PL Poland", "PL", false},
		{"valid SE Sweden", "SE", false},
		{"valid DK Denmark", "DK", false},
		{"valid FI Finland", "FI", false},
		{"valid IE Ireland", "IE", false},
		{"valid PT Portugal", "PT", false},
		{"valid GR Greece", "GR", false},
		{"valid CZ Czechia", "CZ", false},
		{"valid RO Romania", "RO", false},
		{"valid HU Hungary", "HU", false},
		{"valid BG Bulgaria", "BG", false},
		{"valid HR Croatia", "HR", false},
		{"valid SK Slovakia", "SK", false},
		{"valid LT Lithuania", "LT", false},
		{"valid SI Slovenia", "SI", false},
		{"valid LV Latvia", "LV", false},
		{"valid EE Estonia", "EE", false},
		{"valid CY Cyprus", "CY", false},
		{"valid LU Luxembourg", "LU", false},
		{"valid MT Malta", "MT", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid - not EU countries
		{"invalid US not EU", "US", true},
		{"invalid GB not EU anymore", "GB", true},
		{"invalid CH not EU", "CH", true},
		{"invalid NO not EU", "NO", true},
		{"invalid IS not EU", "IS", true},
		{"invalid JP not EU", "JP", true},
		{"invalid AU not EU", "AU", true},
		{"invalid XX nonexistent", "XX", true},
		{"invalid lowercase de", "de", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
	})
}

// TestISO3166Alpha3Constraint tests iso3166Alpha3Constraint.Validate() for valid alpha-3 codes.
func TestISO3166Alpha3Constraint(t *testing.T) {
	runSimpleConstraintTests(t, iso3166Alpha3Constraint{}, []simpleTestCase{
		// Valid ISO 3166-1 alpha-3 country codes
		{"valid USA", "USA", false},
		{"valid GBR", "GBR", false},
		{"valid DEU", "DEU", false},
		{"valid JPN", "JPN", false},
		{"valid FRA", "FRA", false},
		{"valid AUS", "AUS", false},
		{"valid CAN", "CAN", false},
		{"valid CHN", "CHN", false},
		{"valid IND", "IND", false},
		{"valid BRA", "BRA", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid codes
		{"invalid lowercase usa", "usa", true},
		{"invalid mixed case Usa", "Usa", true},
		{"invalid 2 chars US", "US", true},
		{"invalid XXX nonexistent", "XXX", true},
		{"invalid 4 chars USAA", "USAA", true},
		{"invalid 1 char U", "U", true},
		{"invalid with hyphen", "US-A", true},
		{"invalid with space", "USA ", true},
		{"invalid numeric", "123", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestISO3166Alpha3EUConstraint tests iso3166Alpha3EUConstraint.Validate() for valid EU alpha-3 codes.
func TestISO3166Alpha3EUConstraint(t *testing.T) {
	runSimpleConstraintTests(t, iso3166Alpha3EUConstraint{}, []simpleTestCase{
		// Valid EU ISO 3166-1 alpha-3 country codes
		{"valid DEU Germany", "DEU", false},
		{"valid FRA France", "FRA", false},
		{"valid ITA Italy", "ITA", false},
		{"valid ESP Spain", "ESP", false},
		{"valid NLD Netherlands", "NLD", false},
		{"valid BEL Belgium", "BEL", false},
		{"valid AUT Austria", "AUT", false},
		{"valid POL Poland", "POL", false},
		{"valid SWE Sweden", "SWE", false},
		{"valid DNK Denmark", "DNK", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid - not EU countries
		{"invalid USA not EU", "USA", true},
		{"invalid GBR not EU", "GBR", true},
		{"invalid CHE not EU", "CHE", true},
		{"invalid NOR not EU", "NOR", true},
		{"invalid JPN not EU", "JPN", true},
		{"invalid lowercase deu", "deu", true},
		{"invalid XXX nonexistent", "XXX", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
	})
}

// TestISO3166NumericConstraint tests iso3166NumericConstraint.Validate() for valid numeric codes.
func TestISO3166NumericConstraint(t *testing.T) {
	runSimpleConstraintTests(t, iso3166NumericConstraint{}, []simpleTestCase{
		// Valid ISO 3166-1 numeric country codes
		{"valid 840 USA", 840, false},
		{"valid 826 GBR", 826, false},
		{"valid 276 DEU", 276, false},
		{"valid 392 JPN", 392, false},
		{"valid 250 FRA", 250, false},
		{"valid 036 AUS", 36, false}, // Leading zeros optional for int
		{"valid 124 CAN", 124, false},
		{"valid 156 CHN", 156, false},
		{"valid 356 IND", 356, false},
		{"valid 076 BRA", 76, false},
		{"valid 004 AFG", 4, false}, // Afghanistan - single digit when leading zeros stripped
		// Invalid codes
		{"invalid 0", 0, true},
		{"invalid 999", 999, true},
		{"invalid 1000", 1000, true},
		{"invalid 9999", 9999, true},
		{"invalid negative", -1, true},
		{"invalid large number", 100000, true},
		// String input should fail for numeric constraint
		{"invalid string 840", "840", true},
		{"invalid string USA", "USA", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*int)(nil), false},
		// Invalid types
		{"invalid type - string", "840", true},
		{"invalid type - bool", true, true},
		{"invalid type - float", 840.5, true},
	})
}

// TestISO31662Constraint tests iso31662Constraint.Validate() for valid subdivision codes.
func TestISO31662Constraint(t *testing.T) {
	runSimpleConstraintTests(t, iso31662Constraint{}, []simpleTestCase{
		// Valid ISO 3166-2 subdivision codes (US states)
		{"valid US-CA California", "US-CA", false},
		{"valid US-NY New York", "US-NY", false},
		{"valid US-TX Texas", "US-TX", false},
		{"valid US-FL Florida", "US-FL", false},
		{"valid US-WA Washington", "US-WA", false},
		// Valid UK subdivisions
		{"valid GB-ENG England", "GB-ENG", false},
		{"valid GB-SCT Scotland", "GB-SCT", false},
		{"valid GB-WLS Wales", "GB-WLS", false},
		{"valid GB-NIR Northern Ireland", "GB-NIR", false},
		// Valid German states
		{"valid DE-BY Bavaria", "DE-BY", false},
		{"valid DE-BE Berlin", "DE-BE", false},
		{"valid DE-HH Hamburg", "DE-HH", false},
		{"valid DE-NW North Rhine-Westphalia", "DE-NW", false},
		// Valid French regions
		{"valid FR-75 Paris", "FR-75", false},
		{"valid FR-13 Marseille", "FR-13", false},
		{"valid FR-69 Lyon", "FR-69", false},
		// Valid Canadian provinces
		{"valid CA-ON Ontario", "CA-ON", false},
		{"valid CA-QC Quebec", "CA-QC", false},
		{"valid CA-BC British Columbia", "CA-BC", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid codes
		{"invalid missing hyphen USCA", "USCA", true},
		{"invalid only country US", "US", true},
		{"invalid only subdivision CA", "CA", true},
		{"invalid XX-XX nonexistent", "XX-XX", true},
		{"invalid lowercase us-ca", "us-ca", true},
		{"invalid mixed case Us-Ca", "Us-Ca", true},
		{"invalid wrong separator US_CA", "US_CA", true},
		{"invalid double hyphen US--CA", "US--CA", true},
		{"invalid space US CA", "US CA", true},
		{"invalid too short U-C", "U-C", true},
		{"invalid numeric only 01-23", "01-23", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestISO4217Constraint tests iso4217Constraint.Validate() for valid currency codes.
func TestISO4217Constraint(t *testing.T) {
	runSimpleConstraintTests(t, iso4217Constraint{}, []simpleTestCase{
		// Valid ISO 4217 currency codes
		{"valid USD", "USD", false},
		{"valid EUR", "EUR", false},
		{"valid GBP", "GBP", false},
		{"valid JPY", "JPY", false},
		{"valid CNY", "CNY", false},
		{"valid CHF", "CHF", false},
		{"valid AUD", "AUD", false},
		{"valid CAD", "CAD", false},
		{"valid INR", "INR", false},
		{"valid BRL", "BRL", false},
		{"valid RUB", "RUB", false},
		{"valid KRW", "KRW", false},
		{"valid MXN", "MXN", false},
		{"valid SEK", "SEK", false},
		{"valid NOK", "NOK", false},
		{"valid XXX no currency", "XXX", false}, // Special code for "no currency"
		{"valid XTS testing", "XTS", false},     // Testing code
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid codes
		{"invalid lowercase usd", "usd", true},
		{"invalid mixed case Usd", "Usd", true},
		{"invalid ZZZ nonexistent", "ZZZ", true},
		{"invalid 2 chars US", "US", true},
		{"invalid 4 chars USDX", "USDX", true},
		{"invalid 1 char U", "U", true},
		{"invalid numeric 123", "123", true},
		{"invalid with hyphen", "US-D", true},
		{"invalid with space", "USD ", true},
		{"invalid special chars", "U$D", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 840, true},
		{"invalid type - bool", true, true},
		{"invalid type - float", 840.5, true},
	})
}

// TestISO4217NumericConstraint tests iso4217NumericConstraint.Validate() for valid numeric currency codes.
func TestISO4217NumericConstraint(t *testing.T) {
	runSimpleConstraintTests(t, iso4217NumericConstraint{}, []simpleTestCase{
		// Valid ISO 4217 numeric currency codes
		{"valid 840 USD", 840, false},
		{"valid 978 EUR", 978, false},
		{"valid 826 GBP", 826, false},
		{"valid 392 JPY", 392, false},
		{"valid 156 CNY", 156, false},
		{"valid 756 CHF", 756, false},
		{"valid 036 AUD", 36, false}, // Leading zeros optional for int
		{"valid 124 CAD", 124, false},
		{"valid 356 INR", 356, false},
		{"valid 986 BRL", 986, false},
		{"valid 643 RUB", 643, false},
		{"valid 410 KRW", 410, false},
		{"valid 484 MXN", 484, false},
		{"valid 752 SEK", 752, false},
		{"valid 578 NOK", 578, false},
		{"valid 999 XXX", 999, false}, // Special code for testing
		// Invalid codes
		{"invalid 0", 0, true},
		{"invalid 1", 1, true},
		{"invalid 1000", 1000, true},
		{"invalid 9999", 9999, true},
		{"invalid negative", -1, true},
		{"invalid large number", 100000, true},
		// String input should fail for numeric constraint
		{"invalid string 840", "840", true},
		{"invalid string USD", "USD", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*int)(nil), false},
		// Invalid types
		{"invalid type - string", "840", true},
		{"invalid type - bool", true, true},
		{"invalid type - float", 840.5, true},
	})
}

// TestPostcodeConstraint tests postcodeConstraint.Validate() for valid postal codes.
func TestPostcodeConstraint(t *testing.T) {
	// Test US postal codes
	t.Run("US postal codes", func(t *testing.T) {
		runSimpleConstraintTests(t, postcodeConstraint{countryCode: "US"}, []simpleTestCase{
			{"valid 5 digit", "12345", false},
			{"valid 5+4 with hyphen", "12345-6789", false},
			{"valid 5+4 with space", "12345 6789", false},
			{"valid 00000", "00000", false},
			{"valid 99999", "99999", false},
			{"valid 90210", "90210", false},
			{"valid 10001 NYC", "10001", false},
			// Empty string - should be skipped
			{"empty string", "", false},
			// Invalid US postal codes
			{"invalid 4 digits", "1234", true},
			{"invalid 6 digits", "123456", true},
			{"invalid letters", "ABCDE", true},
			{"invalid mixed alphanumeric", "1234A", true},
			{"invalid too short zip+4", "12345-678", true},
			{"invalid too long zip+4", "12345-67890", true},
			{"invalid special chars", "12345!", true},
			// Nil pointer - should skip validation
			{"nil pointer", (*string)(nil), false},
			// Invalid types
			{"invalid type - int", 12345, true},
		})
	})

	// Test UK postal codes
	t.Run("UK postal codes", func(t *testing.T) {
		runSimpleConstraintTests(t, postcodeConstraint{countryCode: "GB"}, []simpleTestCase{
			{"valid SW1A 1AA", "SW1A 1AA", false},
			{"valid EC1A 1BB", "EC1A 1BB", false},
			{"valid W1A 0AX", "W1A 0AX", false},
			{"valid M1 1AE", "M1 1AE", false},
			{"valid B33 8TH", "B33 8TH", false},
			{"valid CR2 6XH", "CR2 6XH", false},
			{"valid DN55 1PT", "DN55 1PT", false},
			{"valid GU16 7HF", "GU16 7HF", false},
			{"valid PO16 7GZ", "PO16 7GZ", false},
			// Empty string - should be skipped
			{"empty string", "", false},
			// Invalid UK postal codes
			{"invalid US format", "12345", true},
			{"invalid lowercase", "sw1a 1aa", true},
			{"valid no space", "SW1A1AA", false}, // UK postcodes commonly written without space
			{"invalid wrong format", "1SW1 AA1", true},
			// Nil pointer - should skip validation
			{"nil pointer", (*string)(nil), false},
			// Invalid types
			{"invalid type - int", 12345, true},
		})
	})

	// Test German postal codes
	t.Run("DE postal codes", func(t *testing.T) {
		runSimpleConstraintTests(t, postcodeConstraint{countryCode: "DE"}, []simpleTestCase{
			{"valid 10115 Berlin", "10115", false},
			{"valid 80331 Munich", "80331", false},
			{"valid 20095 Hamburg", "20095", false},
			{"valid 50667 Cologne", "50667", false},
			{"valid 60311 Frankfurt", "60311", false},
			{"valid 70173 Stuttgart", "70173", false},
			{"valid 01067 Dresden", "01067", false},
			// Empty string - should be skipped
			{"empty string", "", false},
			// Invalid German postal codes
			{"invalid 4 digits", "1011", true},
			{"invalid 6 digits", "101155", true},
			{"invalid letters", "ABCDE", true},
			{"invalid with hyphen", "10115-", true},
			// Nil pointer - should skip validation
			{"nil pointer", (*string)(nil), false},
			// Invalid types
			{"invalid type - int", 10115, true},
		})
	})

	// Test French postal codes
	t.Run("FR postal codes", func(t *testing.T) {
		runSimpleConstraintTests(t, postcodeConstraint{countryCode: "FR"}, []simpleTestCase{
			{"valid 75001 Paris", "75001", false},
			{"valid 13001 Marseille", "13001", false},
			{"valid 69001 Lyon", "69001", false},
			{"valid 31000 Toulouse", "31000", false},
			// Empty string - should be skipped
			{"empty string", "", false},
			// Invalid French postal codes
			{"invalid 4 digits", "7500", true},
			{"invalid 6 digits", "750001", true},
			{"invalid letters", "ABCDE", true},
			// Nil pointer - should skip validation
			{"nil pointer", (*string)(nil), false},
		})
	})

	// Test Canadian postal codes
	t.Run("CA postal codes", func(t *testing.T) {
		runSimpleConstraintTests(t, postcodeConstraint{countryCode: "CA"}, []simpleTestCase{
			{"valid K1A 0B1 Ottawa", "K1A 0B1", false},
			{"valid M5H 2N2 Toronto", "M5H 2N2", false},
			{"valid V6B 1A1 Vancouver", "V6B 1A1", false},
			{"valid H3B 1A1 Montreal", "H3B 1A1", false},
			// Empty string - should be skipped
			{"empty string", "", false},
			// Invalid Canadian postal codes
			{"invalid US format", "12345", true},
			{"invalid lowercase", "k1a 0b1", true},
			{"valid no space", "K1A0B1", false}, // CA postcodes commonly written without space
			// Nil pointer - should skip validation
			{"nil pointer", (*string)(nil), false},
		})
	})

	// Test unsupported country
	t.Run("unsupported country", func(t *testing.T) {
		runSimpleConstraintTests(t, postcodeConstraint{countryCode: "ZZ"}, []simpleTestCase{
			{"any value unsupported", "12345", true},
			{"empty string unsupported", "", false}, // Empty should still skip
		})
	})
}

// TestBCP47Constraint tests bcp47Constraint.Validate() for valid BCP 47 language tags.
func TestBCP47Constraint(t *testing.T) {
	runSimpleConstraintTests(t, bcp47Constraint{}, []simpleTestCase{
		// Valid BCP 47 language tags - simple language codes
		{"valid en", "en", false},
		{"valid de", "de", false},
		{"valid fr", "fr", false},
		{"valid es", "es", false},
		{"valid it", "it", false},
		{"valid pt", "pt", false},
		{"valid ja", "ja", false},
		{"valid zh", "zh", false},
		{"valid ko", "ko", false},
		{"valid ar", "ar", false},
		// Valid language + region
		{"valid en-US", "en-US", false},
		{"valid en-GB", "en-GB", false},
		{"valid en-AU", "en-AU", false},
		{"valid de-DE", "de-DE", false},
		{"valid de-AT", "de-AT", false},
		{"valid de-CH", "de-CH", false},
		{"valid fr-FR", "fr-FR", false},
		{"valid fr-CA", "fr-CA", false},
		{"valid es-ES", "es-ES", false},
		{"valid es-MX", "es-MX", false},
		{"valid es-419", "es-419", false}, // Spanish Latin America (numeric region)
		{"valid pt-BR", "pt-BR", false},
		{"valid pt-PT", "pt-PT", false},
		// Valid language + script
		{"valid zh-Hans", "zh-Hans", false}, // Simplified Chinese
		{"valid zh-Hant", "zh-Hant", false}, // Traditional Chinese
		{"valid sr-Latn", "sr-Latn", false}, // Serbian Latin
		{"valid sr-Cyrl", "sr-Cyrl", false}, // Serbian Cyrillic
		{"valid az-Latn", "az-Latn", false}, // Azerbaijani Latin
		{"valid az-Cyrl", "az-Cyrl", false}, // Azerbaijani Cyrillic
		// Valid language + script + region
		{"valid zh-Hans-CN", "zh-Hans-CN", false}, // Simplified Chinese (China)
		{"valid zh-Hant-TW", "zh-Hant-TW", false}, // Traditional Chinese (Taiwan)
		{"valid zh-Hant-HK", "zh-Hant-HK", false}, // Traditional Chinese (Hong Kong)
		{"valid sr-Latn-RS", "sr-Latn-RS", false}, // Serbian Latin (Serbia)
		{"valid sr-Cyrl-RS", "sr-Cyrl-RS", false}, // Serbian Cyrillic (Serbia)
		// Valid with variants
		{"valid de-DE-1996", "de-DE-1996", false},       // German with 1996 spelling reform
		{"valid en-US-x-twain", "en-US-x-twain", false}, // Private use subtag
		// Valid grandfathered tags
		{"valid i-default", "i-default", false},
		{"valid en-GB-oed", "en-GB-oed", false}, // Oxford English Dictionary spelling
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid BCP 47 language tags
		{"invalid xx nonexistent", "xx", true},
		{"invalid xyz nonexistent", "xyz", true},
		{"invalid zz nonexistent", "zz", true},
		{"invalid too long abcdefgh", "abcdefgh", true},
		{"invalid numeric only", "123", true},
		{"invalid special chars at", "en@US", true},
		{"invalid special chars hash", "en#US", true},
		{"invalid with space", "en US", true},
		{"valid underscore normalized", "en_US", false}, // language.Parse() normalizes underscore to hyphen
		{"invalid trailing hyphen", "en-", true},
		{"invalid leading hyphen", "-en", true},
		{"invalid double hyphen", "en--US", true},
		// Note: BCP 47 is case-insensitive - these are all valid and normalized by the parser
		{"valid lowercase region", "en-us", false},   // Normalized to "en-US"
		{"valid uppercase language", "EN-US", false}, // Normalized to "en-US"
		{"valid mixed case", "En-Us", false},         // Normalized to "en-US"
		{"invalid wrong order region-lang", "US-en", true},
		{"invalid single char", "e", true},
		{"invalid empty parts", "en--US", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
		{"invalid type - float", 12.34, true},
	})
}
