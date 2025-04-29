package pkg

import (
	"testing"
	"time"

	"github.com/luhtfiimanal/go-license/pkg/licgen"
	"github.com/luhtfiimanal/go-license/pkg/licverify"
)

func TestBinaryLicenseIntegration(t *testing.T) {
	// Generate a key pair for testing
	privateKeyPEM, publicKeyPEM, err := licgen.GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Parse the private key
	privateKey, err := licgen.ParsePrivateKey(privateKeyPEM)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	// Create a license verifier
	verifier, err := licverify.NewVerifier(publicKeyPEM)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	// Create a test license
	licenseID := "test-license-binary-123"
	customerID := "customer-456"
	productID := "product-789"
	serialNumber := "SN-ABCDEF"
	expiryDuration := 365 * 24 * time.Hour // 1 year
	features := []string{"feature1", "feature2", "feature3"}
	hardwareIDs := licverify.HardwareBinding{
		MACAddresses: []string{"00:11:22:33:44:55"},
		HostNames:    []string{"localhost"},
	}

	// Generate the license
	licenseData, err := licgen.GenerateLicense(
		licenseID,
		customerID,
		productID,
		serialNumber,
		expiryDuration,
		features,
		hardwareIDs,
		privateKey,
	)
	if err != nil {
		t.Fatalf("Failed to generate license: %v", err)
	}

	// Create a temporary file for the license
	tempFile := t.TempDir() + "/test-license.bin"

	// Save the license to a file
	err = licgen.SaveLicenseToFile(licenseData, tempFile)
	if err != nil {
		t.Fatalf("Failed to save license to file: %v", err)
	}

	// Load the license from the file
	license, err := verifier.LoadLicense(tempFile)
	if err != nil {
		t.Fatalf("Failed to load license: %v", err)
	}

	// Verify the license fields
	if license.ID != licenseID {
		t.Errorf("License ID mismatch: expected %s, got %s", licenseID, license.ID)
	}
	if license.CustomerID != customerID {
		t.Errorf("Customer ID mismatch: expected %s, got %s", customerID, license.CustomerID)
	}
	if license.ProductID != productID {
		t.Errorf("Product ID mismatch: expected %s, got %s", productID, license.ProductID)
	}
	if license.SerialNumber != serialNumber {
		t.Errorf("Serial number mismatch: expected %s, got %s", serialNumber, license.SerialNumber)
	}

	// Verify the license signature
	err = verifier.VerifySignature(license)
	if err != nil {
		t.Errorf("Signature verification failed: %v", err)
	}

	// Verify the license expiry
	err = verifier.VerifyExpiry(license)
	if err != nil {
		t.Errorf("Expiry verification failed: %v", err)
	}

	// Skip hardware binding verification in tests
	// Just verify signature and expiry separately
	err = verifier.VerifySignature(license)
	if err != nil {
		t.Errorf("Signature verification failed: %v", err)
	}

	err = verifier.VerifyExpiry(license)
	if err != nil {
		t.Errorf("Expiry verification failed: %v", err)
	}
}

func TestBinaryLicenseTampering(t *testing.T) {
	// Generate a key pair for testing
	privateKeyPEM, publicKeyPEM, err := licgen.GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Parse the private key
	privateKey, err := licgen.ParsePrivateKey(privateKeyPEM)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	// Create a license verifier
	verifier, err := licverify.NewVerifier(publicKeyPEM)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	// Create a test license
	licenseID := "test-license-binary-456"
	customerID := "customer-789"
	productID := "product-123"
	serialNumber := "SN-XYZABC"
	expiryDuration := 365 * 24 * time.Hour // 1 year
	features := []string{"feature1", "feature2", "feature3"}
	hardwareIDs := licverify.HardwareBinding{
		MACAddresses: []string{"00:11:22:33:44:55"},
		HostNames:    []string{"localhost"},
	}

	// Generate the license
	licenseData, err := licgen.GenerateLicense(
		licenseID,
		customerID,
		productID,
		serialNumber,
		expiryDuration,
		features,
		hardwareIDs,
		privateKey,
	)
	if err != nil {
		t.Fatalf("Failed to generate license: %v", err)
	}

	// Create a temporary file for the license
	tempFile := t.TempDir() + "/test-license-tampered.bin"

	// Tamper with the license data (change a byte in the middle)
	midPoint := len(licenseData) / 2
	licenseData[midPoint] = licenseData[midPoint] ^ 0xFF // Flip all bits

	// Save the tampered license to a file
	err = licgen.SaveLicenseToFile(licenseData, tempFile)
	if err != nil {
		t.Fatalf("Failed to save license to file: %v", err)
	}

	// Try to load the tampered license
	license, err := verifier.LoadLicense(tempFile)
	if err == nil {
		// If we can load it, the signature should fail
		err = verifier.VerifySignature(license)
		if err == nil {
			t.Errorf("Expected signature verification to fail for tampered license, but it succeeded")
		}
	}
}
