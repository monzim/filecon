# FileCon: File Concatenator CLI

**FileCon** is a simple and efficient CLI tool that allows you to concatenate specific file types from a directory into a single output file. It provides both a wizard-style interface and command-line options for flexibility. You can also remove tabs and extra spaces from the content during concatenation.

## Features

- Wizard-style interface for ease of use
- Command-line options for advanced users
- Ability to concatenate files based on their extension
- Option to remove tabs and extra spaces from the concatenated files
- Ignore specific file extensions during concatenation
- Ignore specific folders during search
- Support for `.fileconignore` file to define patterns to ignore
- Support for global environment variable `FILECON_EXTIGNORE`
- Detailed summary of concatenation results
- Automatic "filecon\_" prefix for output files

## Usage

```bash
filecon [flags]
```

### Flags

| Flag                   | Description                                        | Default                          |
| ---------------------- | -------------------------------------------------- | -------------------------------- |
| `-d, --dir string`     | Directory to search for files                      | Current directory                |
| `-e, --ext string`     | File extension to search for (with or without dot) | None                             |
| `-o, --out string`     | Output file name                                   | `filecon_output_<timestamp>.txt` |
| `-r, --remove-spaces`  | Remove all tabs and extra spaces from the content  | `false`                          |
| `-i, --ignore-ext`     | File extensions to ignore (comma-separated)        | None                             |
| `-f, --ignore-folders` | Folders to ignore during search (comma-separated)  | None                             |
| `-h, --help`           | Help for `filecon`                                 |                                  |

### Examples

1. **Run the interactive wizard**:

   ```bash
   filecon
   ```

2. **Concatenate all `.go` files in the current directory**:

   ```bash
   filecon --dir=. --ext=go --out=output.txt
   ```

3. **Concatenate all `.js` files in `/path/to/dir`, removing extra spaces**:

   ```bash
   filecon --dir=/path/to/dir --ext=.js --out=result.js --remove-spaces
   ```

4. **Concatenate all `.py` files in the current directory, ignoring `.pyc` files and the `venv` folder**:

   ```bash
   filecon --dir=. --ext=py --ignore-ext=pyc --ignore-folders=venv,__pycache__
   ```

5. **Using `.fileconignore` file**:
   Create a file named `.fileconignore` in your target directory with patterns to ignore:

   ```
   # Comments are supported
   *.tmp
   *.bak
   node_modules/
   .git/
   ```

6. **Using environment variable**:
   ```bash
   export FILECON_EXTIGNORE=tmp,bak,test.js
   filecon --dir=. --ext=js
   ```

> **Note**: If you don't provide all required flags (`dir`, `ext`, `out`), the interactive wizard will start by default.

## Ignore Patterns

### Using `.fileconignore` File

You can create a `.fileconignore` file in the target directory to specify files and folders to ignore. The syntax is similar to `.gitignore`:

- Use `*.ext` or `.ext` to ignore file extensions
- Use folder names with or without trailing slashes to ignore directories
- Use `#` for comments

### Using Environment Variable

Set the `FILECON_EXTIGNORE` environment variable to specify file extensions to ignore globally:

```bash
export FILECON_EXTIGNORE=tmp,bak,log
```

## Output Format

All output files are automatically prefixed with "filecon\_" if not already present. The output file will contain:

- A signature with the FileCon version and generation timestamp
- Information about any ignore patterns used
- Each file's content preceded by its path and a separator

## Installation

### For macOS

```bash
# Note: Change the architecture to arm64 for Apple Silicon-based Macs. Check the 'build' folder for your architecture.
sudo rm -f /usr/local/bin/filecon
sudo curl -L -o /usr/local/bin/filecon https://github.com/monzim/filecon/raw/main/build/0.0.2/filecon-macos-arm64
sudo chmod +x /usr/local/bin/filecon
```

### For Linux

```bash
# Note: Change the architecture to arm64 for ARM-based Linux systems. Check the 'build' folder for your architecture.
sudo rm -f /usr/local/bin/filecon
sudo curl -L -o /usr/local/bin/filecon https://github.com/monzim/filecon/raw/main/build/0.0.2/filecon-linux
sudo chmod +x /usr/local/bin/filecon
```

### For Windows

1. Download the `.exe` file for your architecture from the [latest releases](https://github.com/monzim/filecon/releases).
2. Add the location of the downloaded `.exe` file to your system's PATH, or run it directly from the command prompt.

## Building From Source

If you prefer to build from source, you need Go installed on your system.

1. Clone the repository:

   ```bash
   git clone https://github.com/monzim/filecon.git
   cd filecon
   ```

2. Build the application:

   ```bash
   go build -o filecon
   ```

3. Move the `filecon` binary to your PATH:
   ```bash
   sudo mv filecon /usr/local/bin/
   ```

## Contribution

Feel free to open issues and contribute to the project by creating pull requests.

## License

This project is licensed under the [MIT License](LICENSE).

---

> Developed by [Azraf Al Monzim](https://monzim.com)
