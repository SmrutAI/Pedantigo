// Package constraints provides validation constraint types and builders for pedantigo.
package constraints

import (
	"fmt"
	"net"
	"reflect"
	"regexp"
	"strings"
)

// Hostname regex patterns (compiled once for performance).
var (
	// RFC 952: hostname must start with letter, contain only alphanumeric and hyphens.
	hostnameRFC952LabelRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9\-]*[a-zA-Z0-9]$|^[a-zA-Z]$`)
	// RFC 1123: same as RFC 952 but can start with digit.
	hostnameRFC1123LabelRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9]$|^[a-zA-Z0-9]$`)
)

// isValidPort checks if a port string represents a valid port number (0-65535).
func isValidPort(portStr string) bool {
	// Port must be numeric
	port := 0
	for _, c := range portStr {
		if c < '0' || c > '9' {
			return false
		}
		port = port*10 + int(c-'0')
		// Early exit if port exceeds max
		if port > 65535 {
			return false
		}
	}
	return port >= 0 && port <= 65535
}

// Network constraint types.
type (
	ipv4Constraint            struct{} // ipv4: validates IPv4 address
	ipv6Constraint            struct{} // ipv6: validates IPv6 address
	ipConstraint              struct{} // ip: validates any IPv4 or IPv6 address
	cidrConstraint            struct{} // cidr: validates any CIDR notation (IPv4 or IPv6)
	cidrv4Constraint          struct{} // cidrv4: validates IPv4 CIDR notation
	cidrv6Constraint          struct{} // cidrv6: validates IPv6 CIDR notation
	macConstraint             struct{} // mac: validates MAC address (net.ParseMAC)
	hostnameConstraint        struct{} // hostname: validates RFC 952 hostname
	hostnameRFC1123Constraint struct{} // hostname_rfc1123: validates RFC 1123 hostname (digits first OK)
	fqdnConstraint            struct{} // fqdn: validates fully qualified domain name
	portConstraint            struct{} // port: validates port number 0-65535 (integer)
	tcpAddrConstraint         struct{} // tcp_addr: validates TCP address (host:port)
	udpAddrConstraint         struct{} // udp_addr: validates UDP address (host:port)
	tcp4AddrConstraint        struct{} // tcp4_addr: validates IPv4 TCP address
)

// ipv4Constraint validates that a string is a valid IPv4 address.
func (c ipv4Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("ipv4 constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Parse IP address
	ip := net.ParseIP(str)
	if ip == nil {
		return NewConstraintError(CodeInvalidIPv4, "must be a valid IPv4 address")
	}

	// Check if it's IPv4 (not IPv6)
	// IPv4 addresses return non-nil from To4()
	if ip.To4() == nil {
		return NewConstraintError(CodeInvalidIPv4, "must be a valid IPv4 address")
	}

	return nil
}

// ipv6Constraint validates that a string is a valid IPv6 address.
func (c ipv6Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("ipv6 constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Parse IP address
	ip := net.ParseIP(str)
	if ip == nil {
		return NewConstraintError(CodeInvalidIPv6, "must be a valid IPv6 address")
	}

	// Check if it's IPv6 (not IPv4)
	// IPv6 addresses return nil from To4()
	if ip.To4() != nil {
		return NewConstraintError(CodeInvalidIPv6, "must be a valid IPv6 address")
	}

	return nil
}

// ipConstraint validates that a string is a valid IPv4 or IPv6 address.
func (c ipConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("ip constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Parse IP address
	ip := net.ParseIP(str)
	if ip == nil {
		return NewConstraintError(CodeInvalidIP, "must be a valid IP address")
	}

	return nil
}

// cidrConstraint validates that a string is a valid CIDR notation (IPv4 or IPv6).
func (c cidrConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("cidr constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Parse CIDR notation
	_, _, err = net.ParseCIDR(str)
	if err != nil {
		return NewConstraintError(CodeInvalidCIDR, "must be a valid CIDR notation")
	}

	return nil
}

// cidrv4Constraint validates that a string is a valid IPv4 CIDR notation.
func (c cidrv4Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("cidrv4 constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Parse CIDR notation
	ip, _, err := net.ParseCIDR(str)
	if err != nil {
		return NewConstraintError(CodeInvalidCIDR, "must be a valid IPv4 CIDR notation")
	}

	// Check if it's IPv4 (not IPv6)
	if ip.To4() == nil {
		return NewConstraintError(CodeInvalidCIDR, "must be a valid IPv4 CIDR notation")
	}

	return nil
}

// cidrv6Constraint validates that a string is a valid IPv6 CIDR notation.
func (c cidrv6Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("cidrv6 constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Parse CIDR notation
	ip, _, err := net.ParseCIDR(str)
	if err != nil {
		return NewConstraintError(CodeInvalidCIDR, "must be a valid IPv6 CIDR notation")
	}

	// Check if it's IPv6 (not IPv4)
	if ip.To4() != nil {
		return NewConstraintError(CodeInvalidCIDR, "must be a valid IPv6 CIDR notation")
	}

	return nil
}

// macConstraint validates that a string is a valid MAC address.
func (c macConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("mac constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Parse MAC address
	_, err = net.ParseMAC(str)
	if err != nil {
		return NewConstraintError(CodeInvalidMAC, "must be a valid MAC address")
	}

	return nil
}

// hostnameConstraint validates that a string is a valid RFC 952 hostname.
// A hostname is a single label (no dots) - for domain names use FQDN.
func (c hostnameConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("hostname constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// RFC 952 hostname validation
	// A hostname is a single label (no dots allowed)
	// Max 63 chars, must start with letter, contain only alphanumeric and hyphens
	if strings.Contains(str, ".") {
		return NewConstraintError(CodeInvalidHostname, "must be a valid RFC 952 hostname")
	}

	if len(str) > 63 {
		return NewConstraintError(CodeInvalidHostname, "must be a valid RFC 952 hostname")
	}

	if !hostnameRFC952LabelRegex.MatchString(str) {
		return NewConstraintError(CodeInvalidHostname, "must be a valid RFC 952 hostname")
	}

	return nil
}

// hostnameRFC1123Constraint validates that a string is a valid RFC 1123 hostname.
// A hostname is a single label (no dots) - for domain names use FQDN.
func (c hostnameRFC1123Constraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("hostname_rfc1123 constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// RFC 1123 hostname validation
	// A hostname is a single label (no dots allowed)
	// Same as RFC 952 but can start with digit
	// Max 63 chars
	if strings.Contains(str, ".") {
		return NewConstraintError(CodeInvalidHostname, "must be a valid RFC 1123 hostname")
	}

	if len(str) > 63 {
		return NewConstraintError(CodeInvalidHostname, "must be a valid RFC 1123 hostname")
	}

	if !hostnameRFC1123LabelRegex.MatchString(str) {
		return NewConstraintError(CodeInvalidHostname, "must be a valid RFC 1123 hostname")
	}

	return nil
}

// fqdnConstraint validates that a string is a valid fully qualified domain name.
func (c fqdnConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("fqdn constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Reject IP addresses - FQDNs must be domain names, not IPs
	if net.ParseIP(str) != nil {
		return NewConstraintError(CodeInvalidFQDN, "must be a valid FQDN")
	}

	// FQDN must contain at least one dot (to distinguish from hostname)
	// Remove trailing dot if present (valid FQDN notation)
	fqdn := strings.TrimSuffix(str, ".")
	if !strings.Contains(fqdn, ".") {
		return NewConstraintError(CodeInvalidFQDN, "must be a valid FQDN")
	}

	// Max 253 chars total
	if len(fqdn) > 253 {
		return NewConstraintError(CodeInvalidFQDN, "must be a valid FQDN")
	}

	// Each label follows hostname rules (RFC 1123 for broader compatibility)
	labels := strings.Split(fqdn, ".")
	for _, label := range labels {
		if label == "" || len(label) > 63 {
			return NewConstraintError(CodeInvalidFQDN, "must be a valid FQDN")
		}
		if !hostnameRFC1123LabelRegex.MatchString(label) {
			return NewConstraintError(CodeInvalidFQDN, "must be a valid FQDN")
		}
	}

	return nil
}

// portConstraint validates that a value is a valid port number (0-65535).
// Only integer types are accepted; floats and strings are rejected.
func (c portConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // skip nil values
	}

	// Port must be an integer type, not float or string
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		num := v.Int()
		if num < 0 || num > 65535 {
			return NewConstraintError(CodeInvalidPort, "must be a valid port number (0-65535)")
		}
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		num := v.Uint()
		if num > 65535 {
			return NewConstraintError(CodeInvalidPort, "must be a valid port number (0-65535)")
		}
		return nil
	default:
		return NewConstraintError(CodeInvalidPort, "port constraint requires integer value")
	}
}

// tcpAddrConstraint validates that a string is a valid TCP address (host:port).
// This validates the format, not that the address actually resolves.
func (c tcpAddrConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("tcp_addr constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Check host:port format and ensure both host and port are not empty
	host, portStr, splitErr := net.SplitHostPort(str)
	if splitErr != nil {
		return NewConstraintError(CodeInvalidTCPAddr, "must be a valid TCP address")
	}
	if host == "" || portStr == "" {
		return NewConstraintError(CodeInvalidTCPAddr, "must be a valid TCP address")
	}

	// Validate the port is a valid number in range
	if !isValidPort(portStr) {
		return NewConstraintError(CodeInvalidTCPAddr, "must be a valid TCP address")
	}

	return nil
}

// udpAddrConstraint validates that a string is a valid UDP address (host:port).
// This validates the format, not that the address actually resolves.
func (c udpAddrConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("udp_addr constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Check host:port format and ensure both host and port are not empty
	host, portStr, splitErr := net.SplitHostPort(str)
	if splitErr != nil {
		return NewConstraintError(CodeInvalidUDPAddr, "must be a valid UDP address")
	}
	if host == "" || portStr == "" {
		return NewConstraintError(CodeInvalidUDPAddr, "must be a valid UDP address")
	}

	// Validate the port is a valid number in range
	if !isValidPort(portStr) {
		return NewConstraintError(CodeInvalidUDPAddr, "must be a valid UDP address")
	}

	return nil
}

// tcp4AddrConstraint validates that a string is a valid IPv4 TCP address.
// Unlike tcpAddrConstraint, this only accepts literal IPv4 addresses, not hostnames.
func (c tcp4AddrConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("tcp4_addr constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Parse host:port
	host, portStr, err := net.SplitHostPort(str)
	if err != nil {
		return NewConstraintError(CodeInvalidTCPAddr, "must be a valid IPv4 TCP address")
	}

	// Port must not be empty and must be valid
	if portStr == "" || !isValidPort(portStr) {
		return NewConstraintError(CodeInvalidTCPAddr, "must be a valid IPv4 TCP address")
	}

	// Host must be a valid IPv4 address (not hostname)
	ip := net.ParseIP(host)
	if ip == nil {
		return NewConstraintError(CodeInvalidTCPAddr, "must be a valid IPv4 TCP address")
	}

	// Must be IPv4, not IPv6
	if ip.To4() == nil {
		return NewConstraintError(CodeInvalidTCPAddr, "must be a valid IPv4 TCP address")
	}

	return nil
}
