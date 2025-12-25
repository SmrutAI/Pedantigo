package constraints

import (
	"os"
	"path/filepath"
	"testing"
)

// TestImageConstraint tests imageConstraint.Validate() for valid image files.
// This constraint checks the actual filesystem and validates image format by reading magic bytes.
func TestImageConstraint(t *testing.T) {
	// Create temp directory for test files
	tmpDir, err := os.MkdirTemp("", "test_image_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// ============================================================================
	// Create valid image files with proper magic bytes
	// ============================================================================

	// JPEG magic bytes: 0xFF 0xD8 0xFF
	jpegData := []byte{
		0xFF, 0xD8, 0xFF, 0xE0, // JFIF header start
		0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, // JFIF identifier
		0x01, 0x01, // version 1.1
		0x00, 0x00, 0x00, 0x01, 0x00, 0x01, // density info
		0x00, 0x00, // thumbnail dimensions
		0xFF, 0xD9, // JPEG end marker
	}
	jpegFile := createTempFile(t, tmpDir, "test.jpg", jpegData)

	// PNG magic bytes: 0x89 PNG\r\n 0x1A\n
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, 'I', 'H', 'D', 'R', // IHDR chunk
		0x00, 0x00, 0x00, 0x01, // width: 1
		0x00, 0x00, 0x00, 0x01, // height: 1
		0x08, 0x02, 0x00, 0x00, 0x00, // bit depth, color type, etc.
		0x90, 0x77, 0x53, 0xDE, // IHDR CRC
		0x00, 0x00, 0x00, 0x00, 'I', 'E', 'N', 'D', // IEND chunk
		0xAE, 0x42, 0x60, 0x82, // IEND CRC
	}
	pngFile := createTempFile(t, tmpDir, "test.png", pngData)

	// GIF magic bytes: GIF87a or GIF89a
	gifData := []byte{
		'G', 'I', 'F', '8', '9', 'a', // GIF89a signature
		0x01, 0x00, // width: 1
		0x01, 0x00, // height: 1
		0x00, 0x00, 0x00, // color table and background
		0x3B, // GIF terminator
	}
	gifFile := createTempFile(t, tmpDir, "test.gif", gifData)

	// BMP magic bytes: BM
	bmpData := []byte{
		'B', 'M', // BMP signature
		0x46, 0x00, 0x00, 0x00, // file size
		0x00, 0x00, 0x00, 0x00, // reserved
		0x36, 0x00, 0x00, 0x00, // offset to pixel data
		0x28, 0x00, 0x00, 0x00, // DIB header size
		0x01, 0x00, 0x00, 0x00, // width: 1
		0x01, 0x00, 0x00, 0x00, // height: 1
		0x01, 0x00, // color planes
		0x18, 0x00, // bits per pixel
		// ... rest of minimal BMP structure
	}
	bmpFile := createTempFile(t, tmpDir, "test.bmp", bmpData)

	// WebP magic bytes: RIFF....WEBP
	webpData := []byte{
		'R', 'I', 'F', 'F', // RIFF
		0x1A, 0x00, 0x00, 0x00, // file size
		'W', 'E', 'B', 'P', // WEBP
		'V', 'P', '8', ' ', // VP8 chunk
		0x0E, 0x00, 0x00, 0x00, // chunk size
		// ... minimal VP8 data
	}
	webpFile := createTempFile(t, tmpDir, "test.webp", webpData)

	// ============================================================================
	// Create invalid files
	// ============================================================================

	// Plain text file
	textData := []byte("This is a plain text file, not an image.")
	textFile := createTempFile(t, tmpDir, "test.txt", textData)

	// JSON file
	jsonData := []byte(`{"type": "not an image", "data": "json content"}`)
	jsonFile := createTempFile(t, tmpDir, "test.json", jsonData)

	// Empty file
	emptyFile := createTempFile(t, tmpDir, "empty.jpg", []byte{})

	// File with wrong extension (text file named .jpg)
	fakeJpegFile := createTempFile(t, tmpDir, "fake.jpg", []byte("not a real jpeg"))

	// Create a directory for testing
	testDir := filepath.Join(tmpDir, "testdir")
	err = os.Mkdir(testDir, 0o750)
	if err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// ============================================================================
	// Run validation tests
	// ============================================================================

	runSimpleConstraintTests(t, imageConstraint{}, []simpleTestCase{
		// Valid image files
		{"valid jpeg file", jpegFile, false},
		{"valid png file", pngFile, false},
		{"valid gif file", gifFile, false},
		{"valid bmp file", bmpFile, false},
		{"valid webp file", webpFile, false},

		// Empty string - should be skipped
		{"empty string", "", false},

		// Invalid - non-existent file
		{"invalid nonexistent file", "/nonexistent/path/to/image.jpg", true},
		{"invalid nonexistent in temp", filepath.Join(tmpDir, "nonexistent.jpg"), true},

		// Invalid - not image files
		{"invalid text file", textFile, true},
		{"invalid json file", jsonFile, true},
		{"invalid empty file", emptyFile, true},
		{"invalid fake jpeg", fakeJpegFile, true},

		// Invalid - directory, not a file
		{"invalid directory", testDir, true},

		// Nil pointer - should skip validation
		{"nil pointer", (*string)(nil), false},

		// Invalid types
		{"invalid type - int", 123, true},
		{"invalid type - bool", true, true},
		{"invalid type - float", 3.14, true},
	})
}

// TestImageConstraintWithDifferentFormats tests additional image format edge cases.
func TestImageConstraintWithDifferentFormats(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_image_formats_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name     string
		filename string
		data     []byte
		wantErr  bool
	}{
		{
			name:     "jpeg with JFIF",
			filename: "jfif.jpg",
			data:     []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00},
			wantErr:  false,
		},
		{
			name:     "jpeg with Exif",
			filename: "exif.jpg",
			data:     []byte{0xFF, 0xD8, 0xFF, 0xE1, 0x00, 0x16, 'E', 'x', 'i', 'f', 0x00, 0x00},
			wantErr:  false,
		},
		{
			name:     "png signature only",
			filename: "minimal.png",
			data:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			wantErr:  false,
		},
		{
			name:     "gif87a format",
			filename: "old.gif",
			data:     []byte{'G', 'I', 'F', '8', '7', 'a', 0x01, 0x00, 0x01, 0x00},
			wantErr:  false,
		},
		{
			name:     "gif89a format",
			filename: "new.gif",
			data:     []byte{'G', 'I', 'F', '8', '9', 'a', 0x01, 0x00, 0x01, 0x00},
			wantErr:  false,
		},
		{
			name:     "corrupted jpeg - missing markers",
			filename: "corrupt.jpg",
			data:     []byte{0xFF, 0xD8, 0x00, 0x00, 0x00}, // starts like JPEG but corrupted
			wantErr:  true,
		},
		{
			name:     "almost png - wrong signature",
			filename: "fake.png",
			data:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x00, 0x00}, // wrong last bytes
			wantErr:  true,
		},
		{
			name:     "html file",
			filename: "page.html",
			data:     []byte("<!DOCTYPE html><html><body><img src=\"test.jpg\"></body></html>"),
			wantErr:  true,
		},
		{
			name:     "xml file",
			filename: "data.xml",
			data:     []byte("<?xml version=\"1.0\"?><root><image>data</image></root>"),
			wantErr:  true,
		},
		{
			name:     "binary data - not image",
			filename: "binary.dat",
			data:     []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09},
			wantErr:  true,
		},
		{
			name:     "pdf file start",
			filename: "doc.pdf",
			data:     []byte{'%', 'P', 'D', 'F', '-', '1', '.', '4'},
			wantErr:  true,
		},
		{
			name:     "zip file start",
			filename: "archive.zip",
			data:     []byte{'P', 'K', 0x03, 0x04},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := createTempFile(t, tmpDir, tt.filename, tt.data)

			c := imageConstraint{}
			err := c.Validate(filePath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for %s, got nil", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error for %s, got: %v", tt.name, err)
				}
			}
		})
	}
}

// TestImageConstraintEdgeCases tests edge cases and error conditions.
func TestImageConstraintEdgeCases(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_image_edge_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
		wantErr  bool
	}{
		{
			name: "file too small - less than magic bytes",
			testFunc: func(t *testing.T) {
				smallFile := createTempFile(t, tmpDir, "small.jpg", []byte{0xFF, 0xD8})

				c := imageConstraint{}
				err := c.Validate(smallFile)
				if err == nil {
					t.Error("expected error for file too small to contain valid image data")
				}
			},
			wantErr: true,
		},
		{
			name: "symbolic link to valid image",
			testFunc: func(t *testing.T) {
				// Create a valid JPEG
				jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00}
				jpegFile := createTempFile(t, tmpDir, "original.jpg", jpegData)

				// Create a symlink to it
				symlinkPath := filepath.Join(tmpDir, "link.jpg")
				err := os.Symlink(jpegFile, symlinkPath)
				if err != nil {
					t.Skipf("skipping symlink test: %v", err)
				}

				c := imageConstraint{}
				err = c.Validate(symlinkPath)
				if err != nil {
					t.Errorf("expected no error for symlink to valid image, got: %v", err)
				}
			},
			wantErr: false,
		},
		{
			name: "file with no extension",
			testFunc: func(t *testing.T) {
				// Valid PNG data but no extension
				pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
				noExtFile := createTempFile(t, tmpDir, "imagefile", pngData)

				c := imageConstraint{}
				err := c.Validate(noExtFile)
				// Should still validate based on content, not extension
				if err != nil {
					t.Errorf("expected no error for valid PNG without extension, got: %v", err)
				}
			},
			wantErr: false,
		},
		{
			name: "file with wrong extension but valid content",
			testFunc: func(t *testing.T) {
				// PNG data with .txt extension
				pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
				wrongExtFile := createTempFile(t, tmpDir, "image.txt", pngData)

				c := imageConstraint{}
				err := c.Validate(wrongExtFile)
				// Should validate based on content, not extension
				if err != nil {
					t.Errorf("expected no error for valid PNG with wrong extension, got: %v", err)
				}
			},
			wantErr: false,
		},
		{
			name: "very large valid image file",
			testFunc: func(t *testing.T) {
				// Create a larger JPEG-like file
				largeData := make([]byte, 1024*100) // 100KB
				// Add JPEG magic bytes at start
				copy(largeData, []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00})
				// Add JPEG end marker
				largeData[len(largeData)-2] = 0xFF
				largeData[len(largeData)-1] = 0xD9

				largeFile := createTempFile(t, tmpDir, "large.jpg", largeData)

				c := imageConstraint{}
				err := c.Validate(largeFile)
				if err != nil {
					t.Errorf("expected no error for large valid JPEG, got: %v", err)
				}
			},
			wantErr: false,
		},
		{
			name: "file with special characters in name",
			testFunc: func(t *testing.T) {
				pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
				specialFile := createTempFile(t, tmpDir, "image with spaces & special-chars.png", pngData)

				c := imageConstraint{}
				err := c.Validate(specialFile)
				if err != nil {
					t.Errorf("expected no error for file with special chars, got: %v", err)
				}
			},
			wantErr: false,
		},
		{
			name: "file with unicode name",
			testFunc: func(t *testing.T) {
				gifData := []byte{'G', 'I', 'F', '8', '9', 'a', 0x01, 0x00, 0x01, 0x00}
				unicodeFile := createTempFile(t, tmpDir, "图片-image-画像.gif", gifData)

				c := imageConstraint{}
				err := c.Validate(unicodeFile)
				if err != nil {
					t.Errorf("expected no error for file with unicode name, got: %v", err)
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

// TestImageConstraintPermissions tests file permission edge cases.
func TestImageConstraintPermissions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_image_perms_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a valid JPEG
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00}
	jpegFile := createTempFile(t, tmpDir, "test.jpg", jpegData)

	// Try to make it unreadable (may not work on all systems)
	err = os.Chmod(jpegFile, 0o000)
	if err != nil {
		t.Skipf("skipping permission test: %v", err)
	}
	defer func() { _ = os.Chmod(jpegFile, 0o600) }() // restore permissions for cleanup

	c := imageConstraint{}
	err = c.Validate(jpegFile)
	if err == nil {
		// Some systems might allow reading despite permissions
		t.Logf("note: file read succeeded despite 0000 permissions (system-dependent)")
	}
}

// createTempFile is a helper function to create temporary files with specific content.
func createTempFile(t *testing.T, dir, name string, data []byte) string {
	t.Helper()

	filePath := filepath.Join(dir, name)
	f, err := os.Create(filePath) //nolint:gosec // G304: Test helper creating files in temp dir
	if err != nil {
		t.Fatalf("failed to create temp file %s: %v", name, err)
	}

	if len(data) > 0 {
		_, err = f.Write(data)
		if err != nil {
			f.Close()
			t.Fatalf("failed to write to temp file %s: %v", name, err)
		}
	}

	err = f.Close()
	if err != nil {
		t.Fatalf("failed to close temp file %s: %v", name, err)
	}

	return filePath
}
