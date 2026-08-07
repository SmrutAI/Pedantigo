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
	CodeInvalidEmail        = "INVALID_EMAIL"
	CodeInvalidURL          = "INVALID_URL"
	CodeInvalidUUID         = "INVALID_UUID"
	CodeInvalidIPv4         = "INVALID_IPV4"
	CodeInvalidIPv6         = "INVALID_IPV6"
	CodeInvalidIP           = "INVALID_IP"
	CodeInvalidURI          = "INVALID_URI"
	CodeInvalidHostname     = "INVALID_HOSTNAME"
	CodeInvalidMAC          = "INVALID_MAC"
	CodeInvalidCIDR         = "INVALID_CIDR"
	CodeInvalidPort         = "INVALID_PORT"
	CodeInvalidTCPAddr      = "INVALID_TCP_ADDR"
	CodeInvalidUDPAddr      = "INVALID_UDP_ADDR"
	CodeInvalidFQDN         = "INVALID_FQDN"
	CodeInvalidHostnamePort = "INVALID_HOSTNAME_PORT"
	CodeInvalidHTTPURL      = "INVALID_HTTP_URL"
	CodePatternMismatch     = "PATTERN_MISMATCH"
	CodeInvalidDatetime     = "INVALID_DATETIME"

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

	// ISO code constraints.
	CodeInvalidCurrencyCode = "INVALID_CURRENCY_CODE"
	CodeInvalidLanguageTag  = "INVALID_LANGUAGE_TAG"
	CodeInvalidSubdivision  = "INVALID_SUBDIVISION_CODE"

	// File system constraints.
	CodeInvalidPath  = "INVALID_PATH"
	CodeFileNotFound = "FILE_NOT_FOUND"
	CodeDirNotFound  = "DIRECTORY_NOT_FOUND"
	CodeInvalidImage = "INVALID_IMAGE"

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
	CodeInvalidUUIDv3       = "INVALID_UUID_V3"
	CodeInvalidUUIDv4       = "INVALID_UUID_V4"
	CodeInvalidUUIDv5       = "INVALID_UUID_V5"
	CodeInvalidMultibyte    = "INVALID_MULTIBYTE"
	CodeInvalidDataURI      = "INVALID_DATA_URI"
	CodeInvalidBase32       = "INVALID_BASE32"
	CodeInvalidURN          = "INVALID_URN"
	CodeInvalidHTTPSURL     = "INVALID_HTTPS_URL"

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
	CodeMustBeASCII           = "MUST_BE_ASCII"
	CodeMustBeAlpha           = "MUST_BE_ALPHA"
	CodeMustBeAlphanum        = "MUST_BE_ALPHANUM"
	CodeMustBeNumeric         = "MUST_BE_NUMERIC"
	CodeMustBeNumber          = "MUST_BE_NUMBER"
	CodeMustBeHexadecimal     = "MUST_BE_HEXADECIMAL"
	CodeMustBeAlphaUnicode    = "MUST_BE_ALPHA_UNICODE"
	CodeMustBeAlphanumUnicode = "MUST_BE_ALPHANUM_UNICODE"
	CodeMustBeAlphaSpace      = "MUST_BE_ALPHA_SPACE"
	CodeMustBeAlphanumSpace   = "MUST_BE_ALPHANUM_SPACE"
	CodeMustBePrintableASCII  = "MUST_BE_PRINTABLE_ASCII"
	CodeMustContain           = "MUST_CONTAIN"
	CodeMustNotContain        = "MUST_NOT_CONTAIN"
	CodeMustStartWith         = "MUST_START_WITH"
	CodeMustEndWith           = "MUST_END_WITH"
	CodeMustNotStartWith      = "MUST_NOT_START_WITH"
	CodeMustNotEndWith        = "MUST_NOT_END_WITH"
	CodeMustContainAny        = "MUST_CONTAIN_ANY"
	CodeMustExcludeAll        = "MUST_EXCLUDE_ALL"
	CodeMustExcludeRune       = "MUST_EXCLUDE_RUNE"
	CodeContainsRune          = "CONTAINS_RUNE"
	CodeMustBeLowercase       = "MUST_BE_LOWERCASE"
	CodeMustBeUppercase       = "MUST_BE_UPPERCASE"
	CodeMustBeStripped        = "MUST_BE_STRIPPED"

	// Enum/const constraints.
	CodeInvalidEnum   = "INVALID_ENUM"
	CodeConstMismatch = "CONST_MISMATCH"
	CodeEqIgnoreCase  = "EQ_IGNORE_CASE"
	CodeNeIgnoreCase  = "NE_IGNORE_CASE"

	// Collection constraints.
	CodeNotUnique = "NOT_UNIQUE"

	// Cross-field constraints.
	CodeMustEqualField     = "MUST_EQUAL_FIELD"
	CodeMustNotEqualField  = "MUST_NOT_EQUAL_FIELD"
	CodeMustBeGTField      = "MUST_BE_GT_FIELD"
	CodeMustBeGTEField     = "MUST_BE_GTE_FIELD"
	CodeMustBeLTField      = "MUST_BE_LT_FIELD"
	CodeMustBeLTEField     = "MUST_BE_LTE_FIELD"
	CodeExcludedIf         = "EXCLUDED_IF"
	CodeExcludedUnless     = "EXCLUDED_UNLESS"
	CodeExcludedWith       = "EXCLUDED_WITH"
	CodeExcludedWithout    = "EXCLUDED_WITHOUT"
	CodeRequiredWithAll    = "REQUIRED_WITH_ALL"
	CodeRequiredWithoutAll = "REQUIRED_WITHOUT_ALL"
	CodeExcludedWithAll    = "EXCLUDED_WITH_ALL"
	CodeExcludedWithoutAll = "EXCLUDED_WITHOUT_ALL"

	// OR constraints.
	CodeOrConstraintFailed = "OR_CONSTRAINT_FAILED"

	// Type errors.
	CodeUnknownField    = "UNKNOWN_FIELD"
	CodeInvalidType     = "INVALID_TYPE"
	CodeUnsupportedType = "UNSUPPORTED_TYPE"

	// Custom validation constraints.
	CodeFieldPathError    = "FIELD_PATH_ERROR"   // Nil pointer encountered in field path resolution
	CodeCustomValidation  = "CUSTOM_VALIDATION"  // Custom validator failed
	CodeContextValidation = "CONTEXT_VALIDATION" // Context-aware validator failed
)
