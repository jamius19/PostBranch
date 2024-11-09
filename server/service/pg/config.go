package pg

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// CopyPostgresConfig copies the postgresql.conf file and all included files/directories
// to the new data cluster, updating paths in include directives accordingly.
func CopyPostgresConfig(srcConfPath, newDataClusterPath string) error {
	processedFiles := make(map[string]string)
	return processConfigFile(srcConfPath, newDataClusterPath, processedFiles)
}

func processConfigFile(srcFilePath, newDataClusterPath string, processedFiles map[string]string) error {
	// Check if the file has already been processed
	if _, ok := processedFiles[srcFilePath]; ok {
		return nil
	}

	// Determine the destination file path
	var destFilePath string
	if filepath.Base(srcFilePath) == "postgresql.conf" {
		destFilePath = filepath.Join(newDataClusterPath, "postgresql.conf")
	} else {
		destFilePath = filepath.Join(newDataClusterPath, "conf.d", filepath.Base(srcFilePath))
	}

	processedFiles[srcFilePath] = destFilePath

	// Read the source file
	content, err := os.ReadFile(srcFilePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", srcFilePath, err)
	}

	// Prepare to write the updated content
	var updatedContent []string

	// Split content into lines
	lines := strings.Split(string(content), "\n")

	srcDir := filepath.Dir(srcFilePath)

	for _, line := range lines {
		directive, includePath, ok := parseIncludeDirective(line)
		if ok {
			// Resolve the include path
			var resolvedPath string
			if filepath.IsAbs(includePath) {
				resolvedPath = includePath
			} else {
				resolvedPath = filepath.Join(srcDir, includePath)
			}

			var newIncludePath string
			if directive == "include" || directive == "include_if_exists" {
				// Process the included file recursively
				err = processConfigFile(resolvedPath, newDataClusterPath, processedFiles)
				if err != nil {
					return err
				}

				// Get the destination path of the included file
				destIncludePath := processedFiles[resolvedPath]

				// Determine the new include path relative to the newDataClusterPath
				newIncludePath, err = filepath.Rel(newDataClusterPath, destIncludePath)
				if err != nil {
					return fmt.Errorf("failed to get relative path: %v", err)
				}

			} else if directive == "include_dir" {
				// Copy the directory to conf.d
				dirName := filepath.Base(resolvedPath)
				destIncludeDir := filepath.Join(newDataClusterPath, "conf.d", dirName)

				// Ensure the conf.d directory exists
				err := os.MkdirAll(destIncludeDir, 0755)
				if err != nil {
					return fmt.Errorf("failed to create directory %s: %v", destIncludeDir, err)
				}

				// Copy the directory
				err = copyDir(resolvedPath, destIncludeDir)
				if err != nil {
					return fmt.Errorf("failed to copy directory %s to %s: %v", resolvedPath, destIncludeDir, err)
				}

				// Process each file in the directory recursively
				files, err := os.ReadDir(resolvedPath)
				if err != nil {
					return fmt.Errorf("failed to read directory %s: %v", resolvedPath, err)
				}
				for _, file := range files {
					if !file.IsDir() {
						srcFilePath := filepath.Join(resolvedPath, file.Name())
						err = processConfigFile(srcFilePath, newDataClusterPath, processedFiles)
						if err != nil {
							return err
						}
					}
				}

				// Determine the new include_dir path relative to the newDataClusterPath
				newIncludePath, err = filepath.Rel(newDataClusterPath, destIncludeDir)
				if err != nil {
					return fmt.Errorf("failed to get relative path: %v", err)
				}

			} else {
				return fmt.Errorf("unknown include directive: %s", directive)
			}

			// Update the include directive in the line
			// Keep the original quote style
			quoteChar := "'"
			if strings.Contains(line, "\""+includePath+"\"") {
				quoteChar = "\""
			}

			// Reconstruct the include directive
			newLine := fmt.Sprintf("%s %s%s%s", directive, quoteChar, newIncludePath, quoteChar)
			updatedContent = append(updatedContent, newLine)
		} else {
			// Not an include directive, keep the line as is
			updatedContent = append(updatedContent, line)
		}
	}

	// Write the updated content to destFilePath
	updatedData := strings.Join(updatedContent, "\n")
	err = os.WriteFile(destFilePath, []byte(updatedData), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %v", destFilePath, err)
	}

	return nil
}

func parseIncludeDirective(line string) (directive string, includePath string, ok bool) {
	// Remove comments
	if idx := strings.Index(line, "#"); idx != -1 {
		line = line[:idx]
	}
	if idx := strings.Index(line, "--"); idx != -1 {
		line = line[:idx]
	}
	line = strings.TrimSpace(line)

	if len(line) == 0 {
		return "", "", false
	}

	// Regex to match include directives
	includeRegex := regexp.MustCompile(`^(include|include_if_exists|include_dir)\s+'([^']+)'`)
	matches := includeRegex.FindStringSubmatch(line)
	if matches != nil && len(matches) == 3 {
		return matches[1], matches[2], true
	}

	// Try double quotes
	includeRegex = regexp.MustCompile(`^(include|include_if_exists|include_dir)\s+"([^"]+)"`)
	matches = includeRegex.FindStringSubmatch(line)
	if matches != nil && len(matches) == 3 {
		return matches[1], matches[2], true
	}

	// Try without quotes
	includeRegex = regexp.MustCompile(`^(include|include_if_exists|include_dir)\s+(\S+)`)
	matches = includeRegex.FindStringSubmatch(line)
	if matches != nil && len(matches) == 3 {
		return matches[1], matches[2], true
	}

	return "", "", false
}

func copyDir(src string, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			// Copy file
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(srcFile, destFile string) error {
	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(destFile)
	err := os.MkdirAll(destDir, 0755)
	if err != nil {
		return err
	}

	// Open source file
	src, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy data
	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	// Copy permissions
	info, err := os.Stat(srcFile)
	if err != nil {
		return err
	}
	err = os.Chmod(destFile, info.Mode())
	if err != nil {
		return err
	}

	return nil
}
