package env

import (
	"log"
	"os"
	"sni/cmd/sni/config"
	"strings"
)

func GetOrDefault(name string, defaultValue string) (value string) {
	value = os.Getenv(name)
	if value == "" {
		value = defaultValue
		log.Printf("Read env var %s: not found; defaulting to '%s'\n", name, value)
	} else {
		log.Printf("Read env var %s: using '%s'\n", name, value)
	}
	return
}

func GetOrSupply(name string, defaultValueSupplier func() string) (value string) {
	value = config.Config.GetString(name)
	if value == "" {
		value = defaultValueSupplier()
		log.Printf("Read env var SNI_%s: not found; defaulting to '%s'\n", strings.ToUpper(name), value)
	} else {
		log.Printf("Read env var %s: using '%s'\n", name, value)
	}
	return
}
