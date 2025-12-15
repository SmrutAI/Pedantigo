// Package constraints provides validation constraint types and builders for pedantigo.
package constraints

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

// Finance and cryptocurrency constraint types.
type (
	creditCardConstraint    struct{} // credit_card: validates credit card number using Luhn algorithm (ISO/IEC 7812)
	btcAddrConstraint       struct{} // btc_addr: validates Bitcoin P2PKH/P2SH address (Base58Check)
	btcAddrBech32Constraint struct{} // btc_addr_bech32: validates Bitcoin Bech32 address (BIP-0173)
	ethAddrConstraint       struct{} // eth_addr: validates Ethereum address (EIP-55, 40 hex chars with 0x prefix)
	luhnChecksumConstraint  struct{} // luhn_checksum: validates any string passes Luhn algorithm
)

// Regex patterns for cryptocurrency addresses.
var (
	// btcBase58Regex matches Bitcoin P2PKH (starts with 1) and P2SH (starts with 3) addresses.
	// Valid Base58 chars: excludes 0, O, I, l (confusable characters).
	btcBase58Regex = regexp.MustCompile(`^[13][a-km-zA-HJ-NP-Z1-9]{24,33}$`)

	// ethAddrRegex matches Ethereum addresses: 0x prefix followed by 40 hex characters.
	ethAddrRegex = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
)

// Base58 alphabet used by Bitcoin (excludes 0, O, I, l).
const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// Bech32 character set (lowercase only, excludes 1, b, i, o).
const bech32Charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

// bech32Generator contains generator values for Bech32 checksum (BIP-0173).
var bech32Generator = []uint32{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}

// base58Decode decodes a Base58Check encoded string and returns the raw bytes.
// Returns nil if the string contains invalid characters.
func base58Decode(s string) []byte {
	// Build a reverse lookup table for the alphabet
	alphabetMap := make(map[rune]int64)
	for i, c := range base58Alphabet {
		alphabetMap[c] = int64(i)
	}

	// Decode using big integer arithmetic
	result := big.NewInt(0)
	base := big.NewInt(58)

	for _, c := range s {
		val, ok := alphabetMap[c]
		if !ok {
			return nil // Invalid character
		}
		result.Mul(result, base)
		result.Add(result, big.NewInt(val))
	}

	// Convert to bytes
	decoded := result.Bytes()

	// Count leading '1's (which represent leading zero bytes)
	leadingZeros := 0
	for _, c := range s {
		if c == '1' {
			leadingZeros++
		} else {
			break
		}
	}

	// Prepend leading zero bytes
	if leadingZeros > 0 {
		zeros := make([]byte, leadingZeros)
		decoded = append(zeros, decoded...)
	}

	return decoded
}

// validateBase58Check validates a Base58Check encoded address.
// Base58Check format: [version byte][payload][4-byte checksum]
// Checksum is first 4 bytes of double SHA256 of (version + payload).
func validateBase58Check(s string) bool {
	decoded := base58Decode(s)
	if len(decoded) < 5 {
		return false
	}

	// Split into payload and checksum
	payload := decoded[:len(decoded)-4]
	providedChecksum := decoded[len(decoded)-4:]

	// Calculate expected checksum (double SHA256)
	hash1 := sha256.Sum256(payload)
	hash2 := sha256.Sum256(hash1[:])
	expectedChecksum := hash2[:4]

	// Compare checksums
	for i := 0; i < 4; i++ {
		if providedChecksum[i] != expectedChecksum[i] {
			return false
		}
	}

	return true
}

// bech32Polymod computes the Bech32 checksum polynomial.
func bech32Polymod(values []int) uint32 {
	chk := uint32(1)
	for _, v := range values {
		top := chk >> 25
		chk = ((chk & 0x1ffffff) << 5) ^ uint32(v) //nolint:gosec // v is always in range 0-31
		for i := 0; i < 5; i++ {
			if (top>>i)&1 == 1 {
				chk ^= bech32Generator[i]
			}
		}
	}
	return chk
}

// bech32HRPExpand expands the human-readable part for checksum calculation.
func bech32HRPExpand(hrp string) []int {
	result := make([]int, 0, len(hrp)*2+1)
	for _, c := range hrp {
		result = append(result, int(c>>5))
	}
	result = append(result, 0)
	for _, c := range hrp {
		result = append(result, int(c&31))
	}
	return result
}

// bech32VerifyChecksum verifies the Bech32 checksum of an address.
// Returns true if the checksum is valid.
func bech32VerifyChecksum(hrp string, data []int) bool {
	values := append(bech32HRPExpand(hrp), data...)
	return bech32Polymod(values) == 1
}

// bech32Decode decodes a Bech32 string and returns the HRP and data values.
// Returns empty HRP if the string is invalid.
func bech32Decode(s string) (hrp string, data []int) {
	// Bech32 must be all lowercase or all uppercase
	if strings.ToLower(s) != s && strings.ToUpper(s) != s {
		return "", nil
	}
	s = strings.ToLower(s)

	// Find the separator (last '1' in the string)
	pos := strings.LastIndex(s, "1")
	if pos < 1 || pos+7 > len(s) || len(s) > 90 {
		return "", nil
	}

	hrp = s[:pos]
	dataStr := s[pos+1:]

	// Decode data characters
	data = make([]int, len(dataStr))
	for i, c := range dataStr {
		idx := strings.IndexRune(bech32Charset, c)
		if idx == -1 {
			return "", nil
		}
		data[i] = idx
	}

	// Verify checksum
	if !bech32VerifyChecksum(hrp, data) {
		return "", nil
	}

	return hrp, data
}

// validateBech32 validates a Bech32 encoded Bitcoin address (BIP-0173).
func validateBech32(s string) bool {
	// Must be lowercase
	if strings.ToLower(s) != s {
		return false
	}

	// Must start with bc1 (mainnet) or tb1 (testnet)
	if !strings.HasPrefix(s, "bc1") && !strings.HasPrefix(s, "tb1") {
		return false
	}

	// Length check (P2WPKH: 42 chars, P2WSH: 62 chars for witness v0)
	if len(s) < 42 || len(s) > 62 {
		return false
	}

	// Decode and verify checksum
	hrp, data := bech32Decode(s)
	if hrp == "" || data == nil {
		return false
	}

	// Check that HRP is valid Bitcoin prefix
	if hrp != "bc" && hrp != "tb" {
		return false
	}

	return true
}

// luhnValid checks if a string of digits passes the Luhn algorithm.
// The string must contain only digits (no spaces or dashes).
func luhnValid(s string) bool {
	if s == "" {
		return false
	}

	// Validate all characters are digits
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}

	sum := 0
	isSecond := false

	// Process digits from right to left
	for i := len(s) - 1; i >= 0; i-- {
		digit := int(s[i] - '0')

		if isSecond {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isSecond = !isSecond
	}

	return sum%10 == 0
}

// isAllZeros checks if a string consists entirely of zero characters.
func isAllZeros(s string) bool {
	for _, r := range s {
		if r != '0' {
			return false
		}
	}
	return true
}

// creditCardConstraint validates that a string is a valid credit card number using Luhn algorithm.
func (c creditCardConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("credit_card constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Credit cards must be only digits (no spaces, dashes, or other chars)
	for _, r := range str {
		if r < '0' || r > '9' {
			return fmt.Errorf("must be a valid credit card number")
		}
	}

	// Check length (13-19 digits for standard credit cards)
	if len(str) < 13 || len(str) > 19 {
		return fmt.Errorf("must be a valid credit card number")
	}

	// Card numbers cannot be all zeros
	if isAllZeros(str) {
		return fmt.Errorf("must be a valid credit card number")
	}

	// Check Luhn algorithm
	if !luhnValid(str) {
		return fmt.Errorf("must be a valid credit card number")
	}

	return nil
}

// btcAddrConstraint validates that a string is a valid Bitcoin P2PKH/P2SH address.
func (c btcAddrConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("btc_addr constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// First check format with regex
	if !btcBase58Regex.MatchString(str) {
		return fmt.Errorf("must be a valid Bitcoin address")
	}

	// Then validate Base58Check checksum
	if !validateBase58Check(str) {
		return fmt.Errorf("must be a valid Bitcoin address")
	}

	return nil
}

// btcAddrBech32Constraint validates that a string is a valid Bitcoin Bech32 address.
func (c btcAddrBech32Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("btc_addr_bech32 constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Validate Bech32 format and checksum
	if !validateBech32(str) {
		return fmt.Errorf("must be a valid Bitcoin Bech32 address")
	}

	return nil
}

// ethAddrConstraint validates that a string is a valid Ethereum address.
func (c ethAddrConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("eth_addr constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Validate against Ethereum address regex (0x + 40 hex chars)
	if !ethAddrRegex.MatchString(str) {
		return fmt.Errorf("must be a valid Ethereum address")
	}

	return nil
}

// luhnChecksumConstraint validates that a string passes the Luhn algorithm checksum.
func (c luhnChecksumConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("luhn_checksum constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Validate using Luhn algorithm (no spaces or dashes allowed)
	if !luhnValid(str) {
		return fmt.Errorf("must pass Luhn checksum validation")
	}

	return nil
}
