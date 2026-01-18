package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"plexrenamer/internal/database"
	"plexrenamer/internal/renamer"
)

// ApprovalState tracks user approval choices
type ApprovalState struct {
	ApproveAll     bool
	ApprovedShows  map[int64]bool // Show ID -> approved
	SkippedShows   map[int64]bool // Show ID -> skipped
	ApprovedMovies map[int64]bool // Movie ID -> approved
}

// NewApprovalState creates a new approval state
func NewApprovalState() *ApprovalState {
	return &ApprovalState{
		ApprovedShows:  make(map[int64]bool),
		SkippedShows:   make(map[int64]bool),
		ApprovedMovies: make(map[int64]bool),
	}
}

// Prompter handles user interaction
type Prompter struct {
	reader *bufio.Reader
	state  *ApprovalState
}

// NewPrompter creates a new prompter
func NewPrompter() *Prompter {
	return &Prompter{
		reader: bufio.NewReader(os.Stdin),
		state:  NewApprovalState(),
	}
}

// PromptLibrary asks user if they want to process a library
// Returns: proceed, selectedLocations (nil means all), error
func (p *Prompter) PromptLibrary(section database.LibrarySection, locations []database.SectionLocation) (bool, []database.SectionLocation, error) {
	fmt.Println()
	PrintHeader(section.Name)

	sectionType := "Unknown"
	switch section.SectionType {
	case database.SectionTypeMovie:
		sectionType = "Movies"
	case database.SectionTypeShow:
		sectionType = "TV Shows"
	}
	PrintLabel("Type", sectionType)
	PrintLabel("Locations", fmt.Sprintf("%d", len(locations)))

	fmt.Println()
	for i, loc := range locations {
		PrintNumberedItem(i+1, Path(loc.RootPath))
	}
	fmt.Println()

	fmt.Print(pterm.FgWhite.Sprint("Process this library? ") + Dim("[y/n/l(oop)/1-N]: "))
	input, err := p.reader.ReadString('\n')
	if err != nil {
		return false, nil, err
	}

	input = strings.TrimSpace(strings.ToLower(input))

	switch input {
	case "y", "yes":
		return true, nil, nil // nil means all locations
	case "l", "loop":
		// Loop through each location with y/n/a prompts
		return p.promptLocationsLoop(locations)
	case "n", "no":
		return false, nil, nil
	default:
		// Try to parse as location number(s) - comma separated
		var selected []database.SectionLocation
		parts := strings.Split(input, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			var idx int
			if _, err := fmt.Sscanf(part, "%d", &idx); err == nil {
				if idx >= 1 && idx <= len(locations) {
					selected = append(selected, locations[idx-1])
				}
			}
		}
		if len(selected) > 0 {
			return true, selected, nil
		}
		return false, nil, nil
	}
}

// LocationWithOutput pairs a location with its custom output path
type LocationWithOutput struct {
	Location   database.SectionLocation
	OutputPath string // Custom output path for this location (empty = use default)
}

// promptLocationsLoop asks about each location one by one
func (p *Prompter) promptLocationsLoop(locations []database.SectionLocation) (bool, []database.SectionLocation, error) {
	var selected []database.SectionLocation
	approveAll := false

	for i, loc := range locations {
		if approveAll {
			selected = append(selected, loc)
			continue
		}

		fmt.Println()
		fmt.Printf("  %s %s\n", Dim(fmt.Sprintf("[%d/%d]", i+1, len(locations))), Path(loc.RootPath))
		yes, all, err := p.askYesNoAll("  Process this location?")
		if err != nil {
			return false, nil, err
		}

		if all {
			approveAll = true
			selected = append(selected, loc)
		} else if yes {
			selected = append(selected, loc)
		}
	}

	if len(selected) == 0 {
		return false, nil, nil
	}

	return true, selected, nil
}

// PromptLocationOutputs asks for custom output paths for each selected location
func (p *Prompter) PromptLocationOutputs(locations []database.SectionLocation, defaultOutput string) ([]LocationWithOutput, error) {
	var results []LocationWithOutput

	fmt.Println()
	PrintSubHeader("Set output paths for each location")
	PrintDim(fmt.Sprintf("  Default output: %s", defaultOutput))
	PrintDim("  Press Enter to use default, or type a custom path")
	fmt.Println()

	for i, loc := range locations {
		fmt.Printf("  %s %s\n", Dim(fmt.Sprintf("[%d/%d]", i+1, len(locations))), Path(loc.RootPath))
		fmt.Print(pterm.FgWhite.Sprint("  Output path: "))

		input, err := p.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		input = strings.TrimSpace(input)
		outputPath := defaultOutput
		if input != "" {
			outputPath = input
		}

		results = append(results, LocationWithOutput{
			Location:   loc,
			OutputPath: outputPath,
		})
		fmt.Printf("    %s %s\n\n", pterm.FgGreen.Sprint("→"), Path(outputPath))
	}

	return results, nil
}

// PromptShow asks user if they want to process a show
func (p *Prompter) PromptShow(show *database.ShowInfo, episodeCount int, previews []PathPreview) (bool, bool, error) {
	if p.state.ApproveAll {
		return true, false, nil
	}

	fmt.Println()
	PrintSubHeader(fmt.Sprintf("TV Show: %s", show.Metadata.Title))
	if show.Metadata.Year != nil {
		PrintLabel("Year", fmt.Sprintf("%d", *show.Metadata.Year))
	}
	fmt.Printf("  %s %d  %s %d\n",
		Dim("Seasons:"), len(show.Seasons),
		Dim("Episodes:"), episodeCount)

	// Show sample path previews (limit to 3 examples)
	if len(previews) > 0 {
		fmt.Println()
		showCount := len(previews)
		if showCount > 3 {
			showCount = 3
		}
		for i := 0; i < showCount; i++ {
			pv := previews[i]
			fmt.Printf("  %s %s\n", pterm.FgRed.Sprint("From:"), Dim(pv.Source))
			fmt.Printf("  %s %s\n", pterm.FgGreen.Sprint("To:  "), Path(pv.Destination))
			fmt.Println()
		}
		if len(previews) > 3 {
			PrintDim(fmt.Sprintf("  ... and %d more files", len(previews)-3))
		}
	}

	return p.askYesNoAll("Rename files for this show?")
}

// PathPreview holds source and destination path for preview
type PathPreview struct {
	Source      string
	Destination string
}

// PromptMovie asks user if they want to process a movie
func (p *Prompter) PromptMovie(movie *database.MovieInfo, previews []PathPreview) (bool, bool, error) {
	if p.state.ApproveAll {
		return true, false, nil
	}

	fmt.Println()
	PrintSubHeader(fmt.Sprintf("Movie: %s", movie.Metadata.Title))
	if movie.Metadata.Year != nil {
		PrintLabel("Year", fmt.Sprintf("%d", *movie.Metadata.Year))
	}
	fmt.Printf("  %s %d\n", Dim("Files:"), len(movie.Files))

	// Show path previews
	if len(previews) > 0 {
		fmt.Println()
		for _, pv := range previews {
			fmt.Printf("  %s %s\n", pterm.FgRed.Sprint("From:"), Dim(pv.Source))
			fmt.Printf("  %s %s\n", pterm.FgGreen.Sprint("To:  "), Path(pv.Destination))
			if len(previews) > 1 {
				fmt.Println()
			}
		}
	}

	return p.askYesNoAll("Rename files for this movie?")
}

// ShowOperationPreview displays what operations will be performed
func ShowOperationPreview(operations []renamer.Operation, limit int) {
	fmt.Println()
	pterm.DefaultSection.Println("Planned Operations")

	count := len(operations)
	if limit > 0 && count > limit {
		count = limit
	}

	for i := 0; i < count; i++ {
		op := operations[i]
		fmt.Printf("  %s %s\n", pterm.FgRed.Sprint("From:"), Dim(op.Source))
		fmt.Printf("  %s %s\n", pterm.FgGreen.Sprint("To:  "), Path(op.Destination))
		fmt.Println()
	}

	if limit > 0 && len(operations) > limit {
		PrintDim(fmt.Sprintf("  ... and %d more operations", len(operations)-limit))
	}
}

// ShowResults displays the results of operations
func ShowResults(results []renamer.Result) {
	var succeeded, skipped, failed int
	var failures []renamer.Result

	for _, r := range results {
		if r.Error != nil {
			failed++
			failures = append(failures, r)
		} else if r.Skipped {
			skipped++
		} else if r.Success {
			succeeded++
		}
	}

	fmt.Println()
	PrintResultsBox(succeeded, skipped, failed)

	// Show failures in detail
	if failed > 0 {
		fmt.Println()
		pterm.Error.Println("Failed operations:")
		for _, r := range failures {
			fmt.Printf("  %s\n", r.Operation.Source)
			fmt.Printf("    %s %s\n", pterm.FgRed.Sprint("Error:"), r.Error)
		}
	}
}

// ConfirmProceed asks user to confirm before executing
func (p *Prompter) ConfirmProceed(operationCount int, mode renamer.OperationMode, dryRun bool) (bool, error) {
	fmt.Println()

	if dryRun {
		pterm.Info.Printf("DRY RUN: Would %s %d files\n", mode, operationCount)
		return true, nil
	}

	pterm.Warning.Printf("About to %s %d files. This cannot be undone.\n", mode, operationCount)
	return p.askYesNo("Proceed?")
}

func (p *Prompter) askYesNo(prompt string) (bool, error) {
	fmt.Print(pterm.FgWhite.Sprint(prompt) + Dim(" [y/n]: "))
	input, err := p.reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes", nil
}

func (p *Prompter) askYesNoAll(prompt string) (yes bool, approveAll bool, err error) {
	fmt.Print(pterm.FgWhite.Sprint(prompt) + Dim(" [y/n/a(ll)]: "))
	input, err := p.reader.ReadString('\n')
	if err != nil {
		return false, false, err
	}

	input = strings.TrimSpace(strings.ToLower(input))
	switch input {
	case "y", "yes":
		return true, false, nil
	case "a", "all":
		p.state.ApproveAll = true
		return true, true, nil
	default:
		return false, false, nil
	}
}

// PrintProgress shows progress during operations (callback for BatchExecute)
func PrintProgress(current, total int, op renamer.Operation) {
	// This is the old callback-style progress, replaced by progress bar
	fmt.Printf("\r%s [%d/%d] %s",
		Dim("Processing:"),
		current, total,
		truncatePath(op.Source, 50))
}

func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return "..." + path[len(path)-maxLen+3:]
}
