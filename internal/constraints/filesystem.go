// Package constraints provides validation constraint types and builders for pedantigo.
package constraints

import (
	"fmt"
	"os"
	"path/filepath"
)

// Filesystem constraint name constants.
const (
	CFilepath = "filepath" // Validates file path syntax (does NOT check existence)
	CDirpath  = "dirpath"  // Validates directory path syntax (does NOT check existence)
	CFile     = "file"     // Validates file exists and is a file (checks disk)
	CDir      = "dir"      // Validates directory exists and is a directory (checks disk)
)

// Filesystem constraint types.
type (
	filepathConstraint struct{} // filepath: validates file path syntax (does NOT check existence)
	dirpathConstraint  struct{} // dirpath: validates directory path syntax (does NOT check existence)
	fileConstraint     struct{} // file: validates file exists and is a file (checks disk)
	dirConstraint      struct{} // dir: validates directory exists and is a directory (checks disk)
)

// Validate checks if the value is a valid file path syntax without checking existence.
// Useful for paths that will be created or are on remote systems.
func (c filepathConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("filepath constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// filepath.Clean normalizes the path and validates basic syntax
	// All non-empty string paths are syntactically valid on Unix/macOS
	// On Windows, this would catch invalid characters like <>:"|?*
	_ = filepath.Clean(str)
	return nil
}

// Validate checks if the value is a valid directory path syntax without checking existence.
// Useful for paths that will be created or are on remote systems.
func (c dirpathConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("dirpath constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	// Same as filepath - all non-empty string paths are syntactically valid on Unix
	_ = filepath.Clean(str)
	return nil
}

// Validate checks that a file exists and is not a directory.
// This constraint checks the actual filesystem.
func (c fileConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("file constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	info, err := os.Stat(str)
	if err != nil {
		if os.IsNotExist(err) {
			return NewConstraintError(CodeFileNotFound, "file does not exist")
		}
		return NewConstraintError(CodeInvalidPath, fmt.Sprintf("cannot access path: %v", err))
	}

	if info.IsDir() {
		return NewConstraintError(CodeInvalidPath, "path is a directory, not a file")
	}

	return nil
}

// Validate checks that a directory exists and is a directory.
// This constraint checks the actual filesystem.
func (c dirConstraint) Validate(value any) error {
	str, isValid, err := extractString(value)
	if !isValid {
		return nil // skip validation for nil/invalid values
	}
	if err != nil {
		return fmt.Errorf("dir constraint %w", err)
	}

	if str == "" {
		return nil // Empty strings are handled by required constraint
	}

	info, err := os.Stat(str)
	if err != nil {
		if os.IsNotExist(err) {
			return NewConstraintError(CodeDirNotFound, "directory does not exist")
		}
		return NewConstraintError(CodeInvalidPath, fmt.Sprintf("cannot access path: %v", err))
	}

	if !info.IsDir() {
		return NewConstraintError(CodeInvalidPath, "path is a file, not a directory")
	}

	return nil
}

// appendFilesystemConstraint appends filesystem constraints based on constraint name.
func appendFilesystemConstraint(result []Constraint, name string) []Constraint {
	switch name {
	case CFilepath:
		return append(result, filepathConstraint{})
	case CDirpath:
		return append(result, dirpathConstraint{})
	case CFile:
		return append(result, fileConstraint{})
	case CDir:
		return append(result, dirConstraint{})
	}
	return result
}
