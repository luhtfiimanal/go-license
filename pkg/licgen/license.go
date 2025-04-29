package licgen

import (
	"crypto/rsa"
	"fmt"
	"os"
	"time"

	"github.com/luhtfiimanal/go-license/pkg/licformat"
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

	// Convert the license to binary format
	licenseFormatObj := licformat.License{
		ID:           license.ID,
		CustomerID:   license.CustomerID,
		ProductID:    license.ProductID,
		SerialNumber: license.SerialNumber,
		IssueDate:    license.IssueDate,
		ExpiryDate:   license.ExpiryDate,
		Features:     license.Features,
		HardwareIDs: licformat.HardwareBinding{
			MACAddresses: license.HardwareIDs.MACAddresses,
			DiskIDs:      license.HardwareIDs.DiskIDs,
			HostNames:    license.HardwareIDs.HostNames,
			CustomIDs:    license.HardwareIDs.CustomIDs},
	}

	licenseData, err := licformat.EncodeLicense(&licenseFormatObj)
	if err != nil {
		return nil, fmt.Errorf("failed to encode license: %v", err)
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
