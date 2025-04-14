package licverify_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/luhtfiimanal/go-license/pkg/licgen"
	"github.com/luhtfiimanal/go-license/pkg/licverify"
)

// TestVerifyLicense tests the license verification functionality
func TestVerifyLicense(t *testing.T) {
	// Generate a key pair for testing
	privateKeyPEM, publicKeyPEM, err := licgen.GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Parse the private key for signing
	privateKey, err := licgen.ParsePrivateKey(privateKeyPEM)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	// Get hardware info for binding
	hwInfo, err := licverify.GetHardwareInfo()
	if err != nil {
		t.Fatalf("Failed to get hardware info: %v", err)
	}

	// Create a test license
	license := licverify.License{
		ID:         "TEST-LICENSE-001",
		CustomerID: "CUSTOMER-001",
		ProductID:  "PRODUCT-001",
		IssueDate:  time.Now(),
		ExpiryDate: time.Now().AddDate(1, 0, 0), // 1 year from now
		Features:   []string{"feature1", "feature2", "premium"},
		HardwareIDs: licverify.HardwareBinding{
			// Use actual hardware info for testing
			MACAddresses: hwInfo.MACAddresses,
			DiskIDs:      hwInfo.DiskIDs,
			HostNames:    []string{hwInfo.Hostname},
		},
	}

	// Marshal the license to JSON
	licenseData, err := json.Marshal(license)
	if err != nil {
		t.Fatalf("Failed to marshal license: %v", err)
	}

	// Sign the license
	signature, err := licgen.SignData(licenseData, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign license: %v", err)
	}

	// Create a temporary directory for the test license file
	tempDir, err := os.MkdirTemp("", "license-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create the license file (JSON + signature)
	licenseFilePath := filepath.Join(tempDir, "license.lic")
	licenseFileData := append(licenseData, signature...)
	err = os.WriteFile(licenseFilePath, licenseFileData, 0644)
	if err != nil {
		t.Fatalf("Failed to write license file: %v", err)
	}

	// Create a verifier with the public key
	verifier, err := licverify.NewVerifier(publicKeyPEM)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	// Load the license from the file
	loadedLicense, err := verifier.LoadLicense(licenseFilePath)
	if err != nil {
		t.Fatalf("Failed to load license: %v", err)
	}

	// Verify the license
	err = loadedLicense.IsValid(verifier)
	if err != nil {
		t.Fatalf("License validation failed: %v", err)
	}

	// Verify individual components
	t.Run("VerifySignature", func(t *testing.T) {
		err := verifier.VerifySignature(loadedLicense)
		if err != nil {
			t.Errorf("Signature verification failed: %v", err)
		}
	})

	t.Run("VerifyHardwareBinding", func(t *testing.T) {
		err := verifier.VerifyHardwareBinding(loadedLicense)
		if err != nil {
			t.Errorf("Hardware binding verification failed: %v", err)
		}
	})

	t.Run("VerifyExpiry", func(t *testing.T) {
		err := verifier.VerifyExpiry(loadedLicense)
		if err != nil {
			t.Errorf("Expiry verification failed: %v", err)
		}
	})
}

// TestExpiredLicense tests that an expired license is correctly identified
func TestExpiredLicense(t *testing.T) {
	// Generate a key pair for testing
	privateKeyPEM, publicKeyPEM, err := licgen.GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Parse the private key for signing
	privateKey, err := licgen.ParsePrivateKey(privateKeyPEM)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	// Get hardware info for binding
	hwInfo, err := licverify.GetHardwareInfo()
	if err != nil {
		t.Fatalf("Failed to get hardware info: %v", err)
	}

	// Create an expired license
	license := licverify.License{
		ID:         "TEST-LICENSE-002",
		CustomerID: "CUSTOMER-001",
		ProductID:  "PRODUCT-001",
		IssueDate:  time.Now().AddDate(-2, 0, 0), // 2 years ago
		ExpiryDate: time.Now().AddDate(-1, 0, 0), // 1 year ago (expired)
		Features:   []string{"feature1", "feature2"},
		HardwareIDs: licverify.HardwareBinding{
			MACAddresses: hwInfo.MACAddresses,
			DiskIDs:      hwInfo.DiskIDs,
			HostNames:    []string{hwInfo.Hostname},
		},
	}

	// Marshal the license to JSON
	licenseData, err := json.Marshal(license)
	if err != nil {
		t.Fatalf("Failed to marshal license: %v", err)
	}

	// Sign the license
	signature, err := licgen.SignData(licenseData, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign license: %v", err)
	}

	// Create a temporary directory for the test license file
	tempDir, err := os.MkdirTemp("", "license-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create the license file (JSON + signature)
	licenseFilePath := filepath.Join(tempDir, "expired-license.lic")
	licenseFileData := append(licenseData, signature...)
	err = os.WriteFile(licenseFilePath, licenseFileData, 0644)
	if err != nil {
		t.Fatalf("Failed to write license file: %v", err)
	}

	// Create a verifier with the public key
	verifier, err := licverify.NewVerifier(publicKeyPEM)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	// Load the license from the file
	loadedLicense, err := verifier.LoadLicense(licenseFilePath)
	if err != nil {
		t.Fatalf("Failed to load license: %v", err)
	}

	// Verify the license (should fail due to expiry)
	err = loadedLicense.IsValid(verifier)
	if err == nil {
		t.Fatalf("Expected expired license to fail validation")
	}

	// Verify signature (should still pass)
	err = verifier.VerifySignature(loadedLicense)
	if err != nil {
		t.Errorf("Signature verification failed: %v", err)
	}

	// Verify expiry (should fail)
	err = verifier.VerifyExpiry(loadedLicense)
	if err == nil {
		t.Errorf("Expected expired license to fail expiry check")
	}
}
