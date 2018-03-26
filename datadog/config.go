package datadog

import (
	"log"

	"github.com/robinbowes/go-datadog-api"
)

// Config holds API and APP keys to authenticate to Datadog.
type Config struct {
	APIKey string
	APPKey string
}

// Client returns a new Datadog client.
func (c *Config) Client() *datadog.Client {

	client := datadog.NewClient(c.APIKey, c.APPKey)
	log.Printf("[INFO] Datadog Client configured ")

	return client
}
