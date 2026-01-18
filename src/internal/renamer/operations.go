package renamer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// OperationMode defines how files should be processed
type OperationMode string

const (
	ModeCopy OperationMode = "copy"
	ModeMove OperationMode = "move"
)

// Operation represents a file operation to perform
type Operation struct {
	Source      string
	Destination string
	Mode        OperationMode
}

// Result represents the outcome of an operation
type Result struct {
	Operation Operation
	Success   bool
	Skipped   bool
	Error     error
	Message   string
}

// Execute performs the file operation
func (op *Operation) Execute(dryRun bool) Result {
	result := Result{Operation: *op}

	// In dry-run mode, just report success without checking files
	if dryRun {
		result.Success = true
		result.Message = "dry run - no changes made"
		return result
	}

	// Check if source exists (only when actually executing)
	if _, err := os.Stat(op.Source); os.IsNotExist(err) {
		result.Error = fmt.Errorf("source file does not exist: %s", op.Source)
		return result
	}

	// Check if destination exists (skip if it does)
	if _, err := os.Stat(op.Destination); err == nil {
		result.Skipped = true
		result.Success = true
		result.Message = "destination already exists, skipped"
		return result
	}

	// Create destination directory
	destDir := filepath.Dir(op.Destination)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		result.Error = fmt.Errorf("failed to create directory %s: %w", destDir, err)
		return result
	}

	// Perform the operation
	var err error
	switch op.Mode {
	case ModeCopy:
		err = copyFile(op.Source, op.Destination)
	case ModeMove:
		err = moveFile(op.Source, op.Destination)
	default:
		err = fmt.Errorf("unknown operation mode: %s", op.Mode)
	}

	if err != nil {
		result.Error = err
		return result
	}

	result.Success = true
	result.Message = fmt.Sprintf("%s completed", op.Mode)
	return result
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		// Try to clean up partial file
		os.Remove(dst)
		return fmt.Errorf("failed to copy: %w", err)
	}

	// Preserve file permissions
	sourceInfo, err := os.Stat(src)
	if err == nil {
		os.Chmod(dst, sourceInfo.Mode())
	}

	return nil
}

// moveFile moves a file from src to dst
func moveFile(src, dst string) error {
	// Try rename first (works if same filesystem)
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// Fall back to copy + delete
	if err := copyFile(src, dst); err != nil {
		return err
	}

	// Verify the copy before deleting source
	srcInfo, _ := os.Stat(src)
	dstInfo, err := os.Stat(dst)
	if err != nil {
		return fmt.Errorf("failed to verify copy: %w", err)
	}

	if srcInfo.Size() != dstInfo.Size() {
		os.Remove(dst)
		return fmt.Errorf("copy verification failed: size mismatch")
	}

	// Delete source
	if err := os.Remove(src); err != nil {
		return fmt.Errorf("copied successfully but failed to remove source: %w", err)
	}

	return nil
}

// BatchExecute executes multiple operations and returns results
func BatchExecute(operations []Operation, dryRun bool, progressFn func(current, total int, op Operation)) []Result {
	results := make([]Result, len(operations))
	for i, op := range operations {
		if progressFn != nil {
			progressFn(i+1, len(operations), op)
		}
		results[i] = op.Execute(dryRun)
	}
	return results
}
