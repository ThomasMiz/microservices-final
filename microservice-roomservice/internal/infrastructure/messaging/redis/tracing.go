package redis

import (
	"go.opentelemetry.io/otel/propagation"
)

// RedisMessageCarrier implements the TextMapCarrier interface for Redis Stream messages
type RedisMessageCarrier struct {
	values map[string]interface{}
}

// Get returns the value associated with the passed key
func (c *RedisMessageCarrier) Get(key string) string {
	if val, ok := c.values[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// Set stores the key-value pair
func (c *RedisMessageCarrier) Set(key string, value string) {
	c.values[key] = value
}

// Keys lists the keys stored in this carrier
func (c *RedisMessageCarrier) Keys() []string {
	keys := make([]string, 0, len(c.values))
	for k := range c.values {
		keys = append(keys, k)
	}
	return keys
}

// Ensure RedisMessageCarrier implements TextMapCarrier
var _ propagation.TextMapCarrier = (*RedisMessageCarrier)(nil)
