package isocodes

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsISO3166Alpha2(t *testing.T) {
	assert.True(t, IsISO3166Alpha2("US"))
	assert.True(t, IsISO3166Alpha2("GB"))
	assert.False(t, IsISO3166Alpha2("us"))
	assert.False(t, IsISO3166Alpha2("XX"))
	assert.False(t, IsISO3166Alpha2(""))
}

func TestIsISO3166Alpha2EU(t *testing.T) {
	assert.True(t, IsISO3166Alpha2EU("DE"))
	assert.True(t, IsISO3166Alpha2EU("FR"))
	assert.False(t, IsISO3166Alpha2EU("US"))
	assert.False(t, IsISO3166Alpha2EU("GB")) // no longer an EU member
}

func TestIsISO3166Alpha3(t *testing.T) {
	assert.True(t, IsISO3166Alpha3("USA"))
	assert.True(t, IsISO3166Alpha3("GBR"))
	assert.False(t, IsISO3166Alpha3("usa"))
	assert.False(t, IsISO3166Alpha3("XXX"))
}

func TestIsISO3166Alpha3EU(t *testing.T) {
	assert.True(t, IsISO3166Alpha3EU("DEU"))
	assert.False(t, IsISO3166Alpha3EU("USA"))
}

func TestIsISO3166Numeric(t *testing.T) {
	assert.True(t, IsISO3166Numeric(840)) // USA
	assert.True(t, IsISO3166Numeric(826)) // GBR
	assert.False(t, IsISO3166Numeric(0))
	assert.False(t, IsISO3166Numeric(999))
}

func TestIsISO3166NumericEU(t *testing.T) {
	assert.True(t, IsISO3166NumericEU(276)) // Germany
	assert.False(t, IsISO3166NumericEU(840))
}

func TestIsISO31662(t *testing.T) {
	assert.True(t, IsISO31662("US-CA"))
	assert.True(t, IsISO31662("GB-ENG"))
	assert.False(t, IsISO31662("XX-XX"))
	assert.False(t, IsISO31662("US"))
}

func TestIsISO4217(t *testing.T) {
	assert.True(t, IsISO4217("USD"))
	assert.True(t, IsISO4217("EUR"))
	assert.False(t, IsISO4217("usd"))
	assert.False(t, IsISO4217("ZZZ"))
}

func TestIsISO4217Numeric(t *testing.T) {
	assert.True(t, IsISO4217Numeric(840)) // USD
	assert.False(t, IsISO4217Numeric(0))
	assert.False(t, IsISO4217Numeric(1))
}

func TestHasPostcodePattern(t *testing.T) {
	assert.True(t, HasPostcodePattern("US"))
	assert.True(t, HasPostcodePattern("GB"))
	assert.False(t, HasPostcodePattern("ZZ"))
}

func TestIsPostcode(t *testing.T) {
	assert.True(t, IsPostcode("12345", "US"))
	assert.True(t, IsPostcode("12345-6789", "US"))
	assert.False(t, IsPostcode("1234", "US"))
	assert.True(t, IsPostcode("SW1A 1AA", "GB"))
	assert.False(t, IsPostcode("12345", "GB"))
	// Unsupported country: lookup miss, not a regex match failure.
	assert.False(t, IsPostcode("12345", "ZZ"))
}

// TestIsPostcodeConcurrent exercises the double-checked locking in
// ensurePostcodeRegexes under concurrent first-use, with -race enabled.
func TestIsPostcodeConcurrent(t *testing.T) {
	postcodeMu.Lock()
	postcodeRegexDict = nil // force re-initialization for this test
	postcodeMu.Unlock()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.True(t, IsPostcode("12345", "US"))
		}()
	}
	wg.Wait()
}

func TestIsBCP47LanguageTag(t *testing.T) {
	assert.True(t, IsBCP47LanguageTag("en"))
	assert.True(t, IsBCP47LanguageTag("en-US"))
	assert.True(t, IsBCP47LanguageTag("zh-Hans-CN"))
	assert.False(t, IsBCP47LanguageTag("xyz"))
	assert.False(t, IsBCP47LanguageTag("en@US"))
}
