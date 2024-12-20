package env

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/amirrezaask/pkg/must"
	"github.com/amirrezaask/pkg/set"

	"github.com/joho/godotenv"
)

//TODO: Command to generate a list of all environment variables needed, so we can use it to generate a sample .env file.

var dotEnvMap = must.Do(godotenv.Unmarshal(".env"))

func getEnv(key string) string {
	// .env
	value := dotEnvMap[key]

	// os.Getenv
	if v := os.Getenv(key); v != "" {
		value = v
	}

	return value
}

func Default(key, def string) string {
	value := getEnv(key)
	if value == "" {
		return def
	}
	return value
}

func RequiredNotEmpty(key string) string {
	value := getEnv(key)
	if value == "" {
		if !testing.Testing() {
			panic(fmt.Sprintf("`%s` is not set or is empty", key))
		}
	}
	return value
}

func Required(key string) string {
	_, osSet := os.LookupEnv(key)
	_, dotEnvSet := dotEnvMap[key]
	if !osSet && !dotEnvSet {
		if !testing.Testing() {
			panic(fmt.Sprintf("`%s` is not set", key))
		}
	}
	return getEnv(key)
}

func ParseCommaSeperatedAsSet(input string) set.Set[string] {
	output := set.Set[string]{}
	segs := strings.Split(input, ",")
	for _, seg := range segs {
		output.Add(seg)
	}

	return output
}
