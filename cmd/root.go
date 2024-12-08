package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	oldCompanyName string
	newCompanyName string
	oldProjectName string
	newProjectName string
	newRoot        string
)

var skipDirs = map[string]struct{}{
	"node_modules": {},
	"bin":          {},
	"obj":          {},
}

var rootCmd = &cobra.Command{
	Use:   "gk-rename",
	Short: "gk-rename is a CLI tool for cloning and renaming an template project",
	Long:  "gk-rename is a CLI tool for cloning and renaming an template project",
	Run: func(cmd *cobra.Command, args []string) {
		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")

		if key == "" || value == "" {
			fmt.Println("Please provide both key and value")
			os.Exit(1)
		}

		if err := startProcess(key, value); err != nil {
			fmt.Fprintf(os.Stderr, "Oops. An error while executing Zero '%s'\n", err)
		}
	},
}

func Execute() {
	rootCmd.Flags().StringP("key", "k", "", "The key to be renamed.")
	rootCmd.Flags().StringP("value", "v", "", "The new name or value for the key.")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Oops. An error while executing gk-rename '%s'\n", err)
		os.Exit(1)
	}
}

func startProcess(key string, value string) error {
	oldCompanyName = "Cloned"
	oldProjectName = key
	newProjectName = value

	startTime := time.Now()

	// Set the root folder and directories to process
	newRoot = oldCompanyName + "_" + newProjectName
	fmt.Printf("Start renaming process...\n")

	// Create root folder for new project
	fs := afero.NewOsFs()
	err := fs.Mkdir(newRoot, 0755)
	if err != nil {
		return fmt.Errorf("error creating new root folder: %w", err)
	}

	// Copy required files and folders
	if err := copyItems(fs); err != nil {
		return err
	}

	// Perform the rename operation on specific folders
	renameFolderContent(fs, newRoot)

	// Output elapsed time
	elapsed := time.Since(startTime)
	fmt.Printf("Renaming process completed in %v\n", elapsed)

	return nil
}

// Function to copy items from one directory to another
func copyItems(fs afero.Fs) error {
	directoriesToCopy := []struct {
		src  string
		dest string
	}{
		{src: "./src", dest: filepath.Join(newRoot, "aspnet-core", "src")},
		{src: "./tests", dest: filepath.Join(newRoot, "aspnet-core", "tests")},
		{src: "./.vscode", dest: filepath.Join(newRoot, "aspnet-core", ".vscode")},
		{src: "./.github", dest: filepath.Join(newRoot, ".github")},
		{src: "CommerceMono.sln", dest: filepath.Join(newRoot, "aspnet-core", "CommerceMono.sln")},
		{src: ".gitignore", dest: filepath.Join(newRoot, "aspnet-core", ".gitignore")},
		{src: "add_migration.bat", dest: filepath.Join(newRoot, "aspnet-core", "add_migration.bat")},
		{src: "Makefile", dest: filepath.Join(newRoot, "aspnet-core", "Makefile")},
		{src: "run.bat", dest: filepath.Join(newRoot, "aspnet-core", "run.bat")},
	}

	for _, item := range directoriesToCopy {
		err := copyDirectoriesAndFiles(fs, item.src, item.dest)
		if err != nil {
			return fmt.Errorf("error copying from %s to %s: %w", item.src, item.dest, err)
		}
	}

	return nil
}

// CopyDirectory copies a source folder or file to the destination
func copyDirectoriesAndFiles(fs afero.Fs, src, dest string) error {
	// Check if the source exists
	exists, err := afero.Exists(fs, src)
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	info, err := fs.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		// Skip directories like node_modules, bin, and obj
		dirName := filepath.Base(src)
		if _, exists := skipDirs[dirName]; exists {
			fmt.Printf("Skipping directory: %s\n", src)
			return nil // Skip this directory
		}

		// Create directory at destination
		err := fs.MkdirAll(dest, 0755)
		if err != nil {
			return err
		}

		// Copy files recursively
		files, err := afero.ReadDir(fs, src)
		if err != nil {
			return err
		}

		for _, file := range files {
			srcFile := filepath.Join(src, file.Name())
			destFile := filepath.Join(dest, file.Name())
			err := copyDirectoriesAndFiles(fs, srcFile, destFile)
			if err != nil {
				return err
			}
		}
	} else {
		input, err := afero.ReadFile(fs, src)
		if err != nil {
			return err
		}
		err = afero.WriteFile(fs, dest, input, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

// Function to rename folder content and replace placeholders
func renameFolderContent(fs afero.Fs, targetFolder string) error {
	fmt.Printf("Start renaming content in folder: %s\n", targetFolder)

	// First, walk through files to rename them and update their content
	err := afero.Walk(fs, targetFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Read the file content
		fileContent, err := afero.ReadFile(fs, path)
		if err != nil {
			return err
		}

		// Replace company and project names in file content
		updatedContent := strings.Replace(string(fileContent), oldCompanyName, newCompanyName, -1)
		updatedContent = strings.Replace(updatedContent, oldProjectName, newProjectName, -1)

		// Write the updated content back to the file
		err = afero.WriteFile(fs, path, []byte(updatedContent), 0644)
		if err != nil {
			fmt.Printf("Error updating file content in %s: %v\n", path, err)
		} else {
			fmt.Printf("Updated file content: %s\n", path)
		}

		// Rename files and replace content if they match the placeholders
		if strings.Contains(info.Name(), oldCompanyName) || strings.Contains(info.Name(), oldProjectName) {
			// Rename the file itself
			newFileName := strings.Replace(info.Name(), oldCompanyName, newCompanyName, -1)
			newFileName = strings.Replace(newFileName, oldProjectName, newProjectName, -1)
			newFilePath := filepath.Join(filepath.Dir(path), newFileName)

			// Rename file only if the name has changed
			if newFilePath != path {
				err = fs.Rename(path, newFilePath)
				if err != nil {
					fmt.Printf("Error renaming file %s to %s: %v\n", path, newFilePath, err)
				} else {
					fmt.Printf("Renamed file: %s to %s\n", path, newFilePath)
				}
			}
		}

		err = processBuildTestYml(fs, path, info)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error renaming files in %s: %w", targetFolder, err)
	}

	// Next, walk through directories to rename them
	err = renameFolder(fs, targetFolder)
	if err != nil {
		return fmt.Errorf("error renaming folders in %s: %w", targetFolder, err)
	}

	return nil
}

// Function to rename folder recursively
func renameFolder(fs afero.Fs, targetFolder string) error {
	info, err := fs.Stat(targetFolder)
	if err != nil {
		return err
	}

	if info.IsDir() {
		files, err := afero.ReadDir(fs, targetFolder)
		if err != nil {
			return err
		}

		for _, file := range files {
			targetFolder := filepath.Join(targetFolder, file.Name())
			err := renameFolder(fs, targetFolder)
			if err != nil {
				return err
			}
		}

		// Rename folders if they match the placeholders
		if strings.Contains(info.Name(), oldCompanyName) || strings.Contains(info.Name(), oldProjectName) {
			fmt.Printf("Start renaming folder: %s\n", targetFolder)

			newPath := targetFolder

			if newCompanyName != "" {
				newPath = strings.Replace(newPath, oldCompanyName, newCompanyName, -1)
			}

			newPath = strings.Replace(newPath, oldProjectName, newProjectName, -1)
			err := fs.Rename(targetFolder, newPath)
			if err != nil {
				return fmt.Errorf("failed to rename directory %s: %v", targetFolder, err)
			}

			fmt.Println("Renamed directory:", targetFolder)
		}
	}

	return nil
}

// Function to process build.yml
func processBuildTestYml(fs afero.Fs, path string, info os.FileInfo) error {
	if info.Name() != "build-and-test.yml" {
		return nil
	}

	fmt.Printf("Start processing build-and-test.yml: %s\n", path)

	// Read the file content
	fileContent, err := afero.ReadFile(fs, path)
	if err != nil {
		return err
	}

	updatedContent := strings.Replace(string(fileContent), "\"src/", "\"aspnet-core/src/", -1)
	updatedContent = strings.Replace(updatedContent, "\"tests/", "\"aspnet-core/tests/", -1)
	updatedContent = strings.Replace(updatedContent, "\"./tests", "\"./aspnet-core/tests", -1)

	// Write the updated content back to the file
	err = afero.WriteFile(fs, path, []byte(updatedContent), 0644)
	if err != nil {
		fmt.Printf("Error updating file content in %s: %v\n", path, err)
	} else {
		fmt.Printf("Updated file content: %s\n", path)
	}

	return nil
}
