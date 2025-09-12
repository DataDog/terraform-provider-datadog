package fwutils

import (
	"log"
)

// Ephemeral resource utilities are now minimal - most functionality uses framework APIs directly.

// Private Data Management Constants
const (
	PrivateDataKeyResourceID   = "resource_id"
	PrivateDataKeyResourceType = "resource_type"
	PrivateDataKeyMetadata     = "metadata"
)

// Security-focused logging functions to avoid exposing sensitive data

// LogEphemeralOperation logs ephemeral resource operations without exposing sensitive data
func LogEphemeralOperation(operation, resourceType string, success bool) {
	if success {
		log.Printf("[DEBUG] Ephemeral %s operation succeeded for %s", operation, resourceType)
	} else {
		log.Printf("[WARN] Ephemeral %s operation failed for %s", operation, resourceType)
	}
}

// LogEphemeralError logs ephemeral resource errors without exposing sensitive data
func LogEphemeralError(operation, resourceType string, err error) {
	// Log error message but never sensitive data
	log.Printf("[ERROR] Ephemeral %s operation failed for %s: %v", operation, resourceType, err)
}
