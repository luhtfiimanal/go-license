package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/luhtfiimanal/go-license/pkg/licverify"
)

// Public key will be injected at build time
var publicKey string

func main() {
	// Parse command line flags
	licenseFile := flag.String("license", "license.lic", "Path to the license file")
	verbose := flag.Bool("verbose", false, "Show detailed hardware information")
	flag.Parse()

	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║   Go License Client Verification Example  ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()

	// Check if public key is provided
	if publicKey == "" {
		// For development, use a local public key file if available
		fmt.Println("🔑 Loading public key from file...")
		keyData, err := os.ReadFile("public.pem")
		if err != nil {
			log.Fatalf("❌ No public key available. In production, this should be injected at build time.")
		}
		publicKey = string(keyData)
		fmt.Println("✅ Public key loaded successfully")
	} else {
		fmt.Println("✅ Using injected public key")
	}

	// Create a license verifier
	fmt.Println("🔍 Creating license verifier...")
	verifier, err := licverify.NewVerifier(publicKey)
	if err != nil {
		log.Fatalf("❌ Failed to create license verifier: %v", err)
	}
	fmt.Println("✅ License verifier created successfully")

	// Load and verify the license
	fmt.Printf("📄 Loading license from %s...\n", *licenseFile)
	license, err := verifier.LoadLicense(*licenseFile)
	if err != nil {
		log.Fatalf("❌ Failed to load license: %v", err)
	}
	fmt.Println("✅ License loaded successfully")

	// Verify the license
	fmt.Println("🔐 Verifying license...")
	err = license.IsValid(verifier)
	if err != nil {
		log.Fatalf("❌ License validation failed: %v", err)
	}
	fmt.Println("✅ License is valid!")

	// Display license information
	fmt.Println("\n📋 License Information:")
	fmt.Printf("   License ID: %s\n", license.ID)
	fmt.Printf("   Customer: %s\n", license.CustomerID)
	fmt.Printf("   Product: %s\n", license.ProductID)
	fmt.Printf("   Issued: %s\n", license.IssueDate.Format(time.RFC3339))
	fmt.Printf("   Expires: %s\n", license.ExpiryDate.Format(time.RFC3339))

	// Calculate days remaining
	daysRemaining := int(time.Until(license.ExpiryDate).Hours() / 24)
	if daysRemaining > 0 {
		fmt.Printf("   Status: Active (%d days remaining)\n", daysRemaining)
	} else {
		fmt.Printf("   Status: Expired\n")
	}

	fmt.Printf("   Features: %v\n", license.Features)

	// Show hardware binding information if verbose
	if *verbose {
		fmt.Println("\n🖥️  Hardware Binding Information:")
		if len(license.HardwareIDs.MACAddresses) > 0 {
			fmt.Printf("   MAC Addresses: %v\n", license.HardwareIDs.MACAddresses)
		}
		if len(license.HardwareIDs.DiskIDs) > 0 {
			fmt.Printf("   Disk IDs: %v\n", license.HardwareIDs.DiskIDs)
		}
		if len(license.HardwareIDs.HostNames) > 0 {
			fmt.Printf("   Hostnames: %v\n", license.HardwareIDs.HostNames)
		}

		// Get current hardware info for debugging
		hwInfo, err := licverify.GetHardwareInfo()
		if err != nil {
			fmt.Printf("   ⚠️  Warning: Could not get hardware info: %v\n", err)
		} else {
			fmt.Println("\n🖥️  Current Hardware Information:")
			fmt.Printf("   Hostname: %s\n", hwInfo.Hostname)
			fmt.Printf("   MAC Addresses: %v\n", hwInfo.MACAddresses)
			fmt.Printf("   Disk IDs: %v\n", hwInfo.DiskIDs)
			fmt.Printf("   CPU Info: %s\n", hwInfo.CPUInfo)
		}
	}

	fmt.Println("\n✨ Application is ready to run...")
}
