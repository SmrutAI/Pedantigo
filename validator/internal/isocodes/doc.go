// Package isocodes provides validation for ISO standard codes including
// country codes (ISO 3166-1), currency codes (ISO 4217), and postal codes.
//
// This package contains data derived from go-playground/validator
// (https://github.com/go-playground/validator) under the MIT License.
// See the LICENSE file in this directory for full license text.
//
// Supported validations:
//   - ISO 3166-1 alpha-2 country codes (e.g., "US", "GB", "DE")
//   - ISO 3166-1 alpha-3 country codes (e.g., "USA", "GBR", "DEU")
//   - ISO 3166-1 numeric country codes (e.g., 840, 826, 276)
//   - ISO 3166-2 subdivision codes (e.g., "US-CA", "GB-ENG")
//   - ISO 4217 currency codes (e.g., "USD", "EUR", "GBP")
//   - ISO 4217 numeric currency codes (e.g., 840, 978, 826)
//   - Postal codes for ~120 countries
package isocodes
