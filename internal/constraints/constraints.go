// Package constraints provides validation constraint types and builders for pedantigo.
package constraints

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

// Constraint represents a validation constraint.
type Constraint interface {
	Validate(value any) error
}

// Constraint name constants.
const (
	// Core constraints.
	CMin          = "min"
	CMax          = "max"
	CGt           = "gt"
	CGte          = "gte"
	CLt           = "lt"
	CLte          = "lte"
	CEmail        = "email"
	CUrl          = "url"
	CUri          = "uri"
	CUuid         = "uuid"
	CUuid3        = "uuid3"
	CUuid4        = "uuid4"
	CUuid5        = "uuid5"
	CRegexp       = "regexp"
	CIpv4         = "ipv4"
	CIpv6         = "ipv6"
	COneof        = "oneof"
	COneofci      = "oneofci"
	CEq           = "eq"
	CNe           = "ne"
	CEqIgnoreCase = "eq_ignore_case"
	CNeIgnoreCase = "ne_ignore_case"
	CLen          = "len"

	// String constraints.
	CAscii           = "ascii"
	CAlpha           = "alpha"
	CAlphanum        = "alphanum"
	CAlphaspace      = "alphaspace"
	CAlphanumspace   = "alphanumspace"
	CPrintascii      = "printascii"
	CNumeric         = "numeric"
	CNumber          = "number"
	CHexadecimal     = "hexadecimal"
	CAlphaunicode    = "alphaunicode"
	CAlphanumunicode = "alphanumunicode"
	CContains        = "contains"
	CExcludes        = "excludes"
	CStartswith      = "startswith"
	CEndswith        = "endswith"
	CStartsnotwith   = "startsnotwith"
	CEndsnotwith     = "endsnotwith"
	CContainsany     = "containsany"
	CExcludesall     = "excludesall"
	CExcludesrune    = "excludesrune"
	CContainsRune    = "containsrune"
	CLowercase       = "lowercase"
	CUppercase       = "uppercase"
	CMultibyte       = "multibyte"
	CUrnRfc2141      = "urn_rfc2141"
	CStripWhitespace = "strip_whitespace"
	CToLower         = "to_lower"
	CToUpper         = "to_upper"

	// Numeric constraints.
	CPositive       = "positive"
	CNegative       = "negative"
	CMultipleOf     = "multiple_of"
	CMaxDigits      = "max_digits"
	CDecimalPlaces  = "decimal_places"
	CDisallowInfNan = "disallow_inf_nan"

	// Collection constraints.
	CUnique  = "unique"
	CDefault = "default"

	// Network constraints.
	CIp              = "ip"
	CCidr            = "cidr"
	CCidrv4          = "cidrv4"
	CCidrv6          = "cidrv6"
	CMac             = "mac"
	CHostname        = "hostname"
	CHostnameRfc1123 = "hostname_rfc1123"
	CHostnamePort    = "hostname_port"
	CFqdn            = "fqdn"
	CPort            = "port"
	CTcpAddr         = "tcp_addr"
	CUdpAddr         = "udp_addr"
	CTcp4Addr        = "tcp4_addr"
	CHttpUrl         = "http_url"
	CHttpsUrl        = "https_url"

	// Finance constraints.
	CCreditCard    = "credit_card"
	CBtcAddr       = "btc_addr"
	CBtcAddrBech32 = "btc_addr_bech32"
	CEthAddr       = "eth_addr"
	CLuhnChecksum  = "luhn_checksum"

	// Identity constraints.
	CIsbn   = "isbn"
	CIsbn10 = "isbn10"
	CIsbn13 = "isbn13"
	CIssn   = "issn"
	CSsn    = "ssn"
	CEin    = "ein"
	CE164   = "e164"

	// Geo constraints.
	CLatitude  = "latitude"
	CLongitude = "longitude"

	// Color constraints.
	CHexcolor = "hexcolor"
	CRgb      = "rgb"
	CRgba     = "rgba"
	CHsl      = "hsl"
	CHsla     = "hsla"

	// Encoding constraints.
	CJwt          = "jwt"
	CJson         = "json"
	CBase64       = "base64"
	CBase64url    = "base64url"
	CBase64rawurl = "base64rawurl"
	CDatauri      = "datauri"
	CBase32       = "base32"

	// Hash constraints.
	CMd4     = "md4"
	CMd5     = "md5"
	CSha256  = "sha256"
	CSha384  = "sha384"
	CSha512  = "sha512"
	CMongodb = "mongodb"

	// Misc constraints.
	CHtml     = "html"
	CCron     = "cron"
	CSemver   = "semver"
	CUlid     = "ulid"
	CDatetime = "datetime"
	CTimezone = "timezone"

	// Special.
	CRequired   = "required"
	CSkipUnless = "skip_unless"
	CImage      = "image"
)

// Shared regex patterns used by string constraints.
var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	uuidRegex     = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	alphaRegex    = regexp.MustCompile(`^[a-zA-Z]+$`)
	alphanumRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	numericRegex  = regexp.MustCompile(`^[-+]?\d+(?:\.\d+)?$`)
	// URN regex pattern (RFC 2141): urn:<NID>:<NSS>
	// NID: starts with letter, contains letters/digits/hyphens, max 32 chars
	// NSS: non-empty, no whitespace
	urnRegex = regexp.MustCompile(`(?i)^urn:[a-z][a-z0-9\-]{0,31}:\S+$`)
)

// extractNumericValue converts a reflect.Value to a float64 for numeric comparisons.
// Returns (float64, error) where error is non-nil if the value is not numeric.
func extractNumericValue(v reflect.Value) (float64, error) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return v.Float(), nil
	default:
		return 0, fmt.Errorf("unsupported numeric type: %s", v.Kind())
	}
}

// derefValue dereferences a pointer value, returning the underlying value or nil if invalid.
// Returns (reflect.Value, bool) where bool is false if the value is nil or invalid.
func derefValue(value any) (reflect.Value, bool) {
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return reflect.Value{}, false
	}

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}, false
		}
		v = v.Elem()
	}

	return v, true
}

// extractString extracts a string value from reflect.Value, checking type and dereferencing.
// Returns (string, isValid, error) where isValid is false for nil/invalid values.
func extractString(value any) (str string, isValid bool, err error) {
	v, ok := derefValue(value)
	if !ok {
		return "", false, nil // nil/invalid values should skip validation
	}

	if v.Kind() != reflect.String {
		return "", true, fmt.Errorf("requires string value")
	}

	return v.String(), true, nil
}

// BuildConstraints creates constraint instances from parsed tag map.
func BuildConstraints(constraints map[string]string, fieldType reflect.Type) []Constraint {
	var result []Constraint

	for name, value := range constraints {
		// Handle OR constraints specially
		if len(name) > 6 && name[:6] == "__or__" {
			orExpr := name[6:] // Strip the "__or__" prefix
			if c := buildOrConstraint(orExpr, fieldType); c != nil {
				result = append(result, c)
			}
			continue
		}

		switch name {
		case CRequired:
			// Skip: 'required' is only checked during Unmarshal (missing JSON keys).
			// It doesn't apply to Validate() on manually created structs.
			continue

		// Core constraints.
		case CMin, CMax, CGt, CGte, CLt, CLte, CEmail, CUrl, CUri, CUuid, CUuid3, CUuid4, CUuid5, CRegexp, CIpv4, CIpv6, COneof, COneofci, CEq, CNe, CLen, CHttpUrl:
			result = appendCoreConstraint(result, name, value, fieldType)

		// String constraints.
		case CAscii, CAlpha, CAlphanum, CAlphaspace, CAlphanumspace, CPrintascii, CNumeric, CNumber, CHexadecimal, CAlphaunicode, CAlphanumunicode, CContains, CExcludes, CStartswith, CEndswith, CStartsnotwith, CEndsnotwith, CContainsany, CExcludesall, CExcludesrune, CLowercase, CUppercase, CMultibyte, CUrnRfc2141, CStripWhitespace, CToLower, CToUpper:
			result = appendStringConstraint(result, name, value)

		// Numeric constraints.
		case CPositive, CNegative, CMultipleOf, CMaxDigits, CDecimalPlaces, CDisallowInfNan:
			result = appendNumericConstraint(result, name, value)

		// Collection constraints.
		case CUnique, CDefault:
			result = appendCollectionConstraint(result, name, value)

		// Network constraints.
		case CIp, CCidr, CCidrv4, CCidrv6, CMac, CHostname, CHostnameRfc1123, CHostnamePort, CFqdn, CPort, CTcpAddr, CUdpAddr, CTcp4Addr, CHttpsUrl:
			result = appendNetworkConstraint(result, name)

		// Finance constraints.
		case CCreditCard, CBtcAddr, CBtcAddrBech32, CEthAddr, CLuhnChecksum:
			result = appendFinanceConstraint(result, name)

		// Identity constraints.
		case CIsbn, CIsbn10, CIsbn13, CIssn, CSsn, CEin, CE164:
			result = appendIdentityConstraint(result, name)

		// Geo constraints.
		case CLatitude, CLongitude:
			result = appendGeoConstraint(result, name)

		// Color constraints.
		case CHexcolor, CRgb, CRgba, CHsl, CHsla:
			result = appendColorConstraint(result, name)

		// Encoding constraints.
		case CJwt, CJson, CBase64, CBase64url, CBase64rawurl, CDatauri, CBase32:
			result = appendEncodingConstraint(result, name)

		// Hash constraints.
		case CMd4, CMd5, CSha256, CSha384, CSha512, CMongodb:
			result = appendHashConstraint(result, name)

		// Misc constraints.
		case CHtml, CCron, CSemver, CUlid, CDatetime, CTimezone:
			result = appendMiscConstraint(result, name, value)

		// ISO code constraints.
		case CISO31661Alpha2, CISO3166Alpha2EU, CISO31661Alpha3, CISO3166Alpha3EU, CISO31661AlphaNumeric, CISO31662, CISO4217, CISO4217Numeric, CPostcode, CPostcodeISO3166Alpha2, CBCP47LanguageTag:
			result = appendISOConstraint(result, name, value)

		// Filesystem constraints.
		case CFilepath, CDirpath, CFile, CDir:
			result = appendFilesystemConstraint(result, name)

		default:
			// Check for custom validators
			if c, ok := BuildCustomConstraint(name, value); ok {
				result = append(result, c)
			}
			// Unknown constraints are silently ignored (fail-fast happens at registry level)
		}
	}

	return result
}

// appendCoreConstraint appends core validation constraints if name matches.
func appendCoreConstraint(result []Constraint, name, value string, fieldType reflect.Type) []Constraint {
	switch name {
	case "min":
		if c, ok := buildMinConstraint(value, fieldType); ok {
			return append(result, c)
		}
	case "max":
		if c, ok := buildMaxConstraint(value, fieldType); ok {
			return append(result, c)
		}
	case "gt":
		if threshold, err := strconv.ParseFloat(value, 64); err == nil {
			return append(result, gtConstraint{threshold: threshold})
		}
	case "gte":
		if threshold, err := strconv.ParseFloat(value, 64); err == nil {
			return append(result, geConstraint{threshold: threshold})
		}
	case "lt":
		if threshold, err := strconv.ParseFloat(value, 64); err == nil {
			return append(result, ltConstraint{threshold: threshold})
		}
	case "lte":
		if threshold, err := strconv.ParseFloat(value, 64); err == nil {
			return append(result, leConstraint{threshold: threshold})
		}
	case "email":
		return append(result, emailConstraint{})
	case "url":
		return append(result, urlConstraint{})
	case "uri":
		return append(result, uriConstraint{})
	case "uuid":
		return append(result, uuidConstraint{})
	case "uuid3":
		return append(result, uuid3Constraint{})
	case "uuid4":
		return append(result, uuid4Constraint{})
	case "uuid5":
		return append(result, uuid5Constraint{})
	case "regexp":
		return append(result, buildRegexConstraint(value))
	case "ipv4":
		return append(result, ipv4Constraint{})
	case "ipv6":
		return append(result, ipv6Constraint{})
	case "oneof":
		return append(result, buildEnumConstraint(value))
	case "oneofci":
		return append(result, buildEnumCIConstraint(value))
	case "eq":
		if c, ok := buildEqConstraint(value); ok {
			return append(result, c)
		}
	case "ne":
		if c, ok := buildNeConstraint(value); ok {
			return append(result, c)
		}
	case "eq_ignore_case":
		if c, ok := buildEqIgnoreCaseConstraint(value); ok {
			return append(result, c)
		}
	case "ne_ignore_case":
		if c, ok := buildNeIgnoreCaseConstraint(value); ok {
			return append(result, c)
		}
	case "len":
		if c, ok := buildLenConstraint(value); ok {
			return append(result, c)
		}
	case "http_url":
		return append(result, httpURLConstraint{})
	}
	return result
}

// appendStringConstraint appends string validation constraints if name matches.
func appendStringConstraint(result []Constraint, name, value string) []Constraint {
	switch name {
	case "ascii":
		return append(result, asciiConstraint{})
	case "alpha":
		return append(result, alphaConstraint{})
	case "alphanum":
		return append(result, alphanumConstraint{})
	case "alphaspace":
		return append(result, alphaspaceConstraint{})
	case "alphanumspace":
		return append(result, alphanumspaceConstraint{})
	case "printascii":
		return append(result, printasciiConstraint{})
	case "numeric":
		return append(result, numericConstraint{})
	case "number":
		return append(result, numberConstraint{})
	case "hexadecimal":
		return append(result, hexadecimalConstraint{})
	case "alphaunicode":
		return append(result, alphaunicodeConstraint{})
	case "alphanumunicode":
		return append(result, alphanumunicodeConstraint{})
	case "contains":
		if c, ok := buildContainsConstraint(value); ok {
			return append(result, c)
		}
	case "excludes":
		if c, ok := buildExcludesConstraint(value); ok {
			return append(result, c)
		}
	case "startswith":
		if c, ok := buildStartswithConstraint(value); ok {
			return append(result, c)
		}
	case "endswith":
		if c, ok := buildEndswithConstraint(value); ok {
			return append(result, c)
		}
	case "startsnotwith":
		if c, ok := buildStartsnotwithConstraint(value); ok {
			return append(result, c)
		}
	case "endsnotwith":
		if c, ok := buildEndsnotwithConstraint(value); ok {
			return append(result, c)
		}
	case "containsany":
		if c, ok := buildContainsanyConstraint(value); ok {
			return append(result, c)
		}
	case "excludesall":
		if c, ok := buildExcludesallConstraint(value); ok {
			return append(result, c)
		}
	case "excludesrune":
		if c, ok := buildExcludesruneConstraint(value); ok {
			return append(result, c)
		}
	case "containsrune":
		if c, ok := buildContainsRuneConstraint(value); ok {
			return append(result, c)
		}
	case "lowercase":
		return append(result, lowercaseConstraint{})
	case "uppercase":
		return append(result, uppercaseConstraint{})
	case "multibyte":
		return append(result, multibyteConstraint{})
	case "urn_rfc2141":
		return append(result, urnRfc2141Constraint{})
	case "strip_whitespace":
		// In Validate mode: check if string has no leading/trailing whitespace
		return append(result, stripWhitespaceConstraint{})
	case "to_lower":
		// In Validate mode: check if string is all lowercase
		return append(result, lowercaseConstraint{})
	case "to_upper":
		// In Validate mode: check if string is all uppercase
		return append(result, uppercaseConstraint{})
	}
	return result
}

// appendNumericConstraint appends numeric validation constraints if name matches.
func appendNumericConstraint(result []Constraint, name, value string) []Constraint {
	switch name {
	case "positive":
		return append(result, positiveConstraint{})
	case "negative":
		return append(result, negativeConstraint{})
	case "multiple_of":
		if c, ok := buildMultipleOfConstraint(value); ok {
			return append(result, c)
		}
	case "max_digits":
		if c, ok := buildMaxDigitsConstraint(value); ok {
			return append(result, c)
		}
	case "decimal_places":
		if c, ok := buildDecimalPlacesConstraint(value); ok {
			return append(result, c)
		}
	case "disallow_inf_nan":
		return append(result, disallowInfNanConstraint{})
	}
	return result
}

// appendCollectionConstraint appends collection validation constraints if name matches.
func appendCollectionConstraint(result []Constraint, name, value string) []Constraint {
	switch name {
	case "unique":
		return append(result, uniqueConstraint{field: value})
	case "default":
		return append(result, defaultConstraint{value: value})
	}
	return result
}

// appendNetworkConstraint appends network format validators if name matches.
func appendNetworkConstraint(result []Constraint, name string) []Constraint {
	switch name {
	case "ip":
		return append(result, ipConstraint{})
	case "cidr":
		return append(result, cidrConstraint{})
	case "cidrv4":
		return append(result, cidrv4Constraint{})
	case "cidrv6":
		return append(result, cidrv6Constraint{})
	case "mac":
		return append(result, macConstraint{})
	case "hostname":
		return append(result, hostnameConstraint{})
	case "hostname_rfc1123":
		return append(result, hostnameRFC1123Constraint{})
	case "hostname_port":
		return append(result, hostnamePortConstraint{})
	case "fqdn":
		return append(result, fqdnConstraint{})
	case "port":
		return append(result, portConstraint{})
	case "tcp_addr":
		return append(result, tcpAddrConstraint{})
	case "udp_addr":
		return append(result, udpAddrConstraint{})
	case "tcp4_addr":
		return append(result, tcp4AddrConstraint{})
	case "https_url":
		return append(result, httpsURLConstraint{})
	}
	return result
}

// appendFinanceConstraint appends finance format validators if name matches.
func appendFinanceConstraint(result []Constraint, name string) []Constraint {
	switch name {
	case "credit_card":
		return append(result, creditCardConstraint{})
	case "btc_addr":
		return append(result, btcAddrConstraint{})
	case "btc_addr_bech32":
		return append(result, btcAddrBech32Constraint{})
	case "eth_addr":
		return append(result, ethAddrConstraint{})
	case "luhn_checksum":
		return append(result, luhnChecksumConstraint{})
	}
	return result
}

// appendIdentityConstraint appends identity format validators if name matches.
func appendIdentityConstraint(result []Constraint, name string) []Constraint {
	switch name {
	case "isbn":
		return append(result, isbnConstraint{})
	case "isbn10":
		return append(result, isbn10Constraint{})
	case "isbn13":
		return append(result, isbn13Constraint{})
	case "issn":
		return append(result, issnConstraint{})
	case "ssn":
		return append(result, ssnConstraint{})
	case "ein":
		return append(result, einConstraint{})
	case "e164":
		return append(result, e164Constraint{})
	}
	return result
}

// appendGeoConstraint appends geolocation format validators if name matches.
func appendGeoConstraint(result []Constraint, name string) []Constraint {
	switch name {
	case "latitude":
		return append(result, latitudeConstraint{})
	case "longitude":
		return append(result, longitudeConstraint{})
	}
	return result
}

// appendColorConstraint appends color format validators if name matches.
func appendColorConstraint(result []Constraint, name string) []Constraint {
	switch name {
	case "hexcolor":
		return append(result, hexcolorConstraint{})
	case "rgb":
		return append(result, rgbConstraint{})
	case "rgba":
		return append(result, rgbaConstraint{})
	case "hsl":
		return append(result, hslConstraint{})
	case "hsla":
		return append(result, hslaConstraint{})
	}
	return result
}

// appendEncodingConstraint appends encoding format validators if name matches.
func appendEncodingConstraint(result []Constraint, name string) []Constraint {
	switch name {
	case "jwt":
		return append(result, jwtConstraint{})
	case "json":
		return append(result, jsonConstraint{})
	case "base64":
		return append(result, base64Constraint{})
	case "base64url":
		return append(result, base64urlConstraint{})
	case "base64rawurl":
		return append(result, base64rawurlConstraint{})
	case "datauri":
		return append(result, datauriConstraint{})
	case "base32":
		return append(result, base32Constraint{})
	}
	return result
}

// appendHashConstraint appends hash format validators if name matches.
func appendHashConstraint(result []Constraint, name string) []Constraint {
	switch name {
	case "md4":
		return append(result, md4Constraint{})
	case "md5":
		return append(result, md5Constraint{})
	case "sha256":
		return append(result, sha256Constraint{})
	case "sha384":
		return append(result, sha384Constraint{})
	case "sha512":
		return append(result, sha512Constraint{})
	case "mongodb":
		return append(result, mongodbConstraint{})
	}
	return result
}

// appendMiscConstraint appends miscellaneous format validators if name matches.
func appendMiscConstraint(result []Constraint, name, value string) []Constraint {
	switch name {
	case "html":
		return append(result, htmlConstraint{})
	case "cron":
		return append(result, cronConstraint{})
	case "semver":
		return append(result, semverConstraint{})
	case "ulid":
		return append(result, ulidConstraint{})
	case "datetime":
		if c, ok := buildDatetimeConstraint(value); ok {
			return append(result, c)
		}
	case "timezone":
		return append(result, timezoneConstraint{})
	}
	return result
}
