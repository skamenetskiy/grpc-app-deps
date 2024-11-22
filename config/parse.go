package config

import (
	"github.com/kelseyhightower/envconfig"
)

func Parse(v any) error {
	return ParseWithPrefix("", v)
}

func ParseWithPrefix(prefix string, v any) error {
	return envconfig.Process(prefix, v)
}
