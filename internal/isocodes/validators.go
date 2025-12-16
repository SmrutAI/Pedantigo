package isocodes

import (
	"regexp"
	"sync"

	"golang.org/x/text/language"
)

// Postal code regex caching with lazy initialization.
var (
	postcodeMu        sync.RWMutex
	postcodeRegexDict map[string]*regexp.Regexp
)

// ensurePostcodeRegexes compiles postal code patterns on first use.
// Uses double-checked locking for thread safety.
func ensurePostcodeRegexes() {
	// Fast path: check if already initialized
	postcodeMu.RLock()
	if postcodeRegexDict != nil {
		// Already initialized, fast path
		postcodeMu.RUnlock()
		return
	}
	postcodeMu.RUnlock()

	// SLOW PATH: need to initialize
	postcodeMu.Lock()
	defer postcodeMu.Unlock()

	// Double-check after acquiring write lock
	if postcodeRegexDict != nil {
		// Another goroutine initialized while we waited
		return
	}

	// Compile all patterns
	postcodeRegexDict = make(map[string]*regexp.Regexp, len(postCodePatternDict))
	for countryCode, pattern := range postCodePatternDict {
		postcodeRegexDict[countryCode] = regexp.MustCompile(pattern)
	}
}

// Country code validation (O(1) map lookups - no initialization needed).

// IsISO3166Alpha2 checks if the string is a valid ISO 3166-1 alpha-2 country code.
func IsISO3166Alpha2(code string) bool {
	_, ok := iso3166_1_alpha2[code]
	return ok
}

// IsISO3166Alpha2EU checks if the string is a valid EU ISO 3166-1 alpha-2 country code.
func IsISO3166Alpha2EU(code string) bool {
	_, ok := iso3166_1_alpha2_eu[code]
	return ok
}

// IsISO3166Alpha3 checks if the string is a valid ISO 3166-1 alpha-3 country code.
func IsISO3166Alpha3(code string) bool {
	_, ok := iso3166_1_alpha3[code]
	return ok
}

// IsISO3166Alpha3EU checks if the string is a valid EU ISO 3166-1 alpha-3 country code.
func IsISO3166Alpha3EU(code string) bool {
	_, ok := iso3166_1_alpha3_eu[code]
	return ok
}

// IsISO3166Numeric checks if the int is a valid ISO 3166-1 numeric country code.
func IsISO3166Numeric(code int) bool {
	_, ok := iso3166_1_alpha_numeric[code]
	return ok
}

// IsISO3166NumericEU checks if the int is a valid EU ISO 3166-1 numeric country code.
func IsISO3166NumericEU(code int) bool {
	_, ok := iso3166_1_alpha_numeric_eu[code]
	return ok
}

// IsISO31662 checks if the string is a valid ISO 3166-2 subdivision code.
func IsISO31662(code string) bool {
	_, ok := iso3166_2[code]
	return ok
}

// Currency code validation (O(1) map lookups - no initialization needed).

// IsISO4217 checks if the string is a valid ISO 4217 currency code.
func IsISO4217(code string) bool {
	_, ok := iso4217[code]
	return ok
}

// IsISO4217Numeric checks if the int is a valid ISO 4217 numeric currency code.
func IsISO4217Numeric(code int) bool {
	_, ok := iso4217_numeric[code]
	return ok
}

// Postal code validation (lazy initialization on first use).

// IsPostcode checks if the string is a valid postal code for the given country.
// Country must be an ISO 3166-1 alpha-2 code (e.g., "US", "GB", "DE").
// Returns false if the country is not supported.
func IsPostcode(postcode, countryCode string) bool {
	ensurePostcodeRegexes()

	postcodeMu.RLock()
	regex, ok := postcodeRegexDict[countryCode]
	postcodeMu.RUnlock()

	if !ok {
		return false
	}
	return regex.MatchString(postcode)
}

// HasPostcodePattern checks if a country code has a postal code validation pattern.
// This does NOT trigger regex compilation.
func HasPostcodePattern(countryCode string) bool {
	_, ok := postCodePatternDict[countryCode]
	return ok
}

// IsBCP47LanguageTag validates a BCP 47 language tag using Go's x/text/language parser.
// The parser supports the full IANA language tag registry.
// Examples of valid tags: "en", "en-US", "zh-Hans-CN", "sr-Latn-RS".
func IsBCP47LanguageTag(tag string) bool {
	_, err := language.Parse(tag)
	return err == nil
}
