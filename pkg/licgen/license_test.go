package licgen_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/luhtfiimanal/go-license/pkg/licgen"
	"github.com/luhtfiimanal/go-license/pkg/licverify"
)

// TestGenerateLicense tests the license generation functionality
func TestGenerateLicense(t *testing.T) {
	// Generate a key pair for testing
	privateKeyPEM, _, err := licgen.GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Parse the private key for signing
	privateKey, err := licgen.ParsePrivateKey(privateKeyPEM)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	// Create a temporary directory for the test license file
	tempDir, err := os.MkdirTemp("", "license-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test hardware binding
	hwBinding := licverify.HardwareBinding{
		MACAddresses: []string{"00:11:22:33:44:55"},
		DiskIDs:      []string{"test-disk-id"},
		HostNames:    []string{"test-hostname"},
	}

	// Generate a license
	licenseData, err := licgen.GenerateLicense(
		"TEST-LICENSE-001",
		"CUSTOMER-001",
		"PRODUCT-001",
		"SERIAL-001",
		365*24*time.Hour, // 1 year
		[]string{"feature1", "feature2"},
		hwBinding,
		privateKey,
	)
	if err != nil {
		t.Fatalf("Failed to generate license: %v", err)
	}

	// Save the license to a file
	licenseFilePath := filepath.Join(tempDir, "license.lic")
	err = licgen.SaveLicenseToFile(licenseData, licenseFilePath)
	if err != nil {
		t.Fatalf("Failed to save license: %v", err)
	}

	// Verify the license file exists
	_, err = os.Stat(licenseFilePath)
	if err != nil {
		t.Fatalf("License file does not exist: %v", err)
	}

	// Verify the license file has the expected size
	fileInfo, err := os.Stat(licenseFilePath)
	if err != nil {
		t.Fatalf("Failed to get license file info: %v", err)
	}

	// License file should be larger than 256 bytes (minimum signature size)
	if fileInfo.Size() <= 256 {
		t.Fatalf("License file too small: %d bytes", fileInfo.Size())
	}
}
