package database

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// PlexDB provides access to the Plex Media Server database
type PlexDB struct {
	db *sql.DB
}

// Open opens a Plex database file
func Open(dbPath string) (*PlexDB, error) {
	// Use file: URI with read-only mode and immutable flag for WAL databases
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Convert Windows paths for SQLite URI
	absPath = strings.ReplaceAll(absPath, "\\", "/")

	// Use immutable=1 to handle WAL mode databases that might be in use
	// This allows reading even if WAL files are present
	uri := fmt.Sprintf("file:%s?mode=ro&immutable=1", absPath)

	db, err := sql.Open("sqlite", uri)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection to read-only and disable WAL checkpoint
	db.SetMaxOpenConns(1)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PlexDB{db: db}, nil
}

// Close closes the database connection
func (p *PlexDB) Close() error {
	return p.db.Close()
}

// GetLibrarySections returns all library sections
func (p *PlexDB) GetLibrarySections() ([]LibrarySection, error) {
	query := `
		SELECT id, name, section_type, language, agent
		FROM library_sections
		ORDER BY name
	`

	rows, err := p.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query library sections: %w", err)
	}
	defer rows.Close()

	var sections []LibrarySection
	for rows.Next() {
		var s LibrarySection
		if err := rows.Scan(&s.ID, &s.Name, &s.SectionType, &s.Language, &s.Agent); err != nil {
			return nil, fmt.Errorf("failed to scan library section: %w", err)
		}
		sections = append(sections, s)
	}

	return sections, rows.Err()
}

// GetSectionLocations returns all root paths for a library section
func (p *PlexDB) GetSectionLocations(sectionID int64) ([]SectionLocation, error) {
	query := `
		SELECT id, library_section_id, root_path, available
		FROM section_locations
		WHERE library_section_id = ?
	`

	rows, err := p.db.Query(query, sectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query section locations: %w", err)
	}
	defer rows.Close()

	var locations []SectionLocation
	for rows.Next() {
		var l SectionLocation
		if err := rows.Scan(&l.ID, &l.LibrarySectionID, &l.RootPath, &l.Available); err != nil {
			return nil, fmt.Errorf("failed to scan section location: %w", err)
		}
		locations = append(locations, l)
	}

	return locations, rows.Err()
}

// GetMetadataItems returns metadata items for a section of a specific type
func (p *PlexDB) GetMetadataItems(sectionID int64, metadataType int) ([]MetadataItem, error) {
	query := `
		SELECT id, library_section_id, metadata_type,
		       parent_id,
		       title, title_sort, COALESCE(original_title, ''),
		       COALESCE(studio, ''), year, "index",
		       COALESCE(originally_available_at, '')
		FROM metadata_items
		WHERE library_section_id = ? AND metadata_type = ?
		ORDER BY title_sort
	`

	rows, err := p.db.Query(query, sectionID, metadataType)
	if err != nil {
		return nil, fmt.Errorf("failed to query metadata items: %w", err)
	}
	defer rows.Close()

	var items []MetadataItem
	for rows.Next() {
		var m MetadataItem
		if err := rows.Scan(
			&m.ID, &m.LibrarySectionID, &m.MetadataType,
			&m.ParentID,
			&m.Title, &m.TitleSort, &m.OriginalTitle,
			&m.Studio, &m.Year, &m.Index,
			&m.OriginallyAvailable,
		); err != nil {
			return nil, fmt.Errorf("failed to scan metadata item: %w", err)
		}
		items = append(items, m)
	}

	return items, rows.Err()
}

// GetChildMetadata returns child metadata items (episodes for a season, seasons for a show)
func (p *PlexDB) GetChildMetadata(parentID int64) ([]MetadataItem, error) {
	query := `
		SELECT id, library_section_id, metadata_type,
		       parent_id,
		       title, title_sort, COALESCE(original_title, ''),
		       COALESCE(studio, ''), year, "index",
		       COALESCE(originally_available_at, '')
		FROM metadata_items
		WHERE parent_id = ?
		ORDER BY "index"
	`

	rows, err := p.db.Query(query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query child metadata: %w", err)
	}
	defer rows.Close()

	var items []MetadataItem
	for rows.Next() {
		var m MetadataItem
		if err := rows.Scan(
			&m.ID, &m.LibrarySectionID, &m.MetadataType,
			&m.ParentID,
			&m.Title, &m.TitleSort, &m.OriginalTitle,
			&m.Studio, &m.Year, &m.Index,
			&m.OriginallyAvailable,
		); err != nil {
			return nil, fmt.Errorf("failed to scan child metadata: %w", err)
		}
		items = append(items, m)
	}

	return items, rows.Err()
}

// GetMediaParts returns all file paths for a metadata item
func (p *PlexDB) GetMediaParts(metadataItemID int64) ([]MediaPart, error) {
	query := `
		SELECT mp.id, mp.media_item_id, mp.file, COALESCE(mp.size, 0)
		FROM media_parts mp
		JOIN media_items mi ON mp.media_item_id = mi.id
		WHERE mi.metadata_item_id = ?
	`

	rows, err := p.db.Query(query, metadataItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to query media parts: %w", err)
	}
	defer rows.Close()

	var parts []MediaPart
	for rows.Next() {
		var mp MediaPart
		if err := rows.Scan(&mp.ID, &mp.MediaItemID, &mp.File, &mp.Size); err != nil {
			return nil, fmt.Errorf("failed to scan media part: %w", err)
		}
		parts = append(parts, mp)
	}

	return parts, rows.Err()
}

// GetLibraryContent returns all content for a library section
func (p *PlexDB) GetLibraryContent(section LibrarySection) (*LibraryContent, error) {
	content := &LibraryContent{Section: section}

	// Get locations
	locations, err := p.GetSectionLocations(section.ID)
	if err != nil {
		return nil, err
	}
	content.Locations = locations

	switch section.SectionType {
	case SectionTypeMovie:
		movies, err := p.getMovies(section.ID)
		if err != nil {
			return nil, err
		}
		content.Movies = movies

	case SectionTypeShow:
		shows, err := p.getShows(section.ID)
		if err != nil {
			return nil, err
		}
		content.Shows = shows
	}

	return content, nil
}

func (p *PlexDB) getMovies(sectionID int64) ([]MovieInfo, error) {
	items, err := p.GetMetadataItems(sectionID, MediaTypeMovie)
	if err != nil {
		return nil, err
	}

	var movies []MovieInfo
	for _, item := range items {
		files, err := p.GetMediaParts(item.ID)
		if err != nil {
			return nil, err
		}
		movies = append(movies, MovieInfo{
			Metadata: item,
			Files:    files,
		})
	}

	return movies, nil
}

func (p *PlexDB) getShows(sectionID int64) ([]ShowInfo, error) {
	shows, err := p.GetMetadataItems(sectionID, MediaTypeShow)
	if err != nil {
		return nil, err
	}

	var showInfos []ShowInfo
	for _, show := range shows {
		seasons, err := p.getSeasons(show.ID)
		if err != nil {
			return nil, err
		}
		showInfos = append(showInfos, ShowInfo{
			Metadata: show,
			Seasons:  seasons,
		})
	}

	return showInfos, nil
}

func (p *PlexDB) getSeasons(showID int64) ([]SeasonInfo, error) {
	seasons, err := p.GetChildMetadata(showID)
	if err != nil {
		return nil, err
	}

	var seasonInfos []SeasonInfo
	for _, season := range seasons {
		episodes, err := p.getEpisodes(season.ID)
		if err != nil {
			return nil, err
		}
		seasonInfos = append(seasonInfos, SeasonInfo{
			Metadata: season,
			Episodes: episodes,
		})
	}

	return seasonInfos, nil
}

func (p *PlexDB) getEpisodes(seasonID int64) ([]EpisodeInfo, error) {
	episodes, err := p.GetChildMetadata(seasonID)
	if err != nil {
		return nil, err
	}

	var episodeInfos []EpisodeInfo
	for _, episode := range episodes {
		files, err := p.GetMediaParts(episode.ID)
		if err != nil {
			return nil, err
		}
		episodeInfos = append(episodeInfos, EpisodeInfo{
			Metadata: episode,
			Files:    files,
		})
	}

	return episodeInfos, nil
}
