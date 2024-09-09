# FileCon: File Concatenator CLI

**FileCon** is a simple and efficient CLI tool that allows you to concatenate specific file types from a directory into a single output file. It provides both a wizard-style interface and command-line options for flexibility. You can also remove tabs and extra spaces from the content during concatenation.

## Features

- Wizard-style interface for ease of use.
- Command-line options for advanced users.
- Ability to concatenate files based on their extension.
- Option to remove tabs and extra spaces from the concatenated files.

## Usage

```bash
filecon [flags]
```

### Flags

| Flag                  | Description                                        | Default           |
| --------------------- | -------------------------------------------------- | ----------------- |
| `-d, --dir string`    | Directory to search for files                      | Current directory |
| `-e, --ext string`    | File extension to search for (e.g., .go, .js, .py) | None              |
| `-o, --out string`    | Output file name                                   | `output.txt`      |
| `-r, --remove-spaces` | Remove all tabs and extra spaces from the content  | `false`           |
| `-h, --help`          | Help for `filecon`                                 |                   |

### Examples

1. **Run the interactive wizard**:

   ```bash
   filecon
   ```

2. **Concatenate all `.go` files in the current directory into `output.txt`**:

   ```bash
   filecon --dir=. --ext=.go --out=output.txt
   ```

3. **Concatenate all `.js` files in `/path/to/dir` into `result.js`, removing extra spaces**:
   ```bash
   filecon --dir=/path/to/dir --ext=.js --out=result.js --remove-spaces
   ```

> **Note**: If you don't provide all required flags (`dir`, `ext`, `out`), the interactive wizard will start by default.

## Installation

### For macOS

```bash
# Note: Change the architecture to arm64 for Apple Silicon-based Macs. Check the 'build' folder for your architecture.


sudo curl -L https://github.com/monzim/filecon/raw/main/build/filecon-macos-amd64 -o /usr/local/bin/filecon
sudo chmod +x /usr/local/bin/filecon
```

### For Linux

```bash
# Note: Change the architecture to arm64 for ARM-based Linux systems. Check the 'build' folder for your architecture.

sudo curl -L https://github.com/monzim/filecon/raw/main/build/filecon-linux -o /usr/local/bin/filecon
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

> Developed by [Azraf Al Monzim](https://github.com/monzim)
