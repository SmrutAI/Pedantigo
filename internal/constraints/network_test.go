package constraints

import (
	"testing"
)

// TestIpv4Constraint tests ipv4Constraint.Validate() for valid IPv4 addresses.
func TestIpv4Constraint(t *testing.T) {
	runSimpleConstraintTests(t, ipv4Constraint{}, []simpleTestCase{
		// Valid IPv4 addresses
		{"valid IPv4 - localhost", "127.0.0.1", false},
		{"valid IPv4 - private range", "192.168.1.1", false},
		{"valid IPv4 - private range 10", "10.0.0.1", false},
		{"valid IPv4 - private range 172", "172.16.0.1", false},
		{"valid IPv4 - zeros", "0.0.0.0", false},
		{"valid IPv4 - broadcast", "255.255.255.255", false},
		{"valid IPv4 - google DNS", "8.8.8.8", false},
		{"valid IPv4 - public IP", "1.1.1.1", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid IPv4 addresses
		{"invalid IPv4 - out of range", "256.1.1.1", true},
		{"invalid IPv4 - too few octets", "192.168.1", true},
		{"invalid IPv4 - too many octets", "192.168.1.1.1", true},
		{"invalid IPv4 - letters", "192.168.a.1", true},
		{"invalid IPv4 - empty octets", "192.168..1", true},
		// IPv6 addresses - should fail
		{"IPv6 address fails", "::1", true},
		{"IPv6 full address fails", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", true},
		{"IPv6 compressed fails", "2001:db8::1", true},
		// Other invalid formats
		{"hostname not IP", "example.com", true},
		{"CIDR notation not IP", "192.168.1.0/24", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestIpv6Constraint tests ipv6Constraint.Validate() for valid IPv6 addresses.
func TestIpv6Constraint(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		// Valid IPv6 addresses - full form
		{name: "valid IPv6 - localhost loopback", value: "::1", wantErr: false},
		{name: "valid IPv6 - unspecified", value: "::", wantErr: false},
		{name: "valid IPv6 - full form", value: "2001:0db8:85a3:0000:0000:8a2e:0370:7334", wantErr: false},

		// Valid IPv6 addresses - compressed form
		{name: "valid IPv6 - compressed", value: "2001:db8::1", wantErr: false},
		{name: "valid IPv6 - compressed form 2", value: "2001:db8::8a2e:370:7334", wantErr: false},
		{name: "valid IPv6 - link-local", value: "fe80::1", wantErr: false},
		{name: "valid IPv6 - multicast", value: "ff02::1", wantErr: false},

		// Valid IPv6 addresses - zone ID variants (some may fail depending on implementation)
		{name: "valid IPv6 - with numbers", value: "1234:5678:90ab:cdef:1234:5678:90ab:cdef", wantErr: false},

		// Empty string - should be skipped
		{name: "empty string", value: "", wantErr: false},

		// IPv4 addresses - should fail
		{name: "IPv4 localhost fails", value: "127.0.0.1", wantErr: true},
		{name: "IPv4 private fails", value: "192.168.1.1", wantErr: true},
		{name: "IPv4 google DNS fails", value: "8.8.8.8", wantErr: true},

		// Invalid IPv6 addresses
		{name: "invalid IPv6 - too many colons", value: "2001::db8:::1", wantErr: true},
		{name: "invalid IPv6 - invalid hex", value: "2001:db8::gggg", wantErr: true},
		{name: "invalid IPv6 - too many groups", value: "2001:0db8:85a3:0000:0000:8a2e:0370:7334:extra", wantErr: true},
		{name: "invalid IPv6 - incomplete", value: "2001:db8:", wantErr: true},

		// Other invalid formats
		{name: "hostname not IP", value: "example.com", wantErr: true},
		{name: "IPv6 with port fails", value: "[2001:db8::1]:8080", wantErr: true},

		// Nil pointer - should skip validation
		{name: "nil pointer", value: (*string)(nil), wantErr: false},

		// Invalid types
		{name: "invalid type - int", value: 123, wantErr: true},
		{name: "invalid type - bool", value: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint := ipv6Constraint{}
			err := constraint.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestIpConstraint tests ipConstraint.Validate() for any valid IP address (IPv4 or IPv6).
func TestIpConstraint(t *testing.T) {
	runSimpleConstraintTests(t, ipConstraint{}, []simpleTestCase{
		// Valid IPv4 addresses
		{"valid IPv4 - localhost", "127.0.0.1", false},
		{"valid IPv4 - private range", "192.168.1.1", false},
		{"valid IPv4 - public IP", "8.8.8.8", false},
		{"valid IPv4 - zeros", "0.0.0.0", false},
		{"valid IPv4 - broadcast", "255.255.255.255", false},
		// Valid IPv6 addresses
		{"valid IPv6 - localhost loopback", "::1", false},
		{"valid IPv6 - unspecified", "::", false},
		{"valid IPv6 - full form", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", false},
		{"valid IPv6 - compressed", "2001:db8::1", false},
		// Empty string - should be skipped (handled by required)
		{"empty string", "", false},
		// Invalid IP addresses
		{"invalid IP - not an IP", "not-an-ip", true},
		{"invalid IP - hostname", "example.com", true},
		{"invalid IP - CIDR notation", "192.168.1.0/24", true},
		{"invalid IP - out of range", "256.1.1.1", true},
		{"invalid IP - too few octets", "192.168.1", true},
		{"invalid IP - with port", "192.168.1.1:8080", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestCidrConstraint tests cidrConstraint.Validate() for valid CIDR notation (IPv4 or IPv6).
func TestCidrConstraint(t *testing.T) {
	runSimpleConstraintTests(t, cidrConstraint{}, []simpleTestCase{
		// Valid IPv4 CIDR
		{"valid IPv4 CIDR - /24", "192.168.1.0/24", false},
		{"valid IPv4 CIDR - /8", "10.0.0.0/8", false},
		{"valid IPv4 CIDR - /32", "192.168.1.1/32", false},
		{"valid IPv4 CIDR - /0", "0.0.0.0/0", false},
		{"valid IPv4 CIDR - /16", "172.16.0.0/16", false},
		// Valid IPv6 CIDR
		{"valid IPv6 CIDR - /32", "2001:db8::/32", false},
		{"valid IPv6 CIDR - /64", "2001:db8:abcd:0012::/64", false},
		{"valid IPv6 CIDR - /128", "::1/128", false},
		{"valid IPv6 CIDR - /0", "::/0", false},
		// Empty string - should be skipped (handled by required)
		{"empty string", "", false},
		// Invalid CIDR notation
		{"invalid CIDR - plain IP", "192.168.1.1", true},
		{"invalid CIDR - invalid prefix", "192.168.1.0/33", true},
		{"invalid CIDR - invalid IP", "256.1.1.0/24", true},
		{"invalid CIDR - missing prefix", "192.168.1.0/", true},
		{"invalid CIDR - not a number", "192.168.1.0/abc", true},
		{"invalid CIDR - hostname", "example.com/24", true},
		{"invalid CIDR - negative prefix", "192.168.1.0/-1", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestCidrv4Constraint tests cidrv4Constraint.Validate() for valid IPv4 CIDR notation.
func TestCidrv4Constraint(t *testing.T) {
	runSimpleConstraintTests(t, cidrv4Constraint{}, []simpleTestCase{
		// Valid IPv4 CIDR
		{"valid IPv4 CIDR - /24", "192.168.1.0/24", false},
		{"valid IPv4 CIDR - /8", "10.0.0.0/8", false},
		{"valid IPv4 CIDR - /32", "192.168.1.1/32", false},
		{"valid IPv4 CIDR - /0", "0.0.0.0/0", false},
		{"valid IPv4 CIDR - /16", "172.16.0.0/16", false},
		// Empty string - should be skipped (handled by required)
		{"empty string", "", false},
		// Invalid - IPv6 CIDR (should fail for v4-specific constraint)
		{"invalid - IPv6 CIDR fails", "2001:db8::/32", true},
		{"invalid - IPv6 CIDR /64", "2001:db8:abcd:0012::/64", true},
		// Invalid CIDR notation
		{"invalid CIDR - plain IP", "192.168.1.1", true},
		{"invalid CIDR - invalid prefix", "192.168.1.0/33", true},
		{"invalid CIDR - invalid IP", "256.1.1.0/24", true},
		{"invalid CIDR - not a number", "192.168.1.0/abc", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestCidrv6Constraint tests cidrv6Constraint.Validate() for valid IPv6 CIDR notation.
func TestCidrv6Constraint(t *testing.T) {
	runSimpleConstraintTests(t, cidrv6Constraint{}, []simpleTestCase{
		// Valid IPv6 CIDR
		{"valid IPv6 CIDR - /32", "2001:db8::/32", false},
		{"valid IPv6 CIDR - /64", "2001:db8:abcd:0012::/64", false},
		{"valid IPv6 CIDR - /128", "::1/128", false},
		{"valid IPv6 CIDR - /0", "::/0", false},
		{"valid IPv6 CIDR - full form", "2001:0db8:85a3:0000:0000:8a2e:0370:7334/64", false},
		// Empty string - should be skipped (handled by required)
		{"empty string", "", false},
		// Invalid - IPv4 CIDR (should fail for v6-specific constraint)
		{"invalid - IPv4 CIDR fails", "192.168.1.0/24", true},
		{"invalid - IPv4 CIDR /8", "10.0.0.0/8", true},
		// Invalid CIDR notation
		{"invalid CIDR - plain IPv6", "2001:db8::1", true},
		{"invalid CIDR - invalid prefix", "2001:db8::/129", true},
		{"invalid CIDR - invalid hex", "2001:gggg::/32", true},
		{"invalid CIDR - not a number", "2001:db8::/abc", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestMacConstraint tests macConstraint.Validate() for valid MAC addresses.
func TestMacConstraint(t *testing.T) {
	runSimpleConstraintTests(t, macConstraint{}, []simpleTestCase{
		// Valid MAC addresses - colon separated
		{"valid MAC - colon format", "00:1A:2B:3C:4D:5E", false},
		{"valid MAC - colon lowercase", "00:1a:2b:3c:4d:5e", false},
		{"valid MAC - colon mixed case", "00:1A:2b:3C:4d:5E", false},
		{"valid MAC - all zeros", "00:00:00:00:00:00", false},
		{"valid MAC - broadcast", "FF:FF:FF:FF:FF:FF", false},
		// Valid MAC addresses - hyphen separated
		{"valid MAC - hyphen format", "00-1A-2B-3C-4D-5E", false},
		{"valid MAC - hyphen lowercase", "00-1a-2b-3c-4d-5e", false},
		// Valid MAC addresses - Cisco format (dot separated, 4 hex digits per group)
		{"valid MAC - Cisco format", "001A.2B3C.4D5E", false},
		// Empty string - should be skipped (handled by required)
		{"empty string", "", false},
		// Invalid MAC addresses
		{"invalid MAC - not a MAC", "not-a-mac", true},
		{"invalid MAC - too short", "00:1A:2B:3C:4D", true},
		{"invalid MAC - too long", "00:1A:2B:3C:4D:5E:6F", true},
		{"invalid MAC - invalid hex", "00:GG:2B:3C:4D:5E", true},
		{"invalid MAC - mixed separators", "00:1A-2B:3C-4D:5E", true},
		{"invalid MAC - no separators", "001A2B3C4D5E", true},
		{"invalid MAC - IP address", "192.168.1.1", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestHostnameConstraint tests hostnameConstraint.Validate() for valid RFC 952 hostnames.
func TestHostnameConstraint(t *testing.T) {
	runSimpleConstraintTests(t, hostnameConstraint{}, []simpleTestCase{
		// Valid RFC 952 hostnames
		{"valid hostname - localhost", "localhost", false},
		{"valid hostname - simple", "myhost", false},
		{"valid hostname - with hyphen", "my-host", false},
		{"valid hostname - with numbers", "host123", false},
		{"valid hostname - uppercase", "MYHOST", false},
		{"valid hostname - mixed case", "MyHost", false},
		{"valid hostname - hyphen in middle", "my-server-01", false},
		// Empty string - should be skipped (handled by required)
		{"empty string", "", false},
		// Invalid RFC 952 hostnames
		{"invalid hostname - starts with digit", "123host", true},
		{"invalid hostname - starts with hyphen", "-myhost", true},
		{"invalid hostname - ends with hyphen", "myhost-", true},
		{"invalid hostname - underscore", "my_host", true},
		{"invalid hostname - dot", "my.host", true},
		{"invalid hostname - space", "my host", true},
		{"invalid hostname - special chars", "my@host", true},
		{"invalid hostname - too long", string(make([]byte, 256)), true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestHostnameRFC1123Constraint tests hostnameRFC1123Constraint.Validate() for valid RFC 1123 hostnames.
func TestHostnameRFC1123Constraint(t *testing.T) {
	runSimpleConstraintTests(t, hostnameRFC1123Constraint{}, []simpleTestCase{
		// Valid RFC 1123 hostnames (allows starting with digit)
		{"valid hostname - localhost", "localhost", false},
		{"valid hostname - simple", "myhost", false},
		{"valid hostname - with hyphen", "my-host", false},
		{"valid hostname - starts with digit", "123host", false},
		{"valid hostname - all digits", "12345", false},
		{"valid hostname - digit and hyphen", "1-host", false},
		{"valid hostname - uppercase", "MYHOST", false},
		{"valid hostname - mixed case", "MyHost123", false},
		// Empty string - should be skipped (handled by required)
		{"empty string", "", false},
		// Invalid RFC 1123 hostnames
		{"invalid hostname - starts with hyphen", "-myhost", true},
		{"invalid hostname - ends with hyphen", "myhost-", true},
		{"invalid hostname - underscore", "my_host", true},
		{"invalid hostname - dot", "my.host", true},
		{"invalid hostname - space", "my host", true},
		{"invalid hostname - special chars", "my@host", true},
		{"invalid hostname - too long", string(make([]byte, 256)), true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestFqdnConstraint tests fqdnConstraint.Validate() for valid fully qualified domain names.
func TestFqdnConstraint(t *testing.T) {
	runSimpleConstraintTests(t, fqdnConstraint{}, []simpleTestCase{
		// Valid FQDNs
		{"valid FQDN - simple", "example.com", false},
		{"valid FQDN - subdomain", "sub.example.com", false},
		{"valid FQDN - multiple subdomains", "a.b.c.example.com", false},
		{"valid FQDN - with hyphen", "my-site.example.com", false},
		{"valid FQDN - trailing dot", "example.com.", false},
		{"valid FQDN - mixed case", "Example.COM", false},
		{"valid FQDN - numbers in domain", "example123.com", false},
		{"valid FQDN - numbers in subdomain", "123.example.com", false},
		// Empty string - should be skipped (handled by required)
		{"empty string", "", false},
		// Invalid FQDNs
		{"invalid FQDN - no dot (single label)", "localhost", true},
		{"invalid FQDN - starts with dot", ".example.com", true},
		{"invalid FQDN - double dot", "example..com", true},
		{"invalid FQDN - hyphen at start of label", "-example.com", true},
		{"invalid FQDN - hyphen at end of label", "example-.com", true},
		{"invalid FQDN - underscore", "my_site.example.com", true},
		{"invalid FQDN - space", "my site.example.com", true},
		{"invalid FQDN - IP address", "192.168.1.1", true},
		{"invalid FQDN - label too long", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.com", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestPortConstraint tests portConstraint.Validate() for valid port numbers.
func TestPortConstraint(t *testing.T) {
	runSimpleConstraintTests(t, portConstraint{}, []simpleTestCase{
		// Valid port numbers
		{"valid port - 0", 0, false},
		{"valid port - 80", 80, false},
		{"valid port - 443", 443, false},
		{"valid port - 8080", 8080, false},
		{"valid port - 65535", 65535, false},
		{"valid port - 1", 1, false},
		{"valid port - 1024", 1024, false},
		// Valid port as int variants
		{"valid port - int32", int32(8080), false},
		{"valid port - int64", int64(8080), false},
		{"valid port - uint", uint(8080), false},
		{"valid port - uint16", uint16(8080), false},
		// Invalid port numbers
		{"invalid port - negative", -1, true},
		{"invalid port - too high", 65536, true},
		{"invalid port - way too high", 100000, true},
		// Invalid types
		{"invalid type - string", "8080", true},
		{"invalid type - float", 80.5, true},
		{"invalid type - bool", true, true},
		// Nil pointer - should skip validation
		{"nil pointer", (*int)(nil), false},
	})
}

// TestTcpAddrConstraint tests tcpAddrConstraint.Validate() for valid TCP addresses.
func TestTcpAddrConstraint(t *testing.T) {
	runSimpleConstraintTests(t, tcpAddrConstraint{}, []simpleTestCase{
		// Valid TCP addresses - hostname:port
		{"valid TCP addr - localhost:8080", "localhost:8080", false},
		{"valid TCP addr - hostname:80", "example.com:80", false},
		{"valid TCP addr - hostname:443", "www.example.com:443", false},
		// Valid TCP addresses - IPv4:port
		{"valid TCP addr - IPv4:port", "192.168.1.1:80", false},
		{"valid TCP addr - IPv4:8080", "127.0.0.1:8080", false},
		{"valid TCP addr - IPv4:443", "10.0.0.1:443", false},
		// Valid TCP addresses - [IPv6]:port
		{"valid TCP addr - IPv6:port", "[::1]:8080", false},
		{"valid TCP addr - IPv6 full:port", "[2001:db8::1]:80", false},
		// Empty string - should be skipped (handled by required)
		{"empty string", "", false},
		// Invalid TCP addresses
		{"invalid TCP addr - no port", "localhost", true},
		{"invalid TCP addr - no colon", "localhost8080", true},
		{"invalid TCP addr - empty port", "localhost:", true},
		{"invalid TCP addr - invalid port", "localhost:abc", true},
		{"invalid TCP addr - port too high", "localhost:65536", true},
		{"invalid TCP addr - negative port", "localhost:-1", true},
		{"invalid TCP addr - IPv6 without brackets", "::1:8080", true},
		{"invalid TCP addr - just port", ":8080", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestUdpAddrConstraint tests udpAddrConstraint.Validate() for valid UDP addresses.
func TestUdpAddrConstraint(t *testing.T) {
	runSimpleConstraintTests(t, udpAddrConstraint{}, []simpleTestCase{
		// Valid UDP addresses - hostname:port
		{"valid UDP addr - localhost:8080", "localhost:8080", false},
		{"valid UDP addr - hostname:53", "dns.example.com:53", false},
		// Valid UDP addresses - IPv4:port
		{"valid UDP addr - IPv4:port", "192.168.1.1:53", false},
		{"valid UDP addr - IPv4:8080", "127.0.0.1:8080", false},
		// Valid UDP addresses - [IPv6]:port
		{"valid UDP addr - IPv6:port", "[::1]:53", false},
		{"valid UDP addr - IPv6 full:port", "[2001:db8::1]:53", false},
		// Empty string - should be skipped (handled by required)
		{"empty string", "", false},
		// Invalid UDP addresses
		{"invalid UDP addr - no port", "localhost", true},
		{"invalid UDP addr - no colon", "localhost53", true},
		{"invalid UDP addr - empty port", "localhost:", true},
		{"invalid UDP addr - invalid port", "localhost:abc", true},
		{"invalid UDP addr - port too high", "localhost:65536", true},
		{"invalid UDP addr - IPv6 without brackets", "::1:53", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestTcp4AddrConstraint tests tcp4AddrConstraint.Validate() for valid IPv4 TCP addresses.
func TestTcp4AddrConstraint(t *testing.T) {
	runSimpleConstraintTests(t, tcp4AddrConstraint{}, []simpleTestCase{
		// Valid IPv4 TCP addresses
		{"valid TCP4 addr - IPv4:port", "192.168.1.1:80", false},
		{"valid TCP4 addr - localhost resolved", "127.0.0.1:8080", false},
		{"valid TCP4 addr - private range", "10.0.0.1:443", false},
		{"valid TCP4 addr - private range 172", "172.16.0.1:8080", false},
		// Empty string - should be skipped (handled by required)
		{"empty string", "", false},
		// Invalid - IPv6 addresses (should fail for v4-specific constraint)
		{"invalid TCP4 addr - IPv6 fails", "[::1]:8080", true},
		{"invalid TCP4 addr - IPv6 full fails", "[2001:db8::1]:80", true},
		// Invalid - hostname (only IP allowed for tcp4_addr)
		{"invalid TCP4 addr - hostname fails", "localhost:8080", true},
		{"invalid TCP4 addr - FQDN fails", "example.com:80", true},
		// Invalid TCP addresses
		{"invalid TCP4 addr - no port", "192.168.1.1", true},
		{"invalid TCP4 addr - empty port", "192.168.1.1:", true},
		{"invalid TCP4 addr - invalid port", "192.168.1.1:abc", true},
		{"invalid TCP4 addr - port too high", "192.168.1.1:65536", true},
		{"invalid TCP4 addr - invalid IP", "256.1.1.1:80", true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}
