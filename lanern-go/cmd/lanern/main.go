// Package main provides the entry point for the Lantern EQ extractor.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/config"
	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/eq"
	"github.com/tmyhres/LanternGoExtract/lanern-go/pkg/infrastructure/logger"
)

const (
	exportDir    = "Exports/"
	settingsFile = "settings.txt"
	logFile      = "log.txt"
)

func main() {
	// Define command line flags
	var (
		archiveName string
		showHelp    bool
	)

	flag.StringVar(&archiveName, "archive", "", "Archive name to extract (filename/shortname/all/zones/characters/equipment/sounds)")
	flag.BoolVar(&showHelp, "help", false, "Show help message")
	flag.Parse()

	// If no flags provided, check positional arguments
	if archiveName == "" && flag.NArg() > 0 {
		archiveName = flag.Arg(0)
	}

	if showHelp || archiveName == "" {
		printUsage()
		os.Exit(0)
	}

	// Initialize logger
	log, err := logger.NewFileLogger(logFile, logger.VerbosityInfo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	// Initialize settings
	settings := config.NewSettings(settingsFile, log)
	if err := settings.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not load settings file: %v\n", err)
	}

	// Set logger verbosity from settings
	log.SetVerbosity(logger.Verbosity(settings.LoggerVerbosity))

	start := time.Now()

	// Get list of valid EQ files to process
	eqFiles := eq.GetValidEqFilePaths(settings.EverQuestDirectory, archiveName)
	sort.Strings(eqFiles)

	if len(eqFiles) == 0 && !eq.IsSpecialCaseExtraction(archiveName) {
		fmt.Printf("No valid EQ files found for: '%s' at path: %s\n", archiveName, settings.EverQuestDirectory)
		os.Exit(1)
	}

	// Process each file
	for _, file := range eqFiles {
		if err := extractArchive(file, exportDir, log, settings); err != nil {
			log.LogError(fmt.Sprintf("Failed to extract %s: %v", file, err))
			fmt.Printf("Failed to extract %s: %v\n", file, err)
		}
	}

	// Copy client data and music if applicable
	if err := copyClientData(archiveName, exportDir, log, settings); err != nil {
		log.LogWarning(fmt.Sprintf("Failed to copy client data: %v", err))
	}

	if err := copyMusic(archiveName, log, settings); err != nil {
		log.LogWarning(fmt.Sprintf("Failed to copy music: %v", err))
	}

	elapsed := time.Since(start)
	fmt.Printf("Extraction complete (%.2fs)\n", elapsed.Seconds())
}

// printUsage prints the usage information.
func printUsage() {
	fmt.Println("Lantern EQ Extractor")
	fmt.Println("")
	fmt.Println("Usage: lanern <archive>")
	fmt.Println("       lanern -archive=<archive>")
	fmt.Println("")
	fmt.Println("Archive options:")
	fmt.Println("  <filename>   - Extract a specific archive file (e.g., gfaydark.s3d)")
	fmt.Println("  <shortname>  - Extract zone by shortname (e.g., gfaydark)")
	fmt.Println("  all          - Extract all valid archives")
	fmt.Println("  zones        - Extract all zone archives")
	fmt.Println("  characters   - Extract all character archives")
	fmt.Println("  equipment    - Extract all equipment archives")
	fmt.Println("  sounds       - Extract all sound archives")
	fmt.Println("")
	fmt.Println("Flags:")
	flag.PrintDefaults()
}

// extractArchive extracts a single archive file.
// TODO: Implement actual extraction using the archive package.
func extractArchive(filePath, exportPath string, log logger.Logger, settings *config.Settings) error {
	fmt.Printf("Started extracting %s\n", filePath)
	log.LogInfo(fmt.Sprintf("Extracting archive: %s", filePath))

	// TODO: Call archive.Extract(filePath, exportPath, log, settings)
	// This is a placeholder that will be filled in when the ArchiveExtractor
	// is converted to Go.

	fmt.Printf("Finished extracting %s\n", filePath)
	return nil
}

// copyClientData copies client data files to the export directory.
// TODO: Implement actual copy logic.
func copyClientData(archiveName, exportPath string, log logger.Logger, settings *config.Settings) error {
	// Only copy for "all" or "clientdata" extractions
	if archiveName != "all" && archiveName != "clientdata" {
		return nil
	}

	log.LogInfo("Copying client data files")
	// TODO: Implement ClientDataCopier.Copy equivalent
	return nil
}

// copyMusic copies music files to the export directory.
// TODO: Implement actual copy logic.
func copyMusic(archiveName string, log logger.Logger, settings *config.Settings) error {
	if !settings.CopyMusic {
		return nil
	}

	// Only copy for "all" or "music" extractions
	if archiveName != "all" && archiveName != "music" {
		return nil
	}

	log.LogInfo("Copying music files")
	// TODO: Implement MusicCopier.Copy equivalent
	return nil
}
