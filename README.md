# gk-rename

`gk-rename` is a CLI tool for cloning and renaming a template project. It allows you to create a new project based on an existing template, while renaming specific project name within the project files and directories.

This tool is designed to simplify the process of renaming a project, handling folder names, file contents, and related configuration files, ensuring a consistent and error-free renaming process across the entire project.

## Features

- Clone a template project and rename the project name.
- Automatically update all references to the old project name in the files.
- Rename files and directories to reflect the new project name.
- Skip unnecessary directories like `node_modules`, `bin`, and `obj`.
- Handle renaming and content updating for key project files such as `build-and-test.yml`.

## Installation

To install `gk-rename`, you can download the binary or build it from the source.

### Build from source

```bash
git clone https://github.com/tguankheng016/gk-rename.git
cd gk-rename
go build -o gk-rename
```

## Usage

The `gk-rename` tool allows you to rename a template project by specifying the old and new project names. It updates references to the old project name in files and directories, making it easier to rename an entire project without manual intervention.

### Command Syntax

```bash
gk-rename --key <old_project_name> --value <new_project_name>
```
