package licgen

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/luhtfiimanal/go-license/pkg/licverify"
)

// GenerateLicense creates a new license with the provided parameters and signs it
func GenerateLicense(
	id string,
	customerID string,
	productID string,
	serialNumber string,
	expiryDuration time.Duration,
	features []string,
	hardwareIDs licverify.HardwareBinding,
	privateKey *rsa.PrivateKey,
) ([]byte, error) {
	// Create the license
	license := licverify.License{
		ID:           id,
		CustomerID:   customerID,
		ProductID:    productID,
		SerialNumber: serialNumber,
		IssueDate:    time.Now(),
		ExpiryDate:   time.Now().Add(expiryDuration),
		Features:     features,
		HardwareIDs:  hardwareIDs,
	}

	// Marshal the license to JSON
	licenseData, err := json.Marshal(license)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal license: %v", err)
	}

	// Sign the license
	signature, err := SignData(licenseData, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign license: %v", err)
	}

	// Combine license data and signature
	licenseFile := append(licenseData, signature...)
	return licenseFile, nil
}

// SaveLicenseToFile saves a license to a file
func SaveLicenseToFile(licenseData []byte, filePath string) error {
	return os.WriteFile(filePath, licenseData, 0644)
}
