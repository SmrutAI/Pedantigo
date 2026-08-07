package constraints

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFilepathConstraint tests filepathConstraint.Validate() for valid file path syntax.
// This constraint validates path syntax only, NOT existence.
func TestFilepathConstraint(t *testing.T) {
	runSimpleConstraintTests(t, filepathConstraint{}, []simpleTestCase{
		// Valid file path syntax (existence not checked)
		{"valid absolute path /foo/bar/file.txt", "/foo/bar/file.txt", false},
		{"valid relative path foo/bar/file.txt", "foo/bar/file.txt", false},
		{"valid simple filename file.txt", "file.txt", false},
		{"valid hidden file .hidden", ".hidden", false},
		{"valid with spaces /path/to/my file.txt", "/path/to/my file.txt", false},
		{"valid home tilde ~/file.txt", "~/file.txt", false},
		{"valid current dir ./file.txt", "./file.txt", false},
		{"valid parent dir ../file.txt", "../file.txt", false},
		{"valid dot", ".", false},
		{"valid double dot", "..", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestDirpathConstraint tests dirpathConstraint.Validate() for valid directory path syntax.
// This constraint validates path syntax only, NOT existence.
func TestDirpathConstraint(t *testing.T) {
	runSimpleConstraintTests(t, dirpathConstraint{}, []simpleTestCase{
		// Valid directory path syntax (existence not checked)
		{"valid absolute path /foo/bar", "/foo/bar", false},
		{"valid relative path foo/bar", "foo/bar", false},
		{"valid with trailing slash /foo/bar/", "/foo/bar/", false},
		{"valid single dir mydir", "mydir", false},
		{"valid hidden dir .hidden", ".hidden", false},
		{"valid with spaces /path/to/my folder", "/path/to/my folder", false},
		{"valid home tilde ~/dir", "~/dir", false},
		{"valid current dir .", ".", false},
		{"valid parent dir ..", "..", false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
	})
}

// TestFileConstraint tests fileConstraint.Validate() that file exists.
// This constraint checks the actual filesystem.
func TestFileConstraint(t *testing.T) {
	// Create a temp file for testing
	tmpFile, err := os.CreateTemp("", "test_file_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpFilePath)

	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "test_dir_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	runSimpleConstraintTests(t, fileConstraint{}, []simpleTestCase{
		// Valid - file exists
		{"valid existing file", tmpFilePath, false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid - file does not exist
		{"invalid nonexistent file", "/nonexistent/path/to/file.txt", true},
		{"invalid nonexistent in temp", filepath.Join(os.TempDir(), "nonexistent_file_12345.txt"), true},
		// Invalid - path is a directory, not a file
		{"invalid directory not file", tmpDir, true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
	})
}

// TestDirConstraint tests dirConstraint.Validate() that directory exists.
// This constraint checks the actual filesystem.
func TestDirConstraint(t *testing.T) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "test_dir_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a temp file for testing
	tmpFile, err := os.CreateTemp("", "test_file_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpFilePath)

	runSimpleConstraintTests(t, dirConstraint{}, []simpleTestCase{
		// Valid - directory exists
		{"valid existing directory", tmpDir, false},
		{"valid temp directory", os.TempDir(), false},
		// Empty string - should be skipped
		{"empty string", "", false},
		// Invalid - directory does not exist
		{"invalid nonexistent directory", "/nonexistent/path/to/dir", true},
		{"invalid nonexistent in temp", filepath.Join(os.TempDir(), "nonexistent_dir_12345"), true},
		// Invalid - path is a file, not a directory
		{"invalid file not directory", tmpFilePath, true},
		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},
		// Invalid types
		{"invalid type - int", 123, true},
	})
}
