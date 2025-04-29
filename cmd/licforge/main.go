package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/luhtfiimanal/go-license/v2/pkg/licformat"
	"github.com/luhtfiimanal/go-license/v2/pkg/licgen"
	"github.com/luhtfiimanal/go-license/v2/pkg/licverify"
)

const version = "2.0.0"

func main() {
	// Define command-line flags
	keygenCmd := flag.NewFlagSet("keygen", flag.ExitOnError)
	keygenKeyDir := keygenCmd.String("dir", "keys", "Directory to store keys")
	keygenKeySize := keygenCmd.Int("size", 2048, "RSA key size (2048, 3072, or 4096)")
	keygenForce := keygenCmd.Bool("force", false, "Force overwrite of existing keys")

	genlicenseCmd := flag.NewFlagSet("genlicense", flag.ExitOnError)
	genlicenseID := genlicenseCmd.String("id", "", "License ID")
	genlicenseCustomerID := genlicenseCmd.String("customer", "", "Customer ID")
	genlicenseProductID := genlicenseCmd.String("product", "", "Product ID")
	genlicenseSerialNumber := genlicenseCmd.String("serial", "", "Serial number")
	genlicenseValidDays := genlicenseCmd.Int("days", 365, "License validity in days")
	genlicenseFeatures := genlicenseCmd.String("features", "basic", "Comma-separated list of features")
	genlicenseMACAddresses := genlicenseCmd.String("macs", "", "Comma-separated list of MAC addresses")
	genlicenseDiskIDs := genlicenseCmd.String("diskids", "", "Comma-separated list of disk IDs")
	genlicenseHostnames := genlicenseCmd.String("hostnames", "", "Comma-separated list of hostnames")
	genlicensePrivateKey := genlicenseCmd.String("key", "keys/private.pem", "Path to private key")
	genlicenseOutput := genlicenseCmd.String("output", "license.lic", "Output license file")
	// Format flag removed in v2.0.0 - binary format is now the only option
	genlicenseAutoHardware := genlicenseCmd.Bool("auto-hardware", false, "Automatically use current hardware information")
	genlicenseInteractive := genlicenseCmd.Bool("interactive", false, "Interactive mode")

	infoCmd := flag.NewFlagSet("info", flag.ExitOnError)
	infoLicenseFile := infoCmd.String("license", "license.lic", "License file")
	infoPublicKey := infoCmd.String("key", "keys/public.pem", "Path to public key")

	// Print banner
	printBanner()

	// Parse command-line arguments
	if len(os.Args) < 2 {
		fmt.Println("Error: No command specified")
		printUsage()
		os.Exit(1)
	}

	// Handle commands
	switch os.Args[1] {
	case "keygen":
		keygenCmd.Parse(os.Args[2:])
		generateAndSaveKeyPair(*keygenKeyDir, *keygenKeySize, *keygenForce)

	case "genlicense":
		genlicenseCmd.Parse(os.Args[2:])
		if *genlicenseInteractive {
			runInteractiveGeneration(*genlicensePrivateKey, *genlicenseOutput)
		} else {
			if *genlicenseID == "" || *genlicenseCustomerID == "" || *genlicenseProductID == "" || *genlicenseSerialNumber == "" {
				fmt.Println("❌ Error: License ID, Customer ID, Product ID, and Serial Number are required")
				fmt.Println("\nCommand options:")
				genlicenseCmd.PrintDefaults()
				os.Exit(1)
			}

			generateAndSaveLicense(
				*genlicenseID,
				*genlicenseCustomerID,
				*genlicenseProductID,
				*genlicenseSerialNumber,
				*genlicenseValidDays,
				*genlicenseFeatures,
				*genlicenseMACAddresses,
				*genlicenseDiskIDs,
				*genlicenseHostnames,
				*genlicensePrivateKey,
				*genlicenseOutput,
				*genlicenseAutoHardware,
			)
		}

	case "info":
		infoCmd.Parse(os.Args[2:])
		displayLicenseInfo(*infoLicenseFile, *infoPublicKey)

	case "version":
		fmt.Printf("licforge version %s\n", version)

	case "help":
		printUsage()

	default:
		fmt.Printf("❌ Error: Unknown command '%s'\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

// printBanner prints the application banner
func printBanner() {
	fmt.Println("╔═════════════════════════════════════════════════════════════╗")
	fmt.Println("║                                                             ║")
	fmt.Println("║                   GO LICENSE FORGE                          ║")
	fmt.Println("║                                                             ║")
	fmt.Println("║         Offline License Generation & Management             ║")
	fmt.Println("║                                                             ║")
	fmt.Println("╚═════════════════════════════════════════════════════════════╝")
	fmt.Printf("Version %s\n\n", version)
}

// printUsage prints the usage information
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  licforge [command] [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  keygen      Generate a new RSA key pair")
	fmt.Println("  genlicense  Generate a license")
	fmt.Println("  info        Display license information")
	fmt.Println("  version     Display version information")
	fmt.Println("  help        Display this help message")
	fmt.Println("\nRun 'licforge [command] -h' for more information on a command.")
}

// generateAndSaveKeyPair generates a new RSA key pair and saves it to files
func generateAndSaveKeyPair(keyDir string, keySize int, force bool) {
	fmt.Println("🔑 Generating RSA key pair...")

	// Check if key files already exist
	privateKeyPath := filepath.Join(keyDir, "private.pem")
	publicKeyPath := filepath.Join(keyDir, "public.pem")

	if !force {
		if _, err := os.Stat(privateKeyPath); err == nil {
			fmt.Println("❌ Private key already exists. Use -force to overwrite.")
			os.Exit(1)
		}
		if _, err := os.Stat(publicKeyPath); err == nil {
			fmt.Println("❌ Public key already exists. Use -force to overwrite.")
			os.Exit(1)
		}
	}

	// Validate key size
	validSizes := map[int]bool{2048: true, 3072: true, 4096: true}
	if !validSizes[keySize] {
		fmt.Println("❌ Invalid key size. Must be one of: 2048, 3072, 4096")
		os.Exit(1)
	}

	fmt.Printf("📊 Key size: %d bits\n", keySize)

	privateKeyPEM, publicKeyPEM, err := licgen.GenerateKeyPair(keySize)
	if err != nil {
		fmt.Printf("❌ Failed to generate key pair: %v\n", err)
		os.Exit(1)
	}

	// Create key directory if it doesn't exist
	if err := os.MkdirAll(keyDir, 0755); err != nil {
		fmt.Printf("❌ Failed to create key directory: %v\n", err)
		os.Exit(1)
	}

	// Save private key
	if err := os.WriteFile(privateKeyPath, []byte(privateKeyPEM), 0600); err != nil {
		fmt.Printf("❌ Failed to save private key: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Private key saved to: %s\n", privateKeyPath)

	// Save public key
	if err := os.WriteFile(publicKeyPath, []byte(publicKeyPEM), 0644); err != nil {
		fmt.Printf("❌ Failed to save public key: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Public key saved to: %s\n", publicKeyPath)

	fmt.Println("\n🔐 Key pair generated successfully!")
}

// generateAndSaveLicense generates a license and saves it to a file
func generateAndSaveLicense(
	licenseID string,
	customerID string,
	productID string,
	serialNumber string,
	validDays int,
	featuresStr string,
	macAddressesStr string,
	diskIDsStr string,
	hostnamesStr string,
	privateKeyPath string,
	outputPath string,
	autoHardware bool,
) {
	fmt.Println("📜 Generating license...")

	// Read private key
	privateKeyPEM, err := os.ReadFile(privateKeyPath)
	if err != nil {
		fmt.Printf("❌ Failed to read private key: %v\n", err)
		os.Exit(1)
	}

	// Parse private key
	privateKey, err := licgen.ParsePrivateKey(string(privateKeyPEM))
	if err != nil {
		fmt.Printf("❌ Failed to parse private key: %v\n", err)
		os.Exit(1)
	}

	// Parse features
	features := parseCommaSeparatedList(featuresStr)

	// Parse hardware binding
	var hardwareIDs licverify.HardwareBinding

	if autoHardware {
		// Get current hardware information
		fmt.Println("💻 Detecting current hardware information...")
		hwInfo, err := licverify.GetHardwareInfo()
		if err != nil {
			fmt.Printf("❌ Failed to get hardware information: %v\n", err)
			os.Exit(1)
		}

		// Use current hardware information
		hardwareIDs = licverify.HardwareBinding{
			MACAddresses: hwInfo.MACAddresses,
			DiskIDs:      hwInfo.DiskIDs,
			HostNames:    []string{hwInfo.Hostname},
		}

		fmt.Println("✅ Hardware information detected:")
		if len(hwInfo.MACAddresses) > 0 {
			fmt.Printf("   MAC Addresses: %v\n", hwInfo.MACAddresses)
		}
		if len(hwInfo.DiskIDs) > 0 {
			fmt.Printf("   Disk IDs: %v\n", hwInfo.DiskIDs)
		}
		if hwInfo.Hostname != "" {
			fmt.Printf("   Hostname: %s\n", hwInfo.Hostname)
		}
	} else {
		// Use provided hardware information
		hardwareIDs = licverify.HardwareBinding{
			MACAddresses: parseCommaSeparatedList(macAddressesStr),
			DiskIDs:      parseCommaSeparatedList(diskIDsStr),
			HostNames:    parseCommaSeparatedList(hostnamesStr),
		}
	}

	// Generate license
	fmt.Println("🔐 Signing license with private key...")
	// In v2.0.0, binary format is the only option
	fmt.Println("📦 Using binary format")

	licenseData, err := licgen.GenerateLicense(
		licenseID,
		customerID,
		productID,
		serialNumber,
		time.Duration(validDays)*24*time.Hour,
		features,
		hardwareIDs,
		privateKey,
	)
	if err != nil {
		fmt.Printf("❌ Failed to generate license: %v\n", err)
		os.Exit(1)
	}

	// Save license
	if err := licgen.SaveLicenseToFile(licenseData, outputPath); err != nil {
		fmt.Printf("❌ Failed to save license: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ License saved to: %s\n", outputPath)

	// Print license information
	var license licverify.License
	// Try to decode as binary first, fall back to JSON if that fails
	licenseBytes := licenseData[:len(licenseData)-256] // Remove signature
	// In v2.0.0, binary format is the only option for new licenses
	// JSON format is only supported for reading legacy licenses
	importedLicense, err := licformat.DecodeLicense(licenseBytes)
	if err != nil {
		// Try JSON format as fallback
		if err := json.Unmarshal(licenseBytes, &license); err != nil {
			fmt.Printf("❌ Failed to parse license data: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Convert the imported license to licverify.License format
		license = licverify.License{
			ID:           importedLicense.ID,
			CustomerID:   importedLicense.CustomerID,
			ProductID:    importedLicense.ProductID,
			SerialNumber: importedLicense.SerialNumber,
			IssueDate:    importedLicense.IssueDate,
			ExpiryDate:   importedLicense.ExpiryDate,
			Features:     importedLicense.Features,
			HardwareIDs: licverify.HardwareBinding{
				MACAddresses: importedLicense.HardwareIDs.MACAddresses,
				DiskIDs:      importedLicense.HardwareIDs.DiskIDs,
				HostNames:    importedLicense.HardwareIDs.HostNames,
				CustomIDs:    importedLicense.HardwareIDs.CustomIDs,
			},
		}
	}

	fmt.Println("\n📃 License Information:")
	fmt.Printf("   License ID: %s\n", license.ID)
	fmt.Printf("   Customer ID: %s\n", license.CustomerID)
	fmt.Printf("   Product ID: %s\n", license.ProductID)
	fmt.Printf("   Serial Number: %s\n", license.SerialNumber)
	fmt.Printf("   Issue Date: %s\n", license.IssueDate.Format(time.RFC3339))
	fmt.Printf("   Expiry Date: %s\n", license.ExpiryDate.Format(time.RFC3339))

	// Calculate days until expiry
	daysUntilExpiry := int(time.Until(license.ExpiryDate).Hours() / 24)
	if daysUntilExpiry > 0 {
		fmt.Printf("   Validity: %d days\n", daysUntilExpiry)
	} else {
		fmt.Printf("   Validity: Expired\n")
	}

	fmt.Printf("   Features: %v\n", license.Features)

	if len(license.HardwareIDs.MACAddresses) > 0 {
		fmt.Printf("   MAC Addresses: %v\n", license.HardwareIDs.MACAddresses)
	}
	if len(license.HardwareIDs.DiskIDs) > 0 {
		fmt.Printf("   Disk IDs: %v\n", license.HardwareIDs.DiskIDs)
	}
	if len(license.HardwareIDs.HostNames) > 0 {
		fmt.Printf("   Hostnames: %v\n", license.HardwareIDs.HostNames)
	}

	fmt.Println("\n✨ License generated successfully!")
}

// runInteractiveGeneration generates a license interactively
func runInteractiveGeneration(privateKeyPath, outputPath string) {
	fmt.Println("💬 Interactive License Generation")

	// In v2.0.0, binary is the only format
	fmt.Println("Note: In v2.0.0, licenses are generated in binary format only")

	// Ask about auto-hardware
	autoHardwarePrompt := "Use current hardware information (y/n)"
	autoHardwareStr := promptForInput(autoHardwarePrompt, "y")
	autoHardware := strings.ToLower(autoHardwareStr) == "y" || strings.ToLower(autoHardwareStr) == "yes"

	// Read license information from user
	licenseID := promptForInput("License ID")
	customerID := promptForInput("Customer ID")
	productID := promptForInput("Product ID")
	serialNumber := promptForInput("Serial Number")

	// Read validity period
	validDaysStr := promptForInput("Validity (days)", "365")
	validDays, err := strconv.Atoi(validDaysStr)
	if err != nil || validDays <= 0 {
		fmt.Println("❌ Invalid validity period. Using default (365 days).")
		validDays = 365
	}

	// Read features
	featuresStr := promptForInput("Features (comma-separated)", "basic")

	// Read hardware binding
	fmt.Println("\n💻 Hardware Binding (leave empty if not required):")
	macAddressesStr := promptForInput("MAC Addresses (comma-separated)")
	diskIDsStr := promptForInput("Disk IDs (comma-separated)")
	hostnamesStr := promptForInput("Hostnames (comma-separated)")

	// Generate and save the license
	generateAndSaveLicense(
		licenseID,
		customerID,
		productID,
		serialNumber,
		validDays,
		featuresStr,
		macAddressesStr,
		diskIDsStr,
		hostnamesStr,
		privateKeyPath,
		outputPath,
		autoHardware,
	)
}

// displayLicenseInfo displays information about a license
func displayLicenseInfo(licenseFile, publicKeyFile string) {
	fmt.Printf("🔍 Examining license file: %s\n", licenseFile)

	// Read public key
	publicKeyPEM, err := os.ReadFile(publicKeyFile)
	if err != nil {
		fmt.Printf("❌ Failed to read public key: %v\n", err)
		os.Exit(1)
	}

	// Create verifier
	verifier, err := licverify.NewVerifier(string(publicKeyPEM))
	if err != nil {
		fmt.Printf("❌ Failed to create verifier: %v\n", err)
		os.Exit(1)
	}

	// Load license
	license, err := verifier.LoadLicense(licenseFile)
	if err != nil {
		fmt.Printf("❌ Failed to load license: %v\n", err)
		os.Exit(1)
	}

	// Detect license format (binary or legacy JSON)
	var formatType string
	// Check if we can decode it as binary
	licenseData, err := os.ReadFile(licenseFile)
	if err == nil && len(licenseData) > 256 {
		_, err := licformat.DecodeLicense(licenseData[:len(licenseData)-256])
		if err == nil {
			formatType = "binary"
		} else {
			formatType = "json (legacy)"
		}
	}
	fmt.Printf("📦 License format: %s\n", formatType)

	// Verify license
	fmt.Println("🔐 Verifying license signature...")
	err = verifier.VerifySignature(license)
	if err != nil {
		fmt.Printf("❌ License signature verification failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ License signature is valid")

	// Check expiry
	fmt.Println("📅 Checking license expiry...")
	err = verifier.VerifyExpiry(license)
	if err != nil {
		fmt.Printf("⚠️ License expiry check failed: %v\n", err)
	} else {
		fmt.Println("✅ License is not expired")
	}

	// Check hardware binding
	if len(license.HardwareIDs.MACAddresses) > 0 ||
		len(license.HardwareIDs.DiskIDs) > 0 ||
		len(license.HardwareIDs.HostNames) > 0 {
		fmt.Println("💻 Checking hardware binding...")
		err = verifier.VerifyHardwareBinding(license)
		if err != nil {
			fmt.Printf("⚠️ Hardware binding check failed: %v\n", err)
		} else {
			fmt.Println("✅ Hardware binding is valid for this machine")
		}
	} else {
		fmt.Println("💻 No hardware binding in this license")
	}

	// Print license information
	fmt.Println("\n📃 License Information:")
	fmt.Printf("   License ID: %s\n", license.ID)
	fmt.Printf("   Customer ID: %s\n", license.CustomerID)
	fmt.Printf("   Product ID: %s\n", license.ProductID)
	fmt.Printf("   Serial Number: %s\n", license.SerialNumber)
	fmt.Printf("   Issue Date: %s\n", license.IssueDate.Format(time.RFC3339))
	fmt.Printf("   Expiry Date: %s\n", license.ExpiryDate.Format(time.RFC3339))

	// Calculate days remaining
	daysRemaining := int(time.Until(license.ExpiryDate).Hours() / 24)
	if daysRemaining > 0 {
		fmt.Printf("   Status: Active (%d days remaining)\n", daysRemaining)
	} else {
		fmt.Printf("   Status: Expired (%d days ago)\n", -daysRemaining)
	}

	fmt.Printf("   Features: %v\n", license.Features)

	if len(license.HardwareIDs.MACAddresses) > 0 {
		fmt.Printf("   MAC Addresses: %v\n", license.HardwareIDs.MACAddresses)
	}
	if len(license.HardwareIDs.DiskIDs) > 0 {
		fmt.Printf("   Disk IDs: %v\n", license.HardwareIDs.DiskIDs)
	}
	if len(license.HardwareIDs.HostNames) > 0 {
		fmt.Printf("   Hostnames: %v\n", license.HardwareIDs.HostNames)
	}
}

// promptForInput prompts the user for input with an optional default value
func promptForInput(prompt string, defaultValue ...string) string {
	defaultVal := ""
	if len(defaultValue) > 0 {
		defaultVal = defaultValue[0]
		fmt.Printf("%s [%s]: ", prompt, defaultVal)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("❌ Error reading input: %v\n", err)
		os.Exit(1)
	}

	input = strings.TrimSpace(input)
	if input == "" && defaultVal != "" {
		return defaultVal
	}

	return input
}

// parseCommaSeparatedList parses a comma-separated list into a slice
func parseCommaSeparatedList(list string) []string {
	if list == "" {
		return nil
	}

	var result []string
	for _, item := range strings.Split(list, ",") {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
