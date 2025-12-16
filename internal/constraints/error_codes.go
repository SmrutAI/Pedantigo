package constraints

// Error code constants for validation errors.
// Using SCREAMING_SNAKE_CASE convention.
const (
	// Required constraints.
	CodeRequired        = "REQUIRED"
	CodeRequiredIf      = "REQUIRED_IF"
	CodeRequiredUnless  = "REQUIRED_UNLESS"
	CodeRequiredWith    = "REQUIRED_WITH"
	CodeRequiredWithout = "REQUIRED_WITHOUT"

	// Format constraints.
	CodeInvalidEmail    = "INVALID_EMAIL"
	CodeInvalidURL      = "INVALID_URL"
	CodeInvalidUUID     = "INVALID_UUID"
	CodeInvalidIPv4     = "INVALID_IPV4"
	CodeInvalidIPv6     = "INVALID_IPV6"
	CodeInvalidIP       = "INVALID_IP"
	CodeInvalidURI      = "INVALID_URI"
	CodeInvalidHostname = "INVALID_HOSTNAME"
	CodeInvalidMAC      = "INVALID_MAC"
	CodeInvalidCIDR     = "INVALID_CIDR"
	CodeInvalidPort     = "INVALID_PORT"
	CodeInvalidTCPAddr  = "INVALID_TCP_ADDR"
	CodeInvalidUDPAddr  = "INVALID_UDP_ADDR"
	CodeInvalidFQDN     = "INVALID_FQDN"
	CodePatternMismatch = "PATTERN_MISMATCH"

	// Identity/Publishing constraints.
	CodeInvalidISBN   = "INVALID_ISBN"
	CodeInvalidISBN10 = "INVALID_ISBN10"
	CodeInvalidISBN13 = "INVALID_ISBN13"
	CodeInvalidISSN   = "INVALID_ISSN"
	CodeInvalidSSN    = "INVALID_SSN"
	CodeInvalidEIN    = "INVALID_EIN"
	CodeInvalidE164   = "INVALID_E164"

	// Finance constraints.
	CodeInvalidLuhn            = "INVALID_LUHN"
	CodeInvalidCreditCard      = "INVALID_CREDIT_CARD"
	CodeInvalidBitcoinAddress  = "INVALID_BITCOIN_ADDRESS"
	CodeInvalidBitcoinBech32   = "INVALID_BITCOIN_BECH32"
	CodeInvalidEthereumAddress = "INVALID_ETHEREUM_ADDRESS"

	// Hash constraints.
	CodeInvalidMD4     = "INVALID_MD4"
	CodeInvalidMD5     = "INVALID_MD5"
	CodeInvalidSHA256  = "INVALID_SHA256"
	CodeInvalidSHA384  = "INVALID_SHA384"
	CodeInvalidSHA512  = "INVALID_SHA512"
	CodeInvalidMongoDB = "INVALID_MONGODB"

	// Miscellaneous format constraints.
	CodeInvalidHTML   = "INVALID_HTML"
	CodeInvalidCron   = "INVALID_CRON"
	CodeInvalidSemver = "INVALID_SEMVER"
	CodeInvalidULID   = "INVALID_ULID"

	// Geographic constraints.
	CodeInvalidLatitude    = "INVALID_LATITUDE"
	CodeInvalidLongitude   = "INVALID_LONGITUDE"
	CodeInvalidCountryCode = "INVALID_COUNTRY_CODE"
	CodeInvalidPostalCode  = "INVALID_POSTAL_CODE"
	CodeInvalidTimezone    = "INVALID_TIMEZONE"

	// Color constraints.
	CodeInvalidHexColor = "INVALID_HEX_COLOR"
	CodeInvalidRGBColor = "INVALID_RGB_COLOR"
	CodeInvalidRGBA     = "INVALID_RGBA"
	CodeInvalidHSL      = "INVALID_HSL"
	CodeInvalidHSLA     = "INVALID_HSLA"

	// Encoding constraints.
	CodeInvalidBase64       = "INVALID_BASE64"
	CodeInvalidBase64URL    = "INVALID_BASE64URL"
	CodeInvalidBase64RawURL = "INVALID_BASE64_RAW_URL"
	CodeInvalidJSON         = "INVALID_JSON"
	CodeInvalidJWT          = "INVALID_JWT"

	// Length constraints.
	CodeMinLength   = "MIN_LENGTH"
	CodeMaxLength   = "MAX_LENGTH"
	CodeExactLength = "EXACT_LENGTH"

	// Numeric constraints.
	CodeMinValue         = "MIN_VALUE"
	CodeMaxValue         = "MAX_VALUE"
	CodeExclusiveMin     = "EXCLUSIVE_MIN"
	CodeExclusiveMax     = "EXCLUSIVE_MAX"
	CodeMustBePositive   = "MUST_BE_POSITIVE"
	CodeMustBeNegative   = "MUST_BE_NEGATIVE"
	CodeMultipleOf       = "MULTIPLE_OF"
	CodeMaxDigits        = "MAX_DIGITS"
	CodeDecimalPlaces    = "DECIMAL_PLACES"
	CodeInfNanNotAllowed = "INF_NAN_NOT_ALLOWED"

	// String constraints.
	CodeMustBeASCII     = "MUST_BE_ASCII"
	CodeMustBeAlpha     = "MUST_BE_ALPHA"
	CodeMustBeAlphanum  = "MUST_BE_ALPHANUM"
	CodeMustContain     = "MUST_CONTAIN"
	CodeMustNotContain  = "MUST_NOT_CONTAIN"
	CodeMustStartWith   = "MUST_START_WITH"
	CodeMustEndWith     = "MUST_END_WITH"
	CodeMustBeLowercase = "MUST_BE_LOWERCASE"
	CodeMustBeUppercase = "MUST_BE_UPPERCASE"

	// Enum/const constraints.
	CodeInvalidEnum   = "INVALID_ENUM"
	CodeConstMismatch = "CONST_MISMATCH"

	// Collection constraints.
	CodeNotUnique = "NOT_UNIQUE"

	// Cross-field constraints.
	CodeMustEqualField    = "MUST_EQUAL_FIELD"
	CodeMustNotEqualField = "MUST_NOT_EQUAL_FIELD"
	CodeMustBeGTField     = "MUST_BE_GT_FIELD"
	CodeMustBeGTEField    = "MUST_BE_GTE_FIELD"
	CodeMustBeLTField     = "MUST_BE_LT_FIELD"
	CodeMustBeLTEField    = "MUST_BE_LTE_FIELD"
	CodeExcludedIf        = "EXCLUDED_IF"
	CodeExcludedUnless    = "EXCLUDED_UNLESS"
	CodeExcludedWith      = "EXCLUDED_WITH"
	CodeExcludedWithout   = "EXCLUDED_WITHOUT"

	// Type errors.
	CodeUnknownField    = "UNKNOWN_FIELD"
	CodeInvalidType     = "INVALID_TYPE"
	CodeUnsupportedType = "UNSUPPORTED_TYPE"
)
