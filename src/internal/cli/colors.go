package cli

import (
	"fmt"

	"github.com/pterm/pterm"
)

// Styled printers for consistent output
var (
	HeaderStyle    = pterm.NewStyle(pterm.FgCyan, pterm.Bold)
	SubHeaderStyle = pterm.NewStyle(pterm.FgYellow, pterm.Bold)
	SuccessStyle   = pterm.NewStyle(pterm.FgGreen)
	ErrorStyle     = pterm.NewStyle(pterm.FgRed)
	WarningStyle   = pterm.NewStyle(pterm.FgYellow)
	InfoStyle      = pterm.NewStyle(pterm.FgBlue)
	DimStyle       = pterm.NewStyle(pterm.FgGray)
	PathStyle      = pterm.NewStyle(pterm.FgCyan)
	AccentStyle    = pterm.NewStyle(pterm.FgMagenta, pterm.Bold)
)

// PrintHeader prints a prominent header
func PrintHeader(text string) {
	pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgCyan)).
		WithTextStyle(pterm.NewStyle(pterm.FgBlack, pterm.Bold)).
		Println(text)
}

// PrintSubHeader prints a sub-section header
func PrintSubHeader(text string) {
	SubHeaderStyle.Println(text)
}

// PrintSuccess prints a success message with checkmark
func PrintSuccess(text string) {
	pterm.Success.Println(text)
}

// PrintError prints an error message
func PrintError(text string) {
	pterm.Error.Println(text)
}

// PrintWarning prints a warning message
func PrintWarning(text string) {
	pterm.Warning.Println(text)
}

// PrintInfo prints an info message
func PrintInfo(text string) {
	pterm.Info.Println(text)
}

// PrintDim prints dimmed/muted text
func PrintDim(text string) {
	DimStyle.Println(text)
}

// PrintPath prints a file path with styling
func PrintPath(text string) {
	PathStyle.Print(text)
}

// PrintLabel prints a label: value pair
func PrintLabel(label, value string) {
	fmt.Printf("%s %s\n", DimStyle.Sprint(label+":"), value)
}

// PrintNumberedItem prints a numbered list item
func PrintNumberedItem(num int, text string) {
	fmt.Printf("  %s %s\n", AccentStyle.Sprintf("[%d]", num), text)
}

// PrintBullet prints a bullet point item
func PrintBullet(text string) {
	pterm.Println("  " + pterm.FgGray.Sprint("•") + " " + text)
}

// CreateProgressBar creates a new progress bar
func CreateProgressBar(total int, title string) (*pterm.ProgressbarPrinter, error) {
	return pterm.DefaultProgressbar.
		WithTotal(total).
		WithTitle(title).
		WithShowCount(true).
		WithShowPercentage(true).
		WithShowElapsedTime(false).
		Start()
}

// PrintOperationTable prints operations in a table format
func PrintOperationTable(data [][]string) {
	table := pterm.TableData{{"#", "Source", "Destination"}}
	table = append(table, data...)
	pterm.DefaultTable.WithHasHeader().WithData(table).Render()
}

// PrintResultsBox prints results in a styled box
func PrintResultsBox(succeeded, skipped, failed int) {
	content := fmt.Sprintf(
		"%s %d   %s %d   %s %d",
		pterm.FgGreen.Sprint("Succeeded:"), succeeded,
		pterm.FgYellow.Sprint("Skipped:"), skipped,
		pterm.FgRed.Sprint("Failed:"), failed,
	)
	pterm.DefaultBox.WithTitle("Results").Println(content)
}

// PrintBanner prints the application banner
func PrintBanner() {
	pterm.DefaultBigText.WithLetters(
		pterm.NewLettersFromStringWithStyle("Plex", pterm.NewStyle(pterm.FgCyan)),
		pterm.NewLettersFromStringWithStyle("File", pterm.NewStyle(pterm.FgLightMagenta)),
		pterm.NewLettersFromStringWithStyle("Renamer", pterm.NewStyle(pterm.FgCyan)),
	).Render()
	DimStyle.Println("v1.0 - Rename media files using Plex metadata")
	fmt.Println()
}

// Confirm shows a confirmation prompt
func Confirm(prompt string) (bool, error) {
	return pterm.DefaultInteractiveConfirm.
		WithDefaultText(prompt).
		Show()
}

// SelectOption shows an interactive selection
func SelectOption(prompt string, options []string) (string, error) {
	return pterm.DefaultInteractiveSelect.
		WithDefaultText(prompt).
		WithOptions(options).
		Show()
}

// Styled text helpers for inline use
func Dim(text string) string {
	return DimStyle.Sprint(text)
}

func Accent(text string) string {
	return AccentStyle.Sprint(text)
}

func Success(text string) string {
	return SuccessStyle.Sprint(text)
}

func Error(text string) string {
	return ErrorStyle.Sprint(text)
}

func Warning(text string) string {
	return WarningStyle.Sprint(text)
}

func Path(text string) string {
	return PathStyle.Sprint(text)
}
