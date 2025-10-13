package dirsearch

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.SearchPattern != "" {
		t.Errorf("expected empty SearchPattern, got %q", opts.SearchPattern)
	}

	if opts.StartDir != "." {
		t.Errorf("expected StartDir to be '.', got %q", opts.StartDir)
	}

	if opts.CaseSensitive {
		t.Error("expected CaseSensitive to be false")
	}

	if len(opts.IgnorePatterns) != 1 || opts.IgnorePatterns[0] != "node_modules" {
		t.Errorf("expected IgnorePatterns to be ['node_modules'], got %v", opts.IgnorePatterns)
	}
}

func TestNewDirSearch(t *testing.T) {
	ds := NewDirSearch()

	if ds == nil {
		t.Fatal("expected DirSearch instance, got nil")
	}

	if ds.Options == nil {
		t.Fatal("expected Options to be initialized, got nil")
	}

	// Verify it uses default options
	if ds.Options.StartDir != "." {
		t.Errorf("expected default StartDir '.', got %q", ds.Options.StartDir)
	}
}

func TestSearch_EmptyDirectory(t *testing.T) {
	// Create a temporary empty directory
	tempDir, err := os.MkdirTemp("", "dirsearch-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	opts := &Options{
		SearchPattern:  "",
		StartDir:       tempDir,
		CaseSensitive:  false,
		IgnorePatterns: []string{},
	}

	result := Search(opts)

	if result.Error != nil {
		t.Errorf("unexpected error: %v", result.Error)
	}

	if len(result.Directories) != 0 {
		t.Errorf("expected 0 directories, got %d", len(result.Directories))
	}
}

func TestSearch_WithSubdirectories(t *testing.T) {
	// Create a temporary directory with some subdirectories
	tempDir, err := os.MkdirTemp("", "dirsearch-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test subdirectories
	testDirs := []string{"testdir1", "testdir2", "anotherdir"}
	for _, dir := range testDirs {
		if err := os.Mkdir(filepath.Join(tempDir, dir), 0755); err != nil {
			t.Fatalf("failed to create test dir %s: %v", dir, err)
		}
	}

	opts := &Options{
		SearchPattern:  "",
		StartDir:       tempDir,
		CaseSensitive:  false,
		IgnorePatterns: []string{},
	}

	result := Search(opts)

	if result.Error != nil {
		t.Errorf("unexpected error: %v", result.Error)
	}

	if len(result.Directories) != len(testDirs) {
		t.Errorf("expected %d directories, got %d", len(testDirs), len(result.Directories))
	}

	// Verify all expected directories are in the result
	foundDirs := make(map[string]bool)
	for _, dir := range result.Directories {
		foundDirs[dir] = true
	}

	for _, expectedDir := range testDirs {
		if !foundDirs[expectedDir] {
			t.Errorf("expected to find directory %q in results", expectedDir)
		}
	}
}

func TestSearch_CaseSensitive(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dirsearch-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create directories with different cases
	testDirs := []string{"TestDir", "testdir", "TESTDIR"}
	for _, dir := range testDirs {
		if err := os.Mkdir(filepath.Join(tempDir, dir), 0755); err != nil {
			t.Fatalf("failed to create test dir %s: %v", dir, err)
		}
	}

	t.Run("case-insensitive", func(t *testing.T) {
		opts := &Options{
			SearchPattern:  "test",
			StartDir:       tempDir,
			CaseSensitive:  false,
			IgnorePatterns: []string{},
		}

		result := Search(opts)

		if result.Error != nil {
			t.Errorf("unexpected error: %v", result.Error)
		}

		// Should match all three directories
		if len(result.Directories) != 3 {
			t.Errorf("expected 3 directories with case-insensitive search, got %d", len(result.Directories))
		}
	})

	t.Run("case-sensitive", func(t *testing.T) {
		opts := &Options{
			SearchPattern:  "Test",
			StartDir:       tempDir,
			CaseSensitive:  true,
			IgnorePatterns: []string{},
		}

		result := Search(opts)

		if result.Error != nil {
			t.Errorf("unexpected error: %v", result.Error)
		}

		// Should only match "TestDir"
		if len(result.Directories) != 1 {
			t.Errorf("expected 1 directory with case-sensitive search, got %d", len(result.Directories))
		}

		if len(result.Directories) > 0 && result.Directories[0] != "TestDir" {
			t.Errorf("expected to find 'TestDir', got %q", result.Directories[0])
		}
	})
}

func TestSearch_IgnorePatterns(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dirsearch-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create directories including one that should be ignored
	testDirs := []string{"gooddir", "node_modules", "anotherdir"}
	for _, dir := range testDirs {
		if err := os.Mkdir(filepath.Join(tempDir, dir), 0755); err != nil {
			t.Fatalf("failed to create test dir %s: %v", dir, err)
		}
	}

	opts := &Options{
		SearchPattern:  "",
		StartDir:       tempDir,
		CaseSensitive:  false,
		IgnorePatterns: []string{"node_modules"},
	}

	result := Search(opts)

	if result.Error != nil {
		t.Errorf("unexpected error: %v", result.Error)
	}

	// Should find 2 directories (excluding node_modules)
	if len(result.Directories) != 2 {
		t.Errorf("expected 2 directories, got %d", len(result.Directories))
	}

	// Verify node_modules is not in results
	for _, dir := range result.Directories {
		if dir == "node_modules" {
			t.Error("node_modules should have been ignored")
		}
	}
}

func TestSearch_GitDirectoriesIgnored(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dirsearch-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create directories including .git
	testDirs := []string{"normaldir", ".git", ".github"}
	for _, dir := range testDirs {
		if err := os.Mkdir(filepath.Join(tempDir, dir), 0755); err != nil {
			t.Fatalf("failed to create test dir %s: %v", dir, err)
		}
	}

	opts := &Options{
		SearchPattern:  "",
		StartDir:       tempDir,
		CaseSensitive:  false,
		IgnorePatterns: []string{},
	}

	result := Search(opts)

	if result.Error != nil {
		t.Errorf("unexpected error: %v", result.Error)
	}

	// Verify .git and .github are not in results (git directories are always filtered)
	for _, dir := range result.Directories {
		if dir == ".git" || dir == ".github" {
			t.Errorf("git directory %q should have been ignored", dir)
		}
	}
}

func TestScanDirs(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dirsearch-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test subdirectories
	if err := os.Mkdir(filepath.Join(tempDir, "subdir"), 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	ds := NewDirSearch()
	result := ds.ScanDirs(tempDir)

	if result.Error != nil {
		t.Errorf("unexpected error: %v", result.Error)
	}

	if len(result.Directories) != 1 {
		t.Errorf("expected 1 directory, got %d", len(result.Directories))
	}

	// Verify the StartDir was updated
	if ds.Options.StartDir != tempDir {
		t.Errorf("expected StartDir to be updated to %q, got %q", tempDir, ds.Options.StartDir)
	}
}

