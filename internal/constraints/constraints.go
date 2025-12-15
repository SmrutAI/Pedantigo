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
	CMin    = "min"
	CMax    = "max"
	CGt     = "gt"
	CGte    = "gte"
	CLt     = "lt"
	CLte    = "lte"
	CEmail  = "email"
	CUrl    = "url"
	CUuid   = "uuid"
	CRegexp = "regexp"
	CIpv4   = "ipv4"
	CIpv6   = "ipv6"
	COneof  = "oneof"
	CConst  = "const"
	CLen    = "len"

	// String constraints.
	CAscii      = "ascii"
	CAlpha      = "alpha"
	CAlphanum   = "alphanum"
	CContains   = "contains"
	CExcludes   = "excludes"
	CStartswith = "startswith"
	CEndswith   = "endswith"
	CLowercase  = "lowercase"
	CUppercase  = "uppercase"

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
	CFqdn            = "fqdn"
	CPort            = "port"
	CTcpAddr         = "tcp_addr"
	CUdpAddr         = "udp_addr"
	CTcp4Addr        = "tcp4_addr"

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

	// Hash constraints.
	CMd4     = "md4"
	CMd5     = "md5"
	CSha256  = "sha256"
	CSha384  = "sha384"
	CSha512  = "sha512"
	CMongodb = "mongodb"

	// Misc constraints.
	CHtml   = "html"
	CCron   = "cron"
	CSemver = "semver"
	CUlid   = "ulid"

	// Special.
	CRequired = "required"
)

// Shared regex patterns used by string constraints.
var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	uuidRegex     = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	alphaRegex    = regexp.MustCompile(`^[a-zA-Z]+$`)
	alphanumRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
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
		switch name {
		case CRequired:
			// Skip: 'required' is only checked during Unmarshal (missing JSON keys).
			// It doesn't apply to Validate() on manually created structs.
			continue

		// Core constraints.
		case CMin, CMax, CGt, CGte, CLt, CLte, CEmail, CUrl, CUuid, CRegexp, CIpv4, CIpv6, COneof, CConst, CLen:
			result = appendCoreConstraint(result, name, value, fieldType)

		// String constraints.
		case CAscii, CAlpha, CAlphanum, CContains, CExcludes, CStartswith, CEndswith, CLowercase, CUppercase:
			result = appendStringConstraint(result, name, value)

		// Numeric constraints.
		case CPositive, CNegative, CMultipleOf, CMaxDigits, CDecimalPlaces, CDisallowInfNan:
			result = appendNumericConstraint(result, name, value)

		// Collection constraints.
		case CUnique, CDefault:
			result = appendCollectionConstraint(result, name, value)

		// Network constraints.
		case CIp, CCidr, CCidrv4, CCidrv6, CMac, CHostname, CHostnameRfc1123, CFqdn, CPort, CTcpAddr, CUdpAddr, CTcp4Addr:
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
		case CJwt, CJson, CBase64, CBase64url, CBase64rawurl:
			result = appendEncodingConstraint(result, name)

		// Hash constraints.
		case CMd4, CMd5, CSha256, CSha384, CSha512, CMongodb:
			result = appendHashConstraint(result, name)

		// Misc constraints.
		case CHtml, CCron, CSemver, CUlid:
			result = appendMiscConstraint(result, name)
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
	case "uuid":
		return append(result, uuidConstraint{})
	case "regexp":
		return append(result, buildRegexConstraint(value))
	case "ipv4":
		return append(result, ipv4Constraint{})
	case "ipv6":
		return append(result, ipv6Constraint{})
	case "oneof":
		return append(result, buildEnumConstraint(value))
	case "const":
		if c, ok := buildConstConstraint(value); ok {
			return append(result, c)
		}
	case "len":
		if c, ok := buildLenConstraint(value); ok {
			return append(result, c)
		}
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
	case "lowercase":
		return append(result, lowercaseConstraint{})
	case "uppercase":
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
func appendMiscConstraint(result []Constraint, name string) []Constraint {
	switch name {
	case "html":
		return append(result, htmlConstraint{})
	case "cron":
		return append(result, cronConstraint{})
	case "semver":
		return append(result, semverConstraint{})
	case "ulid":
		return append(result, ulidConstraint{})
	}
	return result
}
