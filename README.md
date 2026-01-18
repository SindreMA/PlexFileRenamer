# PlexFileRenamer

A CLI tool to rename and organize media files based on Plex Media Server metadata.

## Features

- Reads Plex SQLite database directly (with WAL support)
- Supports both **Movies** and **TV Shows**
- **Dry-run mode** to preview changes without modifying files
- **Script generation** for CMD, PowerShell, and Bash
- **Copy** and **move** operation modes
- Interactive per-library and per-item approval
- Custom filename formats with placeholders
- Path mapping for network shares
- Skips existing files to avoid overwrites

## Installation

Download the latest release for your platform from the [Releases](../../releases) page:

- `plexfilerenamer-windows-amd64.exe` - Windows 64-bit
- `plexfilerenamer-linux-amd64` - Linux 64-bit
- `plexfilerenamer-darwin-amd64` - macOS Intel
- `plexfilerenamer-darwin-arm64` - macOS Apple Silicon

### Build from Source

Requires Go 1.21 or later:

```bash
cd src
go build -o plexfilerenamer ./cmd
```

## Usage

```
plexfilerenamer [options] <database-path>
```

The database path should point to your Plex SQLite database file, typically located at:
- **Windows**: `%LOCALAPPDATA%\Plex Media Server\Plug-in Support\Databases\com.plexapp.plugins.library.db`
- **Linux**: `/var/lib/plexmediaserver/Library/Application Support/Plex Media Server/Plug-in Support/Databases/com.plexapp.plugins.library.db`
- **macOS**: `~/Library/Application Support/Plex Media Server/Plug-in Support/Databases/com.plexapp.plugins.library.db`

### Options

| Option | Description |
|--------|-------------|
| `--output <path>` | Output directory for renamed files (default: source location root) |
| `--dry-run` | Preview changes without applying them |
| `--script` | Generate a shell script instead of executing operations |
| `--shell <type>` | Shell format for script: `cmd`, `powershell`, or `bash` (default: `cmd`) |
| `--script-output <file>` | Output file for script (default: `rename.<ext>` based on shell) |
| `--mode <mode>` | Operation mode: `copy` or `move` (default: `move`) |
| `--tv-format <format>` | Custom format for TV show filenames |
| `--movie-format <format>` | Custom format for movie filenames |
| `--path-map <old:new>` | Path mapping for network shares |
| `--auto-approve` | Skip interactive prompts, process all items |

### Format Placeholders

**TV Shows** (default: `{show}/Season {season}/S{snum}E{enum} - {title}{ext}`):
- `{show}` - Series title
- `{season}` - Season number
- `{snum}` - Season number (2-digit, zero-padded)
- `{enum}` - Episode number (2-digit, zero-padded)
- `{title}` - Episode title
- `{year}` - Show's release year
- `{ext}` - File extension (e.g., `.mkv`)

**Movies** (default: `{title} ({year}){ext}`):
- `{title}` - Movie title
- `{year}` - Release year
- `{ext}` - File extension

## Examples

### Preview changes (dry run)

```bash
plexfilerenamer --dry-run /path/to/com.plexapp.plugins.library.db
```

### Copy files to a new location

```bash
plexfilerenamer --mode copy --output /media/organized /path/to/plex.db
```

### Generate a PowerShell script

```bash
plexfilerenamer --script --shell powershell /path/to/plex.db
```

This creates a `rename.ps1` file you can review and execute later.

### Use path mapping for network shares

If Plex sees files at `F:\Media` but your machine accesses them at `H:\Media`:

```bash
plexfilerenamer --path-map "F:\Media:H:\Media" /path/to/plex.db
```

### Custom TV format

```bash
plexfilerenamer --tv-format "{show} - S{snum}E{enum} - {title}{ext}" /path/to/plex.db
```

### Auto-approve all operations

```bash
plexfilerenamer --auto-approve --script /path/to/plex.db
```

## How It Works

1. Opens the Plex database in read-only mode (safe to run while Plex is running)
2. Reads library sections, locations, and media metadata
3. For each library, prompts you to select which locations to process
4. For each movie/show, displays the proposed rename and asks for approval
5. Executes the operations (or generates a script in `--script` mode)

## Notes

- The tool reads the database in **immutable mode**, so it's safe to use while Plex is running
- Files that already exist at the destination are automatically skipped
- Invalid filename characters are automatically sanitized (e.g., `:` becomes ` -`)
- The tool handles Windows long path prefixes (`\\?\`) used by Plex

## License

MIT
