# Plex File Renamer - Development Progress

## Current Status: COMPLETE - v1.0

### Completed
- [x] Created project plan
- [x] Created PROGRESS.md
- [x] Initialize Go module
- [x] Create directory structure in `src/`
- [x] Add SQLite dependency (modernc.org/sqlite)
- [x] Database models (models.go)
- [x] Plex database reader (plex.go)
- [x] Name formatter (formatter.go)
- [x] File operations with copy/move (operations.go)
- [x] Interactive CLI prompts (interactive.go)
- [x] Main CLI entry point (main.go)
- [x] Testing with sample database - SUCCESS

---

## Architecture

```
src/
  cmd/
    main.go              - CLI entry point
  internal/
    database/
      plex.go            - Plex SQLite reader
      models.go          - Database structs
    renamer/
      formatter.go       - Name formatting
      operations.go      - File copy/move
    cli/
      interactive.go     - User prompts
  go.mod
  go.sum
```

## Usage

```bash
# Build
cd src && go build -o ../plexrenamer.exe ./cmd

# Run dry-run preview
./plexrenamer.exe --dry-run --output ./output path/to/plex.db

# Run with auto-approve (no prompts)
./plexrenamer.exe --auto-approve --output ./output path/to/plex.db

# Copy mode (instead of move)
./plexrenamer.exe --mode copy --output ./output path/to/plex.db

# Path mapping for network shares
./plexrenamer.exe --path-map "F:\Media:H:\Media" --output ./output path/to/plex.db

# Generate CMD batch script (redirect to file)
./plexrenamer.exe --script --shell cmd --output ./output path/to/plex.db > rename.bat

# Generate PowerShell script
./plexrenamer.exe --script --shell powershell --output ./output path/to/plex.db > rename.ps1

# Generate Bash script
./plexrenamer.exe --script --shell bash --output ./output path/to/plex.db > rename.sh
```

## Features Implemented

- Reads Plex SQLite database (with WAL support)
- Supports Movies and TV Shows
- Dry-run mode to preview changes
- Copy and move modes
- Interactive per-library approval
- "Approve all" option
- Custom name formats
- Season-based folder organization for TV
- Path replacement for network shares
- Skip-on-conflict for existing files
- **Script mode**: Output shell commands instead of executing
  - Supports CMD (.bat), PowerShell (.ps1), and Bash (.sh)

## Notes

- Using `modernc.org/sqlite` (pure Go, no CGO needed)
- Handles Plex WAL databases with immutable mode
- Sanitizes filenames for Windows compatibility
