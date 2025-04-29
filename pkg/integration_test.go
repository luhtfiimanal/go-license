package pkg_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/luhtfiimanal/go-license/v2/pkg/licgen"
	"github.com/luhtfiimanal/go-license/v2/pkg/licverify"
)

// TestFullLicenseFlow tests the complete license generation and verification flow
func TestFullLicenseFlow(t *testing.T) {
	// 1. Generate a key pair
	privateKeyPEM, publicKeyPEM, err := licgen.GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// 2. Parse the private key for signing
	privateKey, err := licgen.ParsePrivateKey(privateKeyPEM)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	// 3. Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "license-flow-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 4. Get actual hardware info for realistic testing
	hwInfo, err := licverify.GetHardwareInfo()
	if err != nil {
		t.Fatalf("Failed to get hardware info: %v", err)
	}

	// 5. Create a hardware binding with the current machine's info
	hwBinding := licverify.HardwareBinding{
		MACAddresses: hwInfo.MACAddresses,
		DiskIDs:      hwInfo.DiskIDs,
		HostNames:    []string{hwInfo.Hostname},
	}

	// 6. Generate a license
	licenseID := "INTEGRATION-TEST-LICENSE"
	customerID := "INTEGRATION-TEST-CUSTOMER"
	productID := "INTEGRATION-TEST-PRODUCT"
	serialNumber := "INTEGRATION-TEST-SERIAL"
	features := []string{"basic", "premium", "enterprise"}
	expiryDuration := 365 * 24 * time.Hour // 1 year

	licenseData, err := licgen.GenerateLicense(
		licenseID,
		customerID,
		productID,
		serialNumber,
		expiryDuration,
		features,
		hwBinding,
		privateKey,
	)
	if err != nil {
		t.Fatalf("Failed to generate license: %v", err)
	}

	// 7. Save the license to a file
	licenseFilePath := filepath.Join(tempDir, "integration-test-license.lic")
	err = licgen.SaveLicenseToFile(licenseData, licenseFilePath)
	if err != nil {
		t.Fatalf("Failed to save license: %v", err)
	}

	// 8. Create a verifier with the public key
	verifier, err := licverify.NewVerifier(publicKeyPEM)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	// 9. Load the license from the file
	license, err := verifier.LoadLicense(licenseFilePath)
	if err != nil {
		t.Fatalf("Failed to load license: %v", err)
	}

	// 10. Verify the license
	err = license.IsValid(verifier)
	if err != nil {
		t.Fatalf("License validation failed: %v", err)
	}

	// 11. Verify license data
	if license.ID != licenseID {
		t.Errorf("License ID mismatch: expected %s, got %s", licenseID, license.ID)
	}
	if license.CustomerID != customerID {
		t.Errorf("Customer ID mismatch: expected %s, got %s", customerID, license.CustomerID)
	}
	if license.ProductID != productID {
		t.Errorf("Product ID mismatch: expected %s, got %s", productID, license.ProductID)
	}

	// 12. Verify features
	if len(license.Features) != len(features) {
		t.Errorf("Feature count mismatch: expected %d, got %d", len(features), len(license.Features))
	}
	for i, feature := range features {
		if i < len(license.Features) && license.Features[i] != feature {
			t.Errorf("Feature mismatch at index %d: expected %s, got %s", i, feature, license.Features[i])
		}
	}

	// 13. Test tampering detection
	t.Run("TamperingDetection", func(t *testing.T) {
		// Create a copy of the license with a modified expiry date
		tamperedLicense := *license
		tamperedLicense.ExpiryDate = time.Now().AddDate(10, 0, 0) // 10 years in the future

		// Verify the tampered license (should fail)
		err = tamperedLicense.IsValid(verifier)
		if err == nil {
			t.Errorf("Tampered license validation should have failed")
		}
	})

	// 14. Test individual verification components
	t.Run("SignatureVerification", func(t *testing.T) {
		err := verifier.VerifySignature(license)
		if err != nil {
			t.Errorf("Signature verification failed: %v", err)
		}
	})

	t.Run("HardwareBindingVerification", func(t *testing.T) {
		err := verifier.VerifyHardwareBinding(license)
		if err != nil {
			t.Errorf("Hardware binding verification failed: %v", err)
		}
	})

	t.Run("ExpiryVerification", func(t *testing.T) {
		err := verifier.VerifyExpiry(license)
		if err != nil {
			t.Errorf("Expiry verification failed: %v", err)
		}
	})
}
