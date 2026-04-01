package sbckey

import (
	"fmt"
	"regexp"

	sbc "github.com/Autodoc-Technology/streaming-based-config"
)

// maxKafkaTopicLen is the maximum allowed Kafka topic name length.
const maxKafkaTopicLen = 249

var (
	// kafkaValidRe matches valid Kafka topic names (alphanumeric, dot, underscore, hyphen).
	kafkaValidRe = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

	// kafkaReplaceRe matches characters that are not allowed in Kafka topic names.
	kafkaReplaceRe = regexp.MustCompile(`[^a-zA-Z0-9._-]`)
)

// KafkaDefaultKeyBuilder returns a KeyBuilder that derives the Kafka topic name
// from the Go type name of the value. It sanitizes the name to conform to Kafka
// topic naming rules: allowed characters are a-z, A-Z, 0-9, dot (.), underscore (_)
// and hyphen (-), with a maximum length of 249 characters.
//
// See: https://kafka.apache.org/documentation/#multitenancy-topic-naming
func KafkaDefaultKeyBuilder[T any]() sbc.KeyBuilder[T] {
	return sbc.KeyBuilderFunc[T](func(t T) string {
		val := fmt.Sprintf("%T", t)
		if kafkaKeyValid(val) {
			return truncateKafkaKey(val)
		}
		val = kafkaReplaceRe.ReplaceAllString(val, "_")
		return truncateKafkaKey(val)
	})
}

// kafkaKeyValid reports whether key is a valid Kafka topic name.
func kafkaKeyValid(key string) bool {
	if len(key) == 0 || len(key) > maxKafkaTopicLen {
		return false
	}
	return kafkaValidRe.MatchString(key)
}

// truncateKafkaKey truncates key to maxKafkaTopicLen characters.
func truncateKafkaKey(key string) string {
	if len(key) > maxKafkaTopicLen {
		return key[:maxKafkaTopicLen]
	}
	return key
}
