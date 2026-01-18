package renamer

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"plexrenamer/internal/database"
)

// DefaultTVFormat is the default format for TV show episodes
const DefaultTVFormat = "{show}/Season {season}/S{snum}E{enum} - {title}{ext}"

// DefaultMovieFormat is the default format for movies
const DefaultMovieFormat = "{title} ({year}){ext}"

// Formatter handles filename generation from metadata
type Formatter struct {
	TVFormat    string
	MovieFormat string
}

// NewFormatter creates a new formatter with the specified formats
func NewFormatter(tvFormat, movieFormat string) *Formatter {
	if tvFormat == "" {
		tvFormat = DefaultTVFormat
	}
	if movieFormat == "" {
		movieFormat = DefaultMovieFormat
	}
	return &Formatter{
		TVFormat:    tvFormat,
		MovieFormat: movieFormat,
	}
}

// FormatEpisode generates a filename for a TV episode
func (f *Formatter) FormatEpisode(show, season *database.MetadataItem, episode *database.EpisodeInfo, ext string) string {
	result := f.TVFormat

	// Show title
	result = strings.ReplaceAll(result, "{show}", sanitizeFilename(show.Title))

	// Season number
	seasonNum := 0
	if season.Index != nil {
		seasonNum = *season.Index
	}
	result = strings.ReplaceAll(result, "{season}", fmt.Sprintf("%d", seasonNum))
	result = strings.ReplaceAll(result, "{snum}", fmt.Sprintf("%02d", seasonNum))

	// Episode number
	episodeNum := 0
	if episode.Metadata.Index != nil {
		episodeNum = *episode.Metadata.Index
	}
	result = strings.ReplaceAll(result, "{enum}", fmt.Sprintf("%02d", episodeNum))

	// Episode title
	result = strings.ReplaceAll(result, "{title}", sanitizeFilename(episode.Metadata.Title))

	// Year (if available)
	year := ""
	if show.Year != nil {
		year = fmt.Sprintf("%d", *show.Year)
	}
	result = strings.ReplaceAll(result, "{year}", year)

	// Extension
	result = strings.ReplaceAll(result, "{ext}", ext)

	return result
}

// FormatMovie generates a filename for a movie
func (f *Formatter) FormatMovie(movie *database.MovieInfo, ext string) string {
	result := f.MovieFormat

	// Movie title
	result = strings.ReplaceAll(result, "{title}", sanitizeFilename(movie.Metadata.Title))

	// Year
	year := "Unknown"
	if movie.Metadata.Year != nil {
		year = fmt.Sprintf("%d", *movie.Metadata.Year)
	}
	result = strings.ReplaceAll(result, "{year}", year)

	// Extension
	result = strings.ReplaceAll(result, "{ext}", ext)

	return result
}

// sanitizeFilename removes or replaces characters that are invalid in filenames
func sanitizeFilename(name string) string {
	// Characters not allowed in Windows filenames: \ / : * ? " < > |
	// Also handle some other problematic characters

	// Replace some characters with alternatives
	replacements := map[string]string{
		":":  " -",
		"/":  "-",
		"\\": "-",
		"*":  "",
		"?":  "",
		"\"": "'",
		"<":  "",
		">":  "",
		"|":  "-",
	}

	result := name
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	// Remove any control characters
	result = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, result)

	// Trim spaces and dots from the end (Windows doesn't like trailing dots)
	result = strings.TrimRight(result, " .")

	// Collapse multiple spaces
	spaceRegex := regexp.MustCompile(`\s+`)
	result = spaceRegex.ReplaceAllString(result, " ")

	return strings.TrimSpace(result)
}

// ApplyPathMapping replaces the source path prefix with destination prefix
func ApplyPathMapping(path, srcPrefix, dstPrefix string) string {
	if srcPrefix == "" || dstPrefix == "" {
		return path
	}

	// Normalize path separators for comparison
	normalizedPath := filepath.ToSlash(path)
	normalizedSrc := filepath.ToSlash(srcPrefix)

	if strings.HasPrefix(normalizedPath, normalizedSrc) {
		newPath := dstPrefix + normalizedPath[len(normalizedSrc):]
		// Convert back to OS-specific path
		return filepath.FromSlash(newPath)
	}

	return path
}

// GetExtension extracts the file extension including the dot
func GetExtension(path string) string {
	return filepath.Ext(path)
}
