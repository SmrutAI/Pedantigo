package constraints

import "testing"

// TestMd4Constraint tests md4Constraint.Validate() for valid MD4 hash format (32 hex chars).
func TestMd4Constraint(t *testing.T) {
	runSimpleConstraintTests(t, md4Constraint{}, []simpleTestCase{
		// Valid MD4 hashes (32 hex characters)
		{"valid md4 lowercase", "d41d8cd98f00b204e9800998ecf8427e", false},
		{"valid md4 uppercase", "D41D8CD98F00B204E9800998ECF8427E", false},
		{"valid md4 mixed case", "d41D8cd98F00b204E9800998ecf8427E", false},
		{"valid md4 all zeros", "00000000000000000000000000000000", false},
		{"valid md4 all f", "ffffffffffffffffffffffffffffffff", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid MD4 hashes
		{"invalid 31 chars", "d41d8cd98f00b204e9800998ecf8427", true},
		{"invalid 33 chars", "d41d8cd98f00b204e9800998ecf8427e0", true},
		{"invalid non-hex g", "g41d8cd98f00b204e9800998ecf8427e", true},
		{"invalid non-hex z", "z41d8cd98f00b204e9800998ecf8427e", true},
		{"invalid spaces", "d41d8cd98f00b204 e9800998ecf8427e", true},
		{"invalid with hyphen", "d41d8cd9-8f00-b204-e980-0998ecf8427e", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestMd5Constraint tests md5Constraint.Validate() for valid MD5 hash format (32 hex chars).
func TestMd5Constraint(t *testing.T) {
	runSimpleConstraintTests(t, md5Constraint{}, []simpleTestCase{
		// Valid MD5 hashes (32 hex characters)
		{"valid md5 lowercase", "d41d8cd98f00b204e9800998ecf8427e", false},
		{"valid md5 uppercase", "D41D8CD98F00B204E9800998ECF8427E", false},
		{"valid md5 mixed case", "d41D8cd98F00b204E9800998ecf8427E", false},
		{"valid md5 all zeros", "00000000000000000000000000000000", false},
		{"valid md5 all f", "ffffffffffffffffffffffffffffffff", false},
		{"valid md5 hello", "5d41402abc4b2a76b9719d911017c592", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid MD5 hashes
		{"invalid 31 chars", "d41d8cd98f00b204e9800998ecf8427", true},
		{"invalid 33 chars", "d41d8cd98f00b204e9800998ecf8427e0", true},
		{"invalid non-hex g", "g41d8cd98f00b204e9800998ecf8427e", true},
		{"invalid non-hex z", "z41d8cd98f00b204e9800998ecf8427e", true},
		{"invalid spaces", "d41d8cd98f00b204 e9800998ecf8427e", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestSha256Constraint tests sha256Constraint.Validate() for valid SHA256 hash format (64 hex chars).
func TestSha256Constraint(t *testing.T) {
	runSimpleConstraintTests(t, sha256Constraint{}, []simpleTestCase{
		// Valid SHA256 hashes (64 hex characters)
		{"valid sha256 lowercase", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", false},
		{"valid sha256 uppercase", "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855", false},
		{"valid sha256 all zeros", "0000000000000000000000000000000000000000000000000000000000000000", false},
		{"valid sha256 all f", "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid SHA256 hashes
		{"invalid 63 chars", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b85", true},
		{"invalid 65 chars", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b8550", true},
		{"invalid non-hex g", "g3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", true},
		{"invalid spaces", "e3b0c44298fc1c149afbf4c8996fb924 27ae41e4649b934ca495991b7852b855", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestSha384Constraint tests sha384Constraint.Validate() for valid SHA384 hash format (96 hex chars).
func TestSha384Constraint(t *testing.T) {
	runSimpleConstraintTests(t, sha384Constraint{}, []simpleTestCase{
		// Valid SHA384 hashes (96 hex characters)
		{"valid sha384 lowercase", "38b060a751ac96384cd9327eb1b1e36a21fdb71114be07434c0cc7bf63f6e1da274edebfe76f65fbd51ad2f14898b95b", false},
		{"valid sha384 uppercase", "38B060A751AC96384CD9327EB1B1E36A21FDB71114BE07434C0CC7BF63F6E1DA274EDEBFE76F65FBD51AD2F14898B95B", false},
		{"valid sha384 all zeros", "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", false},
		{"valid sha384 all f", "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid SHA384 hashes
		{"invalid 95 chars", "38b060a751ac96384cd9327eb1b1e36a21fdb71114be07434c0cc7bf63f6e1da274edebfe76f65fbd51ad2f14898b95", true},
		{"invalid 97 chars", "38b060a751ac96384cd9327eb1b1e36a21fdb71114be07434c0cc7bf63f6e1da274edebfe76f65fbd51ad2f14898b95b0", true},
		{"invalid non-hex g", "g8b060a751ac96384cd9327eb1b1e36a21fdb71114be07434c0cc7bf63f6e1da274edebfe76f65fbd51ad2f14898b95b", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestSha512Constraint tests sha512Constraint.Validate() for valid SHA512 hash format (128 hex chars).
func TestSha512Constraint(t *testing.T) {
	runSimpleConstraintTests(t, sha512Constraint{}, []simpleTestCase{
		// Valid SHA512 hashes (128 hex characters)
		{"valid sha512 lowercase", "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e", false},
		{"valid sha512 uppercase", "CF83E1357EEFB8BDF1542850D66D8007D620E4050B5715DC83F4A921D36CE9CE47D0D13C5D85F2B0FF8318D2877EEC2F63B931BD47417A81A538327AF927DA3E", false},
		{"valid sha512 all zeros", "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", false},
		{"valid sha512 all f", "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid SHA512 hashes
		{"invalid 127 chars", "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3", true},
		{"invalid 129 chars", "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e0", true},
		{"invalid non-hex g", "gf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestMongodbConstraint tests mongodbConstraint.Validate() for valid MongoDB ObjectId format (24 hex chars).
func TestMongodbConstraint(t *testing.T) {
	runSimpleConstraintTests(t, mongodbConstraint{}, []simpleTestCase{
		// Valid MongoDB ObjectIds (24 hex characters)
		{"valid mongodb lowercase", "507f1f77bcf86cd799439011", false},
		{"valid mongodb uppercase", "507F1F77BCF86CD799439011", false},
		{"valid mongodb mixed case", "507f1F77bcf86CD799439011", false},
		{"valid mongodb all zeros", "000000000000000000000000", false},
		{"valid mongodb all f", "ffffffffffffffffffffffff", false},
		{"valid mongodb example 2", "5d6ede6a0ba62570afcedd3a", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid MongoDB ObjectIds
		{"invalid 23 chars", "507f1f77bcf86cd79943901", true},
		{"invalid 25 chars", "507f1f77bcf86cd7994390110", true},
		{"invalid non-hex g", "g07f1f77bcf86cd799439011", true},
		{"invalid non-hex z", "z07f1f77bcf86cd799439011", true},
		{"invalid spaces", "507f1f77bcf8 6cd799439011", true},
		{"invalid with hyphen", "507f1f77-bcf8-6cd7-9943-9011", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}
