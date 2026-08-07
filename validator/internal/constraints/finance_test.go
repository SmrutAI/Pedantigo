package constraints

import "testing"

// TestCreditCardConstraint tests creditCardConstraint.Validate() for valid credit card numbers.
func TestCreditCardConstraint(t *testing.T) {
	runSimpleConstraintTests(t, creditCardConstraint{}, []simpleTestCase{
		// Valid credit card numbers (pass Luhn algorithm)
		{"valid Visa", "4111111111111111", false},
		{"valid MasterCard", "5500000000000004", false},
		{"valid Amex", "378282246310005", false},
		{"valid Discover", "6011111111111117", false},
		{"valid Visa 13-digit", "4222222222222", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid credit card numbers
		{"invalid luhn checksum", "4111111111111112", true},
		{"too short", "411111", true},
		{"too short single digit", "4", true},
		{"contains letters", "4111111111a11111", true},
		{"contains special chars", "4111-1111-1111-1111", true},
		{"all zeros", "0000000000000000", true},
		{"random invalid", "1234567890123456", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 4111111111111111, true},
		{"invalid type - bool", true, true},
	})
}

// TestBtcAddrConstraint tests btcAddrConstraint.Validate() for valid Bitcoin P2PKH/P2SH addresses.
func TestBtcAddrConstraint(t *testing.T) {
	runSimpleConstraintTests(t, btcAddrConstraint{}, []simpleTestCase{
		// Valid P2PKH addresses (start with 1)
		{"valid P2PKH address 1", "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", false},
		{"valid P2PKH address 2", "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", false},
		// Valid P2SH addresses (start with 3)
		{"valid P2SH address 1", "3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy", false},
		{"valid P2SH address 2", "3EktnHQD7RiAE6uzMj2ZifT9YgRrkSgzQX", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid addresses
		{"invalid checksum P2PKH", "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN3", true},
		{"invalid checksum P2SH", "3J98t1WpEZ73CNmQviecrnyiWrnqRhWNL0", true},
		{"bech32 address (wrong type)", "bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4", true},
		{"too short", "1BvBMSEYstWet", true},
		{"too long", "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2extra", true},
		{"invalid characters (O)", "1BvBMSEYstWetqTFn5Au4m4OFg7xJaNVN2", true},
		{"invalid characters (I)", "1BvBMSEYstWetqTFn5Au4m4IFg7xJaNVN2", true},
		{"invalid characters (l)", "1BvBMSEYstWetqTFn5Au4m4lFg7xJaNVN2", true},
		{"invalid start character", "2BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", true},
		{"contains spaces", "1BvBMSEYstWetqTFn5Au4m4 Fg7xJaNVN2", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestBtcAddrBech32Constraint tests btcAddrBech32Constraint.Validate() for valid Bitcoin Bech32 addresses.
func TestBtcAddrBech32Constraint(t *testing.T) {
	runSimpleConstraintTests(t, btcAddrBech32Constraint{}, []simpleTestCase{
		// Valid Bech32 addresses (mainnet, start with bc1)
		{"valid bech32 mainnet P2WPKH", "bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4", false},
		{"valid bech32 mainnet P2WSH", "bc1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3qccfmv3", false},
		// Valid Bech32 addresses (testnet, start with tb1)
		{"valid bech32 testnet", "tb1qw508d6qejxtdg4y5r3zarvary0c5xw7kxpjzsx", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid addresses
		{"P2PKH address (wrong type)", "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", true},
		{"P2SH address (wrong type)", "3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy", true},
		{"invalid prefix", "bc2qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4", true},
		{"invalid checksum", "bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t5", true},
		{"invalid character (b)", "bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kb8f3t4", true},
		{"too short", "bc1qw508d6", true},
		{"mixed case (invalid)", "BC1QW508D6QEJXTDG4Y5R3ZARVARY0C5XW7KV8F3T4", true},
		{"contains invalid chars", "bc1qw508d6qejxtdg4y5r3zarvar!0c5xw7kv8f3t4", true},
		{"empty after prefix", "bc1", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestEthAddrConstraint tests ethAddrConstraint.Validate() for valid Ethereum addresses.
func TestEthAddrConstraint(t *testing.T) {
	runSimpleConstraintTests(t, ethAddrConstraint{}, []simpleTestCase{
		// Valid Ethereum addresses (0x + 40 hex chars)
		{"valid eth address lowercase", "0x742d35cc6634c0532925a3b844bc9e7595f8fee5", false},
		{"valid eth address mixed case (checksum)", "0x742d35Cc6634C0532925a3b844Bc9e7595f8fEe5", false},
		{"valid eth address all zeros", "0x0000000000000000000000000000000000000000", false},
		{"valid eth address all f", "0xffffffffffffffffffffffffffffffffffffffff", false},
		{"valid eth address uppercase", "0x742D35CC6634C0532925A3B844BC9E7595F8FEE5", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid addresses
		{"missing 0x prefix", "742d35Cc6634C0532925a3b844Bc9e7595f8fEe5", true},
		{"wrong prefix 0X", "0X742d35Cc6634C0532925a3b844Bc9e7595f8fEe5", true},
		{"too short", "0x742d35Cc6634C0532925a3b844Bc9e7595f8fE", true},
		{"too long", "0x742d35Cc6634C0532925a3b844Bc9e7595f8fEe5a", true},
		{"invalid hex char g", "0x742d35Cc6634C0532925a3b844Bc9e7595f8fEeg", true},
		{"invalid hex char z", "0xz42d35Cc6634C0532925a3b844Bc9e7595f8fEe5", true},
		{"contains spaces", "0x742d35Cc6634C0532925 3b844Bc9e7595f8fEe5", true},
		{"only prefix", "0x", true},
		{"bitcoin address", "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestLuhnChecksumConstraint tests luhnChecksumConstraint.Validate() for Luhn algorithm validation.
func TestLuhnChecksumConstraint(t *testing.T) {
	runSimpleConstraintTests(t, luhnChecksumConstraint{}, []simpleTestCase{
		// Valid Luhn checksums
		{"valid luhn example", "79927398713", false},
		{"valid visa card", "4111111111111111", false},
		{"valid mastercard", "5500000000000004", false},
		{"valid amex", "378282246310005", false},
		{"valid short number", "18", false},
		{"valid single zero", "0", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid Luhn checksums
		{"invalid luhn example", "79927398714", true},
		{"invalid sequence", "1234567890", true},
		{"invalid off by one", "4111111111111112", true},
		{"invalid all ones", "1111111111111111", true},
		{"contains letters", "79927398a13", true},
		{"contains spaces", "7992 7398 713", true},
		{"contains dashes", "7992-7398-713", true},
		{"negative number string", "-79927398713", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 79927398713, true},
		{"invalid type - bool", true, true},
	})
}
