package database

// LibrarySection represents a Plex library (e.g., "Movies", "TV Shows")
type LibrarySection struct {
	ID          int64
	Name        string
	SectionType int // 1 = movie, 2 = show
	Language    string
	Agent       string
}

// SectionLocation represents a root path for a library section
type SectionLocation struct {
	ID               int64
	LibrarySectionID int64
	RootPath         string
	Available        int
}

// MetadataItem represents an item in the Plex library (movie, show, season, episode)
type MetadataItem struct {
	ID                  int64
	LibrarySectionID    int64
	MetadataType        int // 1 = movie, 2 = show, 3 = season, 4 = episode
	ParentID            *int64
	Title               string
	TitleSort           string
	OriginalTitle       string
	Studio              string
	Year                *int
	Index               *int // Episode/season number
	OriginallyAvailable string
}

// MediaItem links metadata to physical media files
type MediaItem struct {
	ID             int64
	MetadataItemID int64
	Width          int
	Height         int
	Bitrate        int
	Container      string
	VideoCodec     string
	AudioCodec     string
}

// MediaPart represents a physical file on disk
type MediaPart struct {
	ID          int64
	MediaItemID int64
	File        string // Full file path
	Size        int64
}

// MediaType constants
const (
	MediaTypeMovie   = 1
	MediaTypeShow    = 2
	MediaTypeSeason  = 3
	MediaTypeEpisode = 4
)

// SectionType constants
const (
	SectionTypeMovie = 1
	SectionTypeShow  = 2
)

// RenameOperation represents a single file rename/move operation
type RenameOperation struct {
	SourcePath string
	DestPath   string
	MediaItem  *MetadataItem
	Show       *MetadataItem // Parent show (for TV)
	Season     *MetadataItem // Parent season (for TV)
}

// LibraryContent holds all parsed content for a library
type LibraryContent struct {
	Section   LibrarySection
	Locations []SectionLocation
	Movies    []MovieInfo
	Shows     []ShowInfo
}

// MovieInfo holds movie metadata with file info
type MovieInfo struct {
	Metadata MetadataItem
	Files    []MediaPart
}

// ShowInfo holds TV show metadata with seasons and episodes
type ShowInfo struct {
	Metadata MetadataItem
	Seasons  []SeasonInfo
}

// SeasonInfo holds season metadata with episodes
type SeasonInfo struct {
	Metadata MetadataItem
	Episodes []EpisodeInfo
}

// EpisodeInfo holds episode metadata with file info
type EpisodeInfo struct {
	Metadata MetadataItem
	Files    []MediaPart
}
