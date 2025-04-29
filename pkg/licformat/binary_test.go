package licformat

import (
	"testing"
	"time"
)

func TestBinaryEncodingDecoding(t *testing.T) {
	// Create a test license data
	original := &LicenseData{
		ID:           "test-license-123",
		CustomerID:   "customer-456",
		ProductID:    "product-789",
		SerialNumber: "SN-ABCDEF",
		IssueDate:    time.Now().Truncate(time.Second), // Truncate to avoid nanosecond precision issues
		ExpiryDate:   time.Now().AddDate(1, 0, 0).Truncate(time.Second),
		Features:     []string{"feature1", "feature2", "feature3"},
		HardwareIDs: HardwareBindingData{
			MACAddresses: []string{"00:11:22:33:44:55", "AA:BB:CC:DD:EE:FF"},
			DiskIDs:      []string{"disk-id-1", "disk-id-2"},
			HostNames:    []string{"host1.example.com", "host2.example.com"},
			CustomIDs:    []string{"custom-id-1", "custom-id-2"},
		},
	}

	// Encode to binary
	binaryData, err := EncodeLicenseData(original)
	if err != nil {
		t.Fatalf("Failed to encode license data: %v", err)
	}

	// Decode from binary
	decoded, err := DecodeLicenseData(binaryData)
	if err != nil {
		t.Fatalf("Failed to decode license data: %v", err)
	}

	// Verify all fields match
	if original.ID != decoded.ID {
		t.Errorf("ID mismatch: expected %s, got %s", original.ID, decoded.ID)
	}
	if original.CustomerID != decoded.CustomerID {
		t.Errorf("CustomerID mismatch: expected %s, got %s", original.CustomerID, decoded.CustomerID)
	}
	if original.ProductID != decoded.ProductID {
		t.Errorf("ProductID mismatch: expected %s, got %s", original.ProductID, decoded.ProductID)
	}
	if original.SerialNumber != decoded.SerialNumber {
		t.Errorf("SerialNumber mismatch: expected %s, got %s", original.SerialNumber, decoded.SerialNumber)
	}
	if !original.IssueDate.Equal(decoded.IssueDate) {
		t.Errorf("IssueDate mismatch: expected %v, got %v", original.IssueDate, decoded.IssueDate)
	}
	if !original.ExpiryDate.Equal(decoded.ExpiryDate) {
		t.Errorf("ExpiryDate mismatch: expected %v, got %v", original.ExpiryDate, decoded.ExpiryDate)
	}

	// Check features
	if len(original.Features) != len(decoded.Features) {
		t.Errorf("Features length mismatch: expected %d, got %d", len(original.Features), len(decoded.Features))
	} else {
		for i, feature := range original.Features {
			if feature != decoded.Features[i] {
				t.Errorf("Feature mismatch at index %d: expected %s, got %s", i, feature, decoded.Features[i])
			}
		}
	}

	// Check hardware binding
	checkStringSlice(t, "MACAddresses", original.HardwareIDs.MACAddresses, decoded.HardwareIDs.MACAddresses)
	checkStringSlice(t, "DiskIDs", original.HardwareIDs.DiskIDs, decoded.HardwareIDs.DiskIDs)
	checkStringSlice(t, "HostNames", original.HardwareIDs.HostNames, decoded.HardwareIDs.HostNames)
	checkStringSlice(t, "CustomIDs", original.HardwareIDs.CustomIDs, decoded.HardwareIDs.CustomIDs)
}

func checkStringSlice(t *testing.T, name string, expected, actual []string) {
	if len(expected) != len(actual) {
		t.Errorf("%s length mismatch: expected %d, got %d", name, len(expected), len(actual))
		return
	}
	for i, item := range expected {
		if item != actual[i] {
			t.Errorf("%s mismatch at index %d: expected %s, got %s", name, i, item, actual[i])
		}
	}
}

func TestLicenseAdapter(t *testing.T) {
	// Create a test license
	original := &License{
		ID:           "test-license-123",
		CustomerID:   "customer-456",
		ProductID:    "product-789",
		SerialNumber: "SN-ABCDEF",
		IssueDate:    time.Now().Truncate(time.Second),
		ExpiryDate:   time.Now().AddDate(1, 0, 0).Truncate(time.Second),
		Features:     []string{"feature1", "feature2", "feature3"},
		HardwareIDs: HardwareBinding{
			MACAddresses: []string{"00:11:22:33:44:55", "AA:BB:CC:DD:EE:FF"},
			DiskIDs:      []string{"disk-id-1", "disk-id-2"},
			HostNames:    []string{"host1.example.com", "host2.example.com"},
			CustomIDs:    []string{"custom-id-1", "custom-id-2"},
		},
		Signature: []byte("test-signature"),
	}

	// Convert to LicenseData
	data := ToLicenseData(original)

	// Convert back to License
	converted := FromLicenseData(data)

	// Verify all fields match (except Signature which is not part of LicenseData)
	if original.ID != converted.ID {
		t.Errorf("ID mismatch: expected %s, got %s", original.ID, converted.ID)
	}
	if original.CustomerID != converted.CustomerID {
		t.Errorf("CustomerID mismatch: expected %s, got %s", original.CustomerID, converted.CustomerID)
	}
	if original.ProductID != converted.ProductID {
		t.Errorf("ProductID mismatch: expected %s, got %s", original.ProductID, converted.ProductID)
	}
	if original.SerialNumber != converted.SerialNumber {
		t.Errorf("SerialNumber mismatch: expected %s, got %s", original.SerialNumber, converted.SerialNumber)
	}
	if !original.IssueDate.Equal(converted.IssueDate) {
		t.Errorf("IssueDate mismatch: expected %v, got %v", original.IssueDate, converted.IssueDate)
	}
	if !original.ExpiryDate.Equal(converted.ExpiryDate) {
		t.Errorf("ExpiryDate mismatch: expected %v, got %v", original.ExpiryDate, converted.ExpiryDate)
	}

	// Check features
	checkStringSlice(t, "Features", original.Features, converted.Features)

	// Check hardware binding
	checkStringSlice(t, "MACAddresses", original.HardwareIDs.MACAddresses, converted.HardwareIDs.MACAddresses)
	checkStringSlice(t, "DiskIDs", original.HardwareIDs.DiskIDs, converted.HardwareIDs.DiskIDs)
	checkStringSlice(t, "HostNames", original.HardwareIDs.HostNames, converted.HardwareIDs.HostNames)
	checkStringSlice(t, "CustomIDs", original.HardwareIDs.CustomIDs, converted.HardwareIDs.CustomIDs)
}

func TestEncodeLicense(t *testing.T) {
	// Create a test license
	license := &License{
		ID:           "test-license-123",
		CustomerID:   "customer-456",
		ProductID:    "product-789",
		SerialNumber: "SN-ABCDEF",
		IssueDate:    time.Now().Truncate(time.Second),
		ExpiryDate:   time.Now().AddDate(1, 0, 0).Truncate(time.Second),
		Features:     []string{"feature1", "feature2", "feature3"},
		HardwareIDs: HardwareBinding{
			MACAddresses: []string{"00:11:22:33:44:55", "AA:BB:CC:DD:EE:FF"},
			DiskIDs:      []string{"disk-id-1", "disk-id-2"},
			HostNames:    []string{"host1.example.com", "host2.example.com"},
			CustomIDs:    []string{"custom-id-1", "custom-id-2"},
		},
	}

	// Encode to binary
	binaryData, err := EncodeLicense(license)
	if err != nil {
		t.Fatalf("Failed to encode license: %v", err)
	}

	// Decode from binary
	decoded, err := DecodeLicense(binaryData)
	if err != nil {
		t.Fatalf("Failed to decode license: %v", err)
	}

	// Verify all fields match
	if license.ID != decoded.ID {
		t.Errorf("ID mismatch: expected %s, got %s", license.ID, decoded.ID)
	}
	if license.CustomerID != decoded.CustomerID {
		t.Errorf("CustomerID mismatch: expected %s, got %s", license.CustomerID, decoded.CustomerID)
	}
	if license.ProductID != decoded.ProductID {
		t.Errorf("ProductID mismatch: expected %s, got %s", license.ProductID, decoded.ProductID)
	}
	if license.SerialNumber != decoded.SerialNumber {
		t.Errorf("SerialNumber mismatch: expected %s, got %s", license.SerialNumber, decoded.SerialNumber)
	}
	if !license.IssueDate.Equal(decoded.IssueDate) {
		t.Errorf("IssueDate mismatch: expected %v, got %v", license.IssueDate, decoded.IssueDate)
	}
	if !license.ExpiryDate.Equal(decoded.ExpiryDate) {
		t.Errorf("ExpiryDate mismatch: expected %v, got %v", license.ExpiryDate, decoded.ExpiryDate)
	}

	// Check features
	checkStringSlice(t, "Features", license.Features, decoded.Features)

	// Check hardware binding
	checkStringSlice(t, "MACAddresses", license.HardwareIDs.MACAddresses, decoded.HardwareIDs.MACAddresses)
	checkStringSlice(t, "DiskIDs", license.HardwareIDs.DiskIDs, decoded.HardwareIDs.DiskIDs)
	checkStringSlice(t, "HostNames", license.HardwareIDs.HostNames, decoded.HardwareIDs.HostNames)
	checkStringSlice(t, "CustomIDs", license.HardwareIDs.CustomIDs, decoded.HardwareIDs.CustomIDs)
}
