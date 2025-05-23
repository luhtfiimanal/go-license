package licverify

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/luhtfiimanal/go-license/v2/pkg/licformat"
)

// License represents a software license with hardware binding and expiration
type License struct {
	// Core license data
	ID           string    `json:"id"`
	CustomerID   string    `json:"customer_id"`
	ProductID    string    `json:"product_id"`
	SerialNumber string    `json:"serial_number"`
	IssueDate    time.Time `json:"issue_date"`
	ExpiryDate   time.Time `json:"expiry_date"`
	Features     []string  `json:"features"`

	// Hardware binding data
	HardwareIDs HardwareBinding `json:"hardware_ids"`

	// Signature is stored separately and not included in the JSON for signature verification
	Signature []byte `json:"-"`
}

// HardwareBinding contains hardware identifiers for license binding
type HardwareBinding struct {
	MACAddresses []string `json:"mac_addresses,omitempty"`
	DiskIDs      []string `json:"disk_ids,omitempty"`
	HostNames    []string `json:"host_names,omitempty"`
	CustomIDs    []string `json:"custom_ids,omitempty"`
}

// Verifier handles license verification
type Verifier struct {
	publicKey *rsa.PublicKey
}

// NewVerifier creates a new license verifier with the provided public key
func NewVerifier(publicKeyPEM string) (*Verifier, error) {
	if publicKeyPEM == "" {
		return nil, errors.New("public key cannot be empty")
	}

	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return &Verifier{
		publicKey: rsaPub,
	}, nil
}

// LoadLicense loads a license from the specified file path
func (v *Verifier) LoadLicense(filePath string) (*License, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read license file: %v", err)
	}

	// License file format: Binary data followed by signature
	// The last 256 bytes (for RSA-2048) are the signature
	if len(data) < 256 {
		return nil, errors.New("license file too small")
	}

	licenseData := data[:len(data)-256]
	signature := data[len(data)-256:]

	// Try binary format first (v2.0.0+)
	importedLicense, err := licformat.DecodeLicense(licenseData)
	if err != nil {
		// Fallback to legacy JSON format (v1.x)
		var license License
		if jsonErr := json.Unmarshal(licenseData, &license); jsonErr != nil {
			return nil, fmt.Errorf("failed to parse license data: %v (binary format error: %v)", jsonErr, err)
		}
		// Successfully parsed as legacy JSON format
		license.Signature = signature
		return &license, nil
	}

	// Convert the imported license to our internal format
	license := License{
		ID:           importedLicense.ID,
		CustomerID:   importedLicense.CustomerID,
		ProductID:    importedLicense.ProductID,
		SerialNumber: importedLicense.SerialNumber,
		IssueDate:    importedLicense.IssueDate,
		ExpiryDate:   importedLicense.ExpiryDate,
		Features:     importedLicense.Features,
		HardwareIDs: HardwareBinding{
			MACAddresses: importedLicense.HardwareIDs.MACAddresses,
			DiskIDs:      importedLicense.HardwareIDs.DiskIDs,
			HostNames:    importedLicense.HardwareIDs.HostNames,
			CustomIDs:    importedLicense.HardwareIDs.CustomIDs,
		},
	}

	license.Signature = signature
	return &license, nil
}

// VerifySignature verifies the digital signature of the license
func (v *Verifier) VerifySignature(license *License) error {
	// Create a copy of the license without the signature for verification
	licenseCopy := *license
	licenseCopy.Signature = nil

	// Convert to binary format for verification
	licenseFormatObj := licformat.License{
		ID:           licenseCopy.ID,
		CustomerID:   licenseCopy.CustomerID,
		ProductID:    licenseCopy.ProductID,
		SerialNumber: licenseCopy.SerialNumber,
		IssueDate:    licenseCopy.IssueDate,
		ExpiryDate:   licenseCopy.ExpiryDate,
		Features:     licenseCopy.Features,
		HardwareIDs: licformat.HardwareBinding{
			MACAddresses: licenseCopy.HardwareIDs.MACAddresses,
			DiskIDs:      licenseCopy.HardwareIDs.DiskIDs,
			HostNames:    licenseCopy.HardwareIDs.HostNames,
			CustomIDs:    licenseCopy.HardwareIDs.CustomIDs,
		},
	}

	// Try binary format first
	licenseData, err := licformat.EncodeLicense(&licenseFormatObj)
	if err != nil {
		return fmt.Errorf("failed to encode license data: %v", err)
	}

	// Calculate the hash of the license data
	hashed := sha256.Sum256(licenseData)

	// Verify the signature
	err = rsa.VerifyPKCS1v15(v.publicKey, crypto.SHA256, hashed[:], license.Signature)
	if err != nil {
		// Try JSON format for backward compatibility
		jsonData, jsonErr := json.Marshal(licenseCopy)
		if jsonErr != nil {
			return fmt.Errorf("failed to marshal license data: %v", jsonErr)
		}

		// Calculate the hash of the JSON data
		hashedJSON := sha256.Sum256(jsonData)

		// Verify the signature with JSON data
		jsonVerifyErr := rsa.VerifyPKCS1v15(v.publicKey, crypto.SHA256, hashedJSON[:], license.Signature)
		if jsonVerifyErr != nil {
			return fmt.Errorf("invalid license signature (tried both binary and JSON formats): %v", err)
		}
	}

	return nil
}

// VerifyHardwareBinding verifies that the license is bound to the current hardware
func (v *Verifier) VerifyHardwareBinding(license *License) error {
	// Get hardware info
	hwInfo, err := GetHardwareInfo()
	if err != nil {
		return fmt.Errorf("failed to get hardware info: %v", err)
	}

	// Verify MAC addresses
	if len(license.HardwareIDs.MACAddresses) > 0 {
		if !containsAny(hwInfo.MACAddresses, license.HardwareIDs.MACAddresses) {
			return errors.New("license is not valid for this hardware (MAC address mismatch)")
		}
	}

	// Verify disk IDs if present
	if len(license.HardwareIDs.DiskIDs) > 0 {
		if !containsAny(hwInfo.DiskIDs, license.HardwareIDs.DiskIDs) {
			return errors.New("license is not valid for this hardware (disk ID mismatch)")
		}
	}

	// Verify hostname if present
	if len(license.HardwareIDs.HostNames) > 0 {
		if !contains(license.HardwareIDs.HostNames, hwInfo.Hostname) {
			return errors.New("license is not valid for this hardware (hostname mismatch)")
		}
	}

	return nil
}

// VerifyExpiry checks if the license has expired
func (v *Verifier) VerifyExpiry(license *License) error {
	now := time.Now()
	if now.After(license.ExpiryDate) {
		return fmt.Errorf("license expired on %s", license.ExpiryDate.Format(time.RFC3339))
	}
	return nil
}

// IsValid performs all verification checks on the license
func (license *License) IsValid(verifier *Verifier) error {
	// Verify signature
	if err := verifier.VerifySignature(license); err != nil {
		return err
	}

	// Verify hardware binding
	if err := verifier.VerifyHardwareBinding(license); err != nil {
		return err
	}

	// Verify expiry
	if err := verifier.VerifyExpiry(license); err != nil {
		return err
	}

	return nil
}

// Helper functions

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// containsAny checks if any string from the first slice is in the second slice
func containsAny(slice1, slice2 []string) bool {
	for _, s1 := range slice1 {
		for _, s2 := range slice2 {
			if s1 == s2 {
				return true
			}
		}
	}
	return false
}
