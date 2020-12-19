package libraries

import (
	"fmt"
	"kwanjai/configuration"
	"log"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

// AccessSecretVersion function returns secret value (string) and
func AccessSecretVersion(name string) (string, error) {
	client, err := secretmanager.NewClient(configuration.Context)
	if err != nil {
		return "error", fmt.Errorf("failed to create secretmanager client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Fatalln(err.Error())
		}
	}()
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	result, err := client.AccessSecretVersion(configuration.Context, req)
	if err != nil {
		return "error", fmt.Errorf("failed to access secret version: %v", err)
	}

	return string(result.Payload.Data), nil
}
