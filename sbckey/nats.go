package sbckey

import (
	"fmt"
	"regexp"

	sbc "github.com/Autodoc-Technology/streaming-based-config"
)

var (
	// validKeyRe is a regular expression variable used to match strings that consist
	// of alphanumeric characters, along with a few special characters (-, _, /, =, .).
	// The regular expression ensures that the string does not contain any other
	// characters.
	validKeyRe = regexp.MustCompile(`^[-/_=.a-zA-Z0-9]+$`)

	// replaceRe is a regular expression variable used to match strings that consist
	replaceRe = regexp.MustCompile(`[^-/_=.a-zA-Z0-9]`)
)

// NatsDefaultKeyBuilder returns a KeyBuilder function that formats the type of the input value as a string.
// NATS key should be aligned to following naming convention https://docs.nats.io/nats-concepts/subjects#characters-allowed-for-subject-names
//
// Allowed characters Any ASCII character except null and .,* and >
// Recommended characters: a to z, A to Z and 0 to 9 and - (names are case-sensitive, and cannot contain whitespace).
func NatsDefaultKeyBuilder[T any]() sbc.KeyBuilder[T] {
	return sbc.KeyBuilderFunc[T](func(t T) string {
		val := fmt.Sprintf("%T", t)
		if natsKeyValid(val) {
			return val
		}
		// replace invalid characters with _
		val = replaceRe.ReplaceAllString(val, "_")
		return val
	})
}

// natsKeyValid checks if the key is valid for NATS
func natsKeyValid(key string) bool {
	if len(key) == 0 || key[0] == '.' || key[len(key)-1] == '.' {
		return false
	}
	return validKeyRe.MatchString(key)
}
