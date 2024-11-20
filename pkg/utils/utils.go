package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// MkdirOptions contains configuration options for directory creation
type MkdirOptions struct {
	// Create parent directories as needed
	Parents bool
	// Unix-style permission mode (default: 0755)
	Mode os.FileMode
	// Skip creation if directory exists (no error)
	ExistOk bool
	// Verbose output
	Verbose bool
}

// DefaultMkdirOptions returns the default options for Mkdir
func DefaultMkdirOptions() MkdirOptions {
	return MkdirOptions{
		Parents: false,
		Mode:    0755,
		ExistOk: false,
		Verbose: false,
	}
}

// Mkdir creates a directory with the specified path and options
func Mkdir(path string, opts MkdirOptions) error {
	// Clean the path to remove any unnecessary separators
	path = filepath.Clean(path)

	// Check if path is empty
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Check if the directory already exists
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			if opts.ExistOk {
				if opts.Verbose {
					fmt.Printf("Directory already exists: %s\n", path)
				}
				return nil
			}
			return fmt.Errorf("directory already exists: %s", path)
		}
		return fmt.Errorf("path exists but is not a directory: %s", path)
	}

	// If directory doesn't exist and parents option is true
	if opts.Parents {
		if err := createParentDirs(path, opts); err != nil {
			return err
		}
	}

	// Create the directory
	if err := os.Mkdir(path, opts.Mode); err != nil {
		// Handle race condition where directory might have been created
		// between our check and creation attempt
		if os.IsExist(err) && opts.ExistOk {
			if opts.Verbose {
				fmt.Printf("Directory already exists: %s\n", path)
			}
			return nil
		}
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if opts.Verbose {
		fmt.Printf("Created directory: %s\n", path)
	}

	return nil
}

// createParentDirs creates all parent directories if they don't exist
func createParentDirs(path string, opts MkdirOptions) error {
	// Get parent directory
	parent := filepath.Dir(path)

	// If we're at root or current directory, return
	if parent == "." || parent == "/" {
		return nil
	}

	// Check if parent exists
	_, err := os.Stat(parent)
	if err != nil {
		if os.IsNotExist(err) {
			// Recursively create parent
			if err := createParentDirs(parent, opts); err != nil {
				return err
			}
			// Create this level
			if err := os.Mkdir(parent, opts.Mode); err != nil {
				if !os.IsExist(err) {
					return fmt.Errorf("failed to create parent directory %s: %w", parent, err)
				}
			} else if opts.Verbose {
				fmt.Printf("Created parent directory: %s\n", parent)
			}
		} else {
			return fmt.Errorf("failed to check parent directory %s: %w", parent, err)
		}
	}

	return nil
}

// MkdirAll is a convenience function that creates a directory and all parents
func MkdirAll(path string, mode os.FileMode) error {
	opts := DefaultMkdirOptions()
	opts.Parents = true
	opts.Mode = mode
	opts.ExistOk = true
	return Mkdir(path, opts)
}
