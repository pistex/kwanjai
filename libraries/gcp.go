package libraries

import (
	"kwanjai/configuration"
	"log"
	"os"
)

// InitializeGCP from credential.json.
func InitializeGCP() {
	if configuration.BaseDirectory == "" {
		log.Fatalln("configuration.BaseDirectory has been set to be empty string")
	}
	if os.Getenv("GIN_MODE") == "release" {
		return
	}
	defaultCredential := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if defaultCredential == "" {
		if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", configuration.BaseDirectory+"/.secret/credential.json"); err != nil {
			log.Fatalln(err.Error())
		}
	}
}
