package cmd

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

// Test the `copyItems` function using the in-memory file system
func TestCopyItems(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Prepare a mock file structure
	fs.MkdirAll("src", 0755)
	fs.MkdirAll("tests", 0755)

	// Create dummy files to copy
	afero.WriteFile(fs, "src/testfile.txt", []byte("content"), 0644)
	afero.WriteFile(fs, "tests/testfile_test.txt", []byte("test content"), 0644)

	// Test copyItems function
	err := copyItems(fs)
	assert.NoError(t, err, "copyItems should not return an error")

	// Check if the files were copied to the correct locations
	_, err = fs.Stat(filepath.Join(newRoot, "aspnet-core", "src", "testfile.txt"))
	assert.NoError(t, err, "Source file should be copied")

	_, err = fs.Stat(filepath.Join(newRoot, "aspnet-core", "tests", "testfile_test.txt"))
	assert.NoError(t, err, "Test file should be copied")
}

// Test the `renameFolderContent` function for renaming files and replacing placeholders
func TestRenameFolderContent(t *testing.T) {
	oldProjectName = "CommerceMono"
	newProjectName = "HRMS"

	fs := afero.NewMemMapFs()

	// Create mock files with old project name
	fs.MkdirAll("src", 0755)
	afero.WriteFile(fs, "src/CommerceMono.sln", []byte("namespace CommerceMono"), 0644)

	// Run the rename function
	err := renameFolderContent(fs, "src")
	assert.NoError(t, err, "renameFolderContent should not return an error")

	// Check if file content was updated
	content, err := afero.ReadFile(fs, "src/HRMS.sln")
	assert.NoError(t, err, "File should be renamed and content should be replaced")
	assert.Contains(t, string(content), "namespace HRMS", "Content should be updated with new names")
}

// Test renaming directories using `renameFolder`
func TestRenameFolder(t *testing.T) {
	oldProjectName = "CommerceMono"
	newProjectName = "HRMS"

	fs := afero.NewMemMapFs()

	// Create a test directory
	fs.MkdirAll("src/CommerceMono.Application", 0755)
	fs.MkdirAll("src/CommerceMono.Api", 0755)

	// Run the rename function
	err := renameFolder(fs, "src")
	assert.NoError(t, err, "renameFolder should not return an error")

	// Check if the folder has been renamed
	_, err = fs.Stat("src/HRMS.Application")
	assert.NoError(t, err, "Appliction Folder should be renamed")

	_, err = fs.Stat("src/HRMS.Api")
	assert.NoError(t, err, "Api Folder should be renamed")
}

// Test file processing and replacing content in build.yml
func TestProcessBuildTestYml(t *testing.T) {
	fs := afero.NewMemMapFs()

	buildTestYmlContent := `
on:
  pull_request:
    branches:
      - main
  push:
    paths:
      - "src/**"
      - "tests/**"
  workflow_dispatch:

jobs:
  build-and-test-backend:
    name: Build And Test Backend
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Restore, Build And Run Unit Tests
        uses: tguankheng016/SharedActions/.github/actions/build-and-test-backend@main
        with:
          project-path: "./tests/CommerceMono.UnitTests"
          dotnet-version: "8.0.x"
      - name: Restore, Build And Run Integration Tests
        uses: tguankheng016/SharedActions/.github/actions/build-and-test-backend@main
        with:
          project-path: "./tests/CommerceMono.IntegrationTests"
          dotnet-version: "8.0.x"
`
	expectedYmlContent := `
on:
  pull_request:
    branches:
      - main
  push:
    paths:
      - "aspnet-core/src/**"
      - "aspnet-core/tests/**"
  workflow_dispatch:

jobs:
  build-and-test-backend:
    name: Build And Test Backend
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Restore, Build And Run Unit Tests
        uses: tguankheng016/SharedActions/.github/actions/build-and-test-backend@main
        with:
          project-path: "./aspnet-core/tests/CommerceMono.UnitTests"
          dotnet-version: "8.0.x"
      - name: Restore, Build And Run Integration Tests
        uses: tguankheng016/SharedActions/.github/actions/build-and-test-backend@main
        with:
          project-path: "./aspnet-core/tests/CommerceMono.IntegrationTests"
          dotnet-version: "8.0.x"
`

	// Create a mock build.yml file
	afero.WriteFile(fs, "build-and-test.yml", []byte(buildTestYmlContent), 0644)

	info, err := fs.Stat("build-and-test.yml")
	assert.NoError(t, err, "getting file info should not return an error")

	// Process the file
	err = processBuildTestYml(fs, "build-and-test.yml", info)
	assert.NoError(t, err, "processBuildTestYml should not return an error")

	// Read the updated file
	content, err := afero.ReadFile(fs, "build-and-test.yml")
	assert.NoError(t, err, "File should be read without error")
	assert.Equal(t, expectedYmlContent, string(content), "Content should be updated in build-and-test.yml")
}

// Test the `copyDirectory` function for copying files and skipping certain directories
func TestCopyDirectory(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create mock directories and files
	fs.MkdirAll("/node_modules", 0755)
	fs.MkdirAll("/bin", 0755)
	fs.MkdirAll("/obj", 0755)
	fs.MkdirAll("/src", 0755)

	// Copy directories (should skip node_modules, bin, and obj)
	err := copyDirectoriesAndFiles(fs, "/src", "/newpath/src")
	assert.NoError(t, err, "copyDirectory should not return an error")

	// Check if the "node_modules" directory was skipped
	_, err = fs.Stat("/newpath/src/node_modules")
	assert.Error(t, err, "node_modules directory should be skipped")

	// Check if the "src" directory was copied
	_, err = fs.Stat("/newpath/src")
	assert.NoError(t, err, "src directory should be copied")
}
