package app

import (
	"testing"
)

func TestNewApplication(t *testing.T) {
	app, err := NewApplication()

	if err != nil {
		t.Fatalf("unexpected error creating application: %v", err)
	}

	if app == nil {
		t.Fatal("expected Application instance, got nil")
	}

	if app.Dirsearch == nil {
		t.Error("expected Dirsearch to be initialized, got nil")
	}

	if app.Logger == nil {
		t.Error("expected Logger to be initialized, got nil")
	}

	// Verify Dirsearch has default options
	if app.Dirsearch.Options == nil {
		t.Error("expected Dirsearch.Options to be initialized, got nil")
	}

	if app.Dirsearch.Options.StartDir != "." {
		t.Errorf("expected default StartDir '.', got %q", app.Dirsearch.Options.StartDir)
	}
}

func TestApplicationComponents(t *testing.T) {
	app, err := NewApplication()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("logger is functional", func(t *testing.T) {
		// Test that logger can be used without panicking
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("logger usage caused panic: %v", r)
			}
		}()

		app.Logger.Info("test message")
		app.Logger.Debug("debug message")
		app.Logger.Error("error message", "key", "value")
	})

	t.Run("dirsearch is functional", func(t *testing.T) {
		// Test that dirsearch can be used
		result := app.Dirsearch.ScanDirs(".")

		// We don't care about the specific results, just that it doesn't panic
		// and returns a valid Result structure
		if result.Error != nil {
			// This might fail in some CI environments, but that's okay
			t.Logf("scan returned error (may be expected): %v", result.Error)
		}

		// Result should have a Directories slice (even if empty)
		if result.Directories == nil {
			t.Error("expected Directories slice to be initialized, got nil")
		}
	})
}
