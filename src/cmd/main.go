package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"plexrenamer/internal/cli"
	"plexrenamer/internal/database"
	"plexrenamer/internal/renamer"
)

// Config holds the application configuration
type Config struct {
	DatabasePath string
	OutputDir    string
	DryRun       bool
	ScriptMode   bool
	ScriptShell  string // "cmd", "powershell", or "bash"
	Mode         renamer.OperationMode
	TVFormat     string
	MovieFormat  string
	PathMapSrc   string
	PathMapDst   string
	AutoApprove  bool
}

func main() {
	config := parseFlags()

	if config.DatabasePath == "" {
		fmt.Fprintln(os.Stderr, "Error: database path is required")
		flag.Usage()
		os.Exit(1)
	}

	if err := run(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() *Config {
	config := &Config{}

	flag.StringVar(&config.OutputDir, "output", "", "Output directory for renamed files (required)")
	flag.BoolVar(&config.DryRun, "dry-run", false, "Preview changes without applying them")
	flag.BoolVar(&config.ScriptMode, "script", false, "Output shell commands instead of executing")
	flag.StringVar(&config.ScriptShell, "shell", "cmd", "Shell format for script output: cmd, powershell, or bash")
	modeStr := flag.String("mode", "move", "Operation mode: copy or move")
	flag.StringVar(&config.TVFormat, "tv-format", renamer.DefaultTVFormat, "Format for TV show filenames")
	flag.StringVar(&config.MovieFormat, "movie-format", renamer.DefaultMovieFormat, "Format for movie filenames")
	pathMap := flag.String("path-map", "", "Path mapping (old:new) for network shares")
	flag.BoolVar(&config.AutoApprove, "auto-approve", false, "Automatically approve all operations")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <database-path>\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "A CLI tool to rename/move media files based on Plex metadata.")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nExamples:")
		fmt.Fprintln(os.Stderr, "  plexrenamer --dry-run --output ./renamed ./plex.db")
		fmt.Fprintln(os.Stderr, "  plexrenamer --mode copy --output /media/organized ./plex.db")
		fmt.Fprintln(os.Stderr, "  plexrenamer --path-map 'F:\\Media:H:\\Media' --output ./out ./plex.db")
		fmt.Fprintln(os.Stderr, "  plexrenamer --script --shell powershell --output ./out ./plex.db > rename.ps1")
	}

	flag.Parse()

	if flag.NArg() > 0 {
		config.DatabasePath = flag.Arg(0)
	}

	// Parse mode
	switch strings.ToLower(*modeStr) {
	case "copy":
		config.Mode = renamer.ModeCopy
	case "move":
		config.Mode = renamer.ModeMove
	default:
		fmt.Fprintf(os.Stderr, "Invalid mode: %s (use 'copy' or 'move')\n", *modeStr)
		os.Exit(1)
	}

	// Parse path mapping
	if *pathMap != "" {
		parts := strings.SplitN(*pathMap, ":", 2)
		if len(parts) == 2 {
			config.PathMapSrc = parts[0]
			config.PathMapDst = parts[1]
		} else {
			fmt.Fprintln(os.Stderr, "Invalid path-map format. Use: old:new")
			os.Exit(1)
		}
	}

	return config
}

func run(config *Config) error {
	// In script mode, don't print banner to stdout (it would pollute the script)
	if !config.ScriptMode {
		cli.PrintBanner()

		if config.DryRun {
			pterm.Warning.Println("DRY RUN MODE - No files will be modified")
			fmt.Println()
		}
	}

	// Open database
	if !config.ScriptMode {
		pterm.Info.Printf("Opening database: %s\n", config.DatabasePath)
	}
	db, err := database.Open(config.DatabasePath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Get library sections
	sections, err := db.GetLibrarySections()
	if err != nil {
		return fmt.Errorf("failed to get library sections: %w", err)
	}

	if len(sections) == 0 {
		if !config.ScriptMode {
			pterm.Warning.Println("No library sections found in database.")
		}
		return nil
	}

	if !config.ScriptMode {
		pterm.Success.Printf("Found %d library section(s)\n", len(sections))
	}

	// Initialize formatter and prompter
	formatter := renamer.NewFormatter(config.TVFormat, config.MovieFormat)
	prompter := cli.NewPrompter()

	var allOperations []renamer.Operation

	// Process each library
	for _, section := range sections {
		content, err := db.GetLibraryContent(section)
		if err != nil {
			if !config.ScriptMode {
				pterm.Warning.Printf("Failed to get content for library %s: %v\n", section.Name, err)
			}
			continue
		}

		var selectedLocations []database.SectionLocation
		var locationOutputs []cli.LocationWithOutput

		// Skip prompts in script mode, or if auto-approve is set
		if !config.AutoApprove && !config.ScriptMode {
			proceed, locations, err := prompter.PromptLibrary(section, content.Locations)
			if err != nil {
				return err
			}
			if !proceed {
				continue
			}
			selectedLocations = locations

			// If locations were selected, prompt for output paths
			if selectedLocations != nil && len(selectedLocations) > 0 {
				locationOutputs, err = prompter.PromptLocationOutputs(selectedLocations, config.OutputDir)
				if err != nil {
					return err
				}
			}
		} else if !config.ScriptMode {
			fmt.Println()
			cli.PrintHeader(section.Name)
		}

		// Generate operations for this library
		ops, err := generateOperations(config, formatter, prompter, content, selectedLocations, locationOutputs)
		if err != nil {
			return err
		}
		allOperations = append(allOperations, ops...)
	}

	if len(allOperations) == 0 {
		if !config.ScriptMode {
			fmt.Println()
			pterm.Info.Println("No operations to perform.")
		}
		return nil
	}

	// Script mode: output commands and exit
	if config.ScriptMode {
		outputScript(allOperations, config.ScriptShell, config.Mode)
		return nil
	}

	// Show preview
	cli.ShowOperationPreview(allOperations, 10)

	// Confirm and execute
	proceed, err := prompter.ConfirmProceed(len(allOperations), config.Mode, config.DryRun)
	if err != nil {
		return err
	}
	if !proceed {
		pterm.Info.Println("Operation cancelled.")
		return nil
	}

	// Execute operations with progress bar
	fmt.Println()
	progressBar, _ := cli.CreateProgressBar(len(allOperations), "Processing files")

	results := make([]renamer.Result, len(allOperations))
	for i, op := range allOperations {
		results[i] = op.Execute(config.DryRun)
		if progressBar != nil {
			progressBar.Increment()
		}
	}

	if progressBar != nil {
		progressBar.Stop()
	}

	// Show results
	cli.ShowResults(results)

	return nil
}

// outputScript outputs shell commands for the operations
func outputScript(operations []renamer.Operation, shell string, mode renamer.OperationMode) {
	switch strings.ToLower(shell) {
	case "powershell", "ps", "ps1":
		outputPowerShell(operations, mode)
	case "bash", "sh":
		outputBash(operations, mode)
	default: // cmd
		outputCmd(operations, mode)
	}
}

func outputCmd(operations []renamer.Operation, mode renamer.OperationMode) {
	fmt.Println("@echo off")
	fmt.Println("REM Generated by Plex File Renamer")
	fmt.Println()

	for _, op := range operations {
		// Create destination directory
		destDir := filepath.Dir(op.Destination)
		fmt.Printf("if not exist \"%s\" mkdir \"%s\"\n", destDir, destDir)

		// Copy or move command
		if mode == renamer.ModeCopy {
			fmt.Printf("if not exist \"%s\" copy \"%s\" \"%s\"\n", op.Destination, op.Source, op.Destination)
		} else {
			fmt.Printf("if not exist \"%s\" move \"%s\" \"%s\"\n", op.Destination, op.Source, op.Destination)
		}
	}

	fmt.Println()
	fmt.Println("echo Done!")
	fmt.Println("pause")
}

func outputPowerShell(operations []renamer.Operation, mode renamer.OperationMode) {
	fmt.Println("# Generated by Plex File Renamer")
	fmt.Println()

	for _, op := range operations {
		// Escape for PowerShell
		src := strings.ReplaceAll(op.Source, "'", "''")
		dst := strings.ReplaceAll(op.Destination, "'", "''")
		destDir := strings.ReplaceAll(filepath.Dir(op.Destination), "'", "''")

		// Create destination directory
		fmt.Printf("if (-not (Test-Path '%s')) { New-Item -ItemType Directory -Path '%s' -Force | Out-Null }\n", destDir, destDir)

		// Copy or move command
		if mode == renamer.ModeCopy {
			fmt.Printf("if (-not (Test-Path '%s')) { Copy-Item -Path '%s' -Destination '%s' }\n", dst, src, dst)
		} else {
			fmt.Printf("if (-not (Test-Path '%s')) { Move-Item -Path '%s' -Destination '%s' }\n", dst, src, dst)
		}
	}

	fmt.Println()
	fmt.Println("Write-Host 'Done!'")
}

func outputBash(operations []renamer.Operation, mode renamer.OperationMode) {
	fmt.Println("#!/bin/bash")
	fmt.Println("# Generated by Plex File Renamer")
	fmt.Println()

	for _, op := range operations {
		// Escape for bash (single quotes, escape existing single quotes)
		src := strings.ReplaceAll(op.Source, "'", "'\\''")
		dst := strings.ReplaceAll(op.Destination, "'", "'\\''")
		destDir := strings.ReplaceAll(filepath.Dir(op.Destination), "'", "'\\''")

		// Create destination directory
		fmt.Printf("mkdir -p '%s'\n", destDir)

		// Copy or move command (with skip if exists)
		if mode == renamer.ModeCopy {
			fmt.Printf("[ ! -f '%s' ] && cp '%s' '%s'\n", dst, src, dst)
		} else {
			fmt.Printf("[ ! -f '%s' ] && mv '%s' '%s'\n", dst, src, dst)
		}
	}

	fmt.Println()
	fmt.Println("echo 'Done!'")
}

func generateOperations(config *Config, formatter *renamer.Formatter, prompter *cli.Prompter, content *database.LibraryContent, selectedLocations []database.SectionLocation, locationOutputs []cli.LocationWithOutput) ([]renamer.Operation, error) {
	var operations []renamer.Operation

	// Helper to get output path for a file based on its location
	getOutputPath := func(filePath string) string {
		if len(locationOutputs) > 0 {
			for _, lo := range locationOutputs {
				if pathInLocations(filePath, []database.SectionLocation{lo.Location}) {
					return lo.OutputPath
				}
			}
		}
		return config.OutputDir
	}

	switch content.Section.SectionType {
	case database.SectionTypeMovie:
		for _, movie := range content.Movies {
			// Filter by selected locations if specified
			if selectedLocations != nil && !fileInLocations(movie.Files, selectedLocations) {
				continue
			}

			// Generate path previews for this movie
			var previews []cli.PathPreview
			for _, file := range movie.Files {
				if selectedLocations != nil && !pathInLocations(file.File, selectedLocations) {
					continue
				}
				srcPath := file.File
				if config.PathMapSrc != "" {
					srcPath = renamer.ApplyPathMapping(srcPath, config.PathMapSrc, config.PathMapDst)
				}
				ext := renamer.GetExtension(srcPath)
				destName := formatter.FormatMovie(&movie, ext)
				outputDir := getOutputPath(file.File)
				destPath := filepath.Join(outputDir, destName)
				previews = append(previews, cli.PathPreview{Source: srcPath, Destination: destPath})
			}

			if !config.AutoApprove && !config.ScriptMode {
				proceed, _, err := prompter.PromptMovie(&movie, previews)
				if err != nil {
					return nil, err
				}
				if !proceed {
					continue
				}
			}

			// Add operations from previews
			for _, pv := range previews {
				operations = append(operations, renamer.Operation{
					Source:      pv.Source,
					Destination: pv.Destination,
					Mode:        config.Mode,
				})
			}
		}

	case database.SectionTypeShow:
		for _, show := range content.Shows {
			// Filter by selected locations if specified
			if selectedLocations != nil && !showInLocations(&show, selectedLocations) {
				continue
			}

			// Generate path previews for this show
			var previews []cli.PathPreview
			for _, season := range show.Seasons {
				for _, episode := range season.Episodes {
					for _, file := range episode.Files {
						if selectedLocations != nil && !pathInLocations(file.File, selectedLocations) {
							continue
						}
						srcPath := file.File
						if config.PathMapSrc != "" {
							srcPath = renamer.ApplyPathMapping(srcPath, config.PathMapSrc, config.PathMapDst)
						}
						ext := renamer.GetExtension(srcPath)
						destName := formatter.FormatEpisode(&show.Metadata, &season.Metadata, &episode, ext)
						outputDir := getOutputPath(file.File)
						destPath := filepath.Join(outputDir, destName)
						previews = append(previews, cli.PathPreview{Source: srcPath, Destination: destPath})
					}
				}
			}

			if len(previews) == 0 {
				continue
			}

			if !config.AutoApprove && !config.ScriptMode {
				proceed, _, err := prompter.PromptShow(&show, len(previews), previews)
				if err != nil {
					return nil, err
				}
				if !proceed {
					continue
				}
			}

			// Add operations from previews
			for _, pv := range previews {
				operations = append(operations, renamer.Operation{
					Source:      pv.Source,
					Destination: pv.Destination,
					Mode:        config.Mode,
				})
			}
		}
	}

	return operations, nil
}

// pathInLocations checks if a file path is under any of the selected locations
func pathInLocations(filePath string, locations []database.SectionLocation) bool {
	normalizedPath := strings.ToLower(filepath.ToSlash(filePath))
	for _, loc := range locations {
		normalizedLoc := strings.ToLower(filepath.ToSlash(loc.RootPath))
		if strings.HasPrefix(normalizedPath, normalizedLoc) {
			return true
		}
	}
	return false
}

// fileInLocations checks if any file in the list is under any of the selected locations
func fileInLocations(files []database.MediaPart, locations []database.SectionLocation) bool {
	for _, file := range files {
		if pathInLocations(file.File, locations) {
			return true
		}
	}
	return false
}

// showInLocations checks if any episode in the show is under any of the selected locations
func showInLocations(show *database.ShowInfo, locations []database.SectionLocation) bool {
	for _, season := range show.Seasons {
		for _, episode := range season.Episodes {
			if fileInLocations(episode.Files, locations) {
				return true
			}
		}
	}
	return false
}
