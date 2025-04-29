package licformat

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"
)

// BinaryFormat implements a binary serialization format for licenses
// This is an internal implementation that doesn't change the public API

// Format version to ensure compatibility
const currentVersion byte = 1

// Header for the binary format
type header struct {
	Version byte
	Length  uint32 // Length of the license data (excluding signature)
}

// LicenseData holds the data for a license in a format-agnostic way
type LicenseData struct {
	ID           string
	CustomerID   string
	ProductID    string
	SerialNumber string
	IssueDate    time.Time
	ExpiryDate   time.Time
	Features     []string
	HardwareIDs  HardwareBindingData
}

// HardwareBindingData contains hardware identifiers for license binding
type HardwareBindingData struct {
	MACAddresses []string
	DiskIDs      []string
	HostNames    []string
	CustomIDs    []string
}

// EncodeLicenseData converts license data to binary format
func EncodeLicenseData(data *LicenseData) ([]byte, error) {
	var buf bytes.Buffer

	// Write header placeholder (will update length later)
	h := header{Version: currentVersion, Length: 0}
	if err := binary.Write(&buf, binary.LittleEndian, h); err != nil {
		return nil, err
	}
	headerSize := buf.Len()

	// Write license fields
	writeString(&buf, data.ID)
	writeString(&buf, data.CustomerID)
	writeString(&buf, data.ProductID)
	writeString(&buf, data.SerialNumber)

	// Write timestamps
	binary.Write(&buf, binary.LittleEndian, data.IssueDate.Unix())
	binary.Write(&buf, binary.LittleEndian, data.ExpiryDate.Unix())

	// Write features
	writeStringSlice(&buf, data.Features)

	// Write hardware binding
	writeStringSlice(&buf, data.HardwareIDs.MACAddresses)
	writeStringSlice(&buf, data.HardwareIDs.DiskIDs)
	writeStringSlice(&buf, data.HardwareIDs.HostNames)
	writeStringSlice(&buf, data.HardwareIDs.CustomIDs)

	// Update header with correct length
	bytes := buf.Bytes()
	binary.LittleEndian.PutUint32(bytes[1:5], uint32(len(bytes)-headerSize))

	return bytes, nil
}

// DecodeLicenseData converts binary data back to license data
func DecodeLicenseData(data []byte) (*LicenseData, error) {
	if len(data) < 5 { // Minimum size for header
		return nil, errors.New("data too small to be a valid license")
	}

	buf := bytes.NewReader(data)

	// Read header
	var h header
	if err := binary.Read(buf, binary.LittleEndian, &h); err != nil {
		return nil, err
	}

	// Check version
	if h.Version != currentVersion {
		return nil, errors.New("unsupported license format version")
	}

	licenseData := &LicenseData{}

	// Read license fields
	var err error
	licenseData.ID, err = readString(buf)
	if err != nil {
		return nil, err
	}

	licenseData.CustomerID, err = readString(buf)
	if err != nil {
		return nil, err
	}

	licenseData.ProductID, err = readString(buf)
	if err != nil {
		return nil, err
	}

	licenseData.SerialNumber, err = readString(buf)
	if err != nil {
		return nil, err
	}

	// Read timestamps
	var issueUnix, expiryUnix int64
	if err := binary.Read(buf, binary.LittleEndian, &issueUnix); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &expiryUnix); err != nil {
		return nil, err
	}
	licenseData.IssueDate = time.Unix(issueUnix, 0)
	licenseData.ExpiryDate = time.Unix(expiryUnix, 0)

	// Read features
	licenseData.Features, err = readStringSlice(buf)
	if err != nil {
		return nil, err
	}

	// Read hardware binding
	licenseData.HardwareIDs.MACAddresses, err = readStringSlice(buf)
	if err != nil {
		return nil, err
	}

	licenseData.HardwareIDs.DiskIDs, err = readStringSlice(buf)
	if err != nil {
		return nil, err
	}

	licenseData.HardwareIDs.HostNames, err = readStringSlice(buf)
	if err != nil {
		return nil, err
	}

	licenseData.HardwareIDs.CustomIDs, err = readStringSlice(buf)
	if err != nil {
		return nil, err
	}

	return licenseData, nil
}

// Helper functions for reading/writing strings and slices

func writeString(buf *bytes.Buffer, s string) {
	data := []byte(s)
	binary.Write(buf, binary.LittleEndian, uint16(len(data)))
	buf.Write(data)
}

func readString(buf *bytes.Reader) (string, error) {
	var length uint16
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return "", err
	}

	data := make([]byte, length)
	if _, err := buf.Read(data); err != nil {
		return "", err
	}

	return string(data), nil
}

func writeStringSlice(buf *bytes.Buffer, slice []string) {
	binary.Write(buf, binary.LittleEndian, uint16(len(slice)))
	for _, s := range slice {
		writeString(buf, s)
	}
}

func readStringSlice(buf *bytes.Reader) ([]string, error) {
	var length uint16
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return nil, err
	}

	result := make([]string, length)
	for i := 0; i < int(length); i++ {
		s, err := readString(buf)
		if err != nil {
			return nil, err
		}
		result[i] = s
	}

	return result, nil
}
