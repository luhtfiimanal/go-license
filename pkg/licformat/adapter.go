package licformat

import (
	"time"
)

// LicenseAdapter provides functions to convert between license types
// without creating import cycles between packages

// License represents the structure of a license from the licverify package
// This is a copy of the struct to avoid import cycles
type License struct {
	ID           string
	CustomerID   string
	ProductID    string
	SerialNumber string
	IssueDate    time.Time
	ExpiryDate   time.Time
	Features     []string
	HardwareIDs  HardwareBinding
	Signature    []byte
}

// HardwareBinding is a copy of the struct from licverify
type HardwareBinding struct {
	MACAddresses []string
	DiskIDs      []string
	HostNames    []string
	CustomIDs    []string
}

// ToLicenseData converts a License to LicenseData
func ToLicenseData(license *License) *LicenseData {
	return &LicenseData{
		ID:           license.ID,
		CustomerID:   license.CustomerID,
		ProductID:    license.ProductID,
		SerialNumber: license.SerialNumber,
		IssueDate:    license.IssueDate,
		ExpiryDate:   license.ExpiryDate,
		Features:     license.Features,
		HardwareIDs: HardwareBindingData{
			MACAddresses: license.HardwareIDs.MACAddresses,
			DiskIDs:      license.HardwareIDs.DiskIDs,
			HostNames:    license.HardwareIDs.HostNames,
			CustomIDs:    license.HardwareIDs.CustomIDs,
		},
	}
}

// FromLicenseData converts LicenseData to a License
func FromLicenseData(data *LicenseData) *License {
	return &License{
		ID:           data.ID,
		CustomerID:   data.CustomerID,
		ProductID:    data.ProductID,
		SerialNumber: data.SerialNumber,
		IssueDate:    data.IssueDate,
		ExpiryDate:   data.ExpiryDate,
		Features:     data.Features,
		HardwareIDs: HardwareBinding{
			MACAddresses: data.HardwareIDs.MACAddresses,
			DiskIDs:      data.HardwareIDs.DiskIDs,
			HostNames:    data.HardwareIDs.HostNames,
			CustomIDs:    data.HardwareIDs.CustomIDs,
		},
	}
}

// EncodeLicense converts a License to binary format
func EncodeLicense(license *License) ([]byte, error) {
	data := ToLicenseData(license)
	return EncodeLicenseData(data)
}

// DecodeLicense converts binary data to a License
func DecodeLicense(data []byte) (*License, error) {
	licenseData, err := DecodeLicenseData(data)
	if err != nil {
		return nil, err
	}
	return FromLicenseData(licenseData), nil
}
