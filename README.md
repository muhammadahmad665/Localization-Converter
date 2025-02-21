
# Localization Converter

`localization_converter` is a Go-based command-line tool for converting between `.xcstrings` JSON files (used in Xcode String Catalogs for iOS localization) and CSV files. It supports bidirectional conversion, allowing you to:
- Export string keys and translations from `.xcstrings` to a CSV file for editing.
- Import edited CSV files back into the `.xcstrings` format for use in Xcode.

This tool is particularly useful for managing translations in iOS projects, providing a simple way to work with localization data in a spreadsheet format.

## Features

- **JSON to CSV Conversion**: Extracts string keys and translations from an `.xcstrings` file and writes them to a CSV file, with the source language (e.g., "en") as the default column.
- **CSV to JSON Conversion**: Reads a CSV file with keys and translations, reconstructing a properly formatted `.xcstrings` file compatible with Xcode.
- **Error Handling**: Includes robust checks for file existence, format validity, and duplicate keys.
- **Command-Line Interface**: Uses flags for easy operation in the terminal.

## Installation

### Prerequisites
- Go 1.16 or later installed on your system. [Download Go](https://golang.org/dl/)

### Setup
1. Clone or download this repository:
   ```bash
   git clone https://github.com/muhammadahmad665/Localization-Converter.git
   cd localization_converter
   ```
2. No external dependencies are required beyond the Go standard library, so you can run the program directly.

## Usage

The program supports two modes: `json2csv` and `csv2json`. Use the `-mode`, `-input`, and `-output` flags to specify the operation and file paths.

### Commands

#### 1. Convert `.xcstrings` to CSV
```bash
go run localization_converter.go -mode json2csv -input Localizable.xcstrings -output translations.csv
```
- **Input**: An `.xcstrings` JSON file (e.g., `Localizable.xcstrings`).
- **Output**: A CSV file (e.g., `translations.csv`) with keys in the first column and translations in subsequent columns.

#### 2. Convert CSV to `.xcstrings`
```bash
go run localization_converter.go -mode csv2json -input translations.csv -output Localizable.xcstrings
```
- **Input**: A CSV file (e.g., `translations.csv`).
- **Output**: An `.xcstrings` JSON file (e.g., `Localizable.xcstrings`) compatible with Xcode.

### Example Workflow

#### Starting with `.xcstrings`
Create a file named `Localizable.xcstrings`:
```json
{
  "sourceLanguage": "en",
  "strings": {
    "greeting": {
      "extractionState": "manual"
    },
    "farewell": {
      "extractionState": "manual"
    }
  },
  "version": "1.0"
}
```

Run the conversion to CSV:
```bash
go run localization_converter.go -mode json2csv -input Localizable.xcstrings -output translations.csv
```

**Output** (`translations.csv`):
```
,en
greeting,greeting
farewell,farewell
```

#### Editing Translations
Edit `translations.csv` to add French translations:
```
,en,fr
greeting,Hello,Bonjour
farewell,Goodbye,Au revoir
```

Convert back to `.xcstrings`:
```bash
go run localization_converter.go -mode csv2json -input translations.csv -output Localizable.xcstrings
```

**Output** (`Localizable.xcstrings`):
```json
{
  "sourceLanguage": "en",
  "strings": {
    "farewell": {
      "extractionState": "manual",
      "localizations": {
        "en": {
          "stringUnit": {
            "state": "translated",
            "value": "Goodbye"
          }
        },
        "fr": {
          "stringUnit": {
            "state": "translated",
            "value": "Au revoir"
          }
        }
      }
    },
    "greeting": {
      "extractionState": "manual",
      "localizations": {
        "en": {
          "stringUnit": {
            "state": "translated",
            "value": "Hello"
          }
        },
        "fr": {
          "stringUnit": {
            "state": "translated",
            "value": "Bonjour"
          }
        }
      }
    }
  },
  "version": "1.0"
}
```

## File Format Details

### `.xcstrings` JSON
- Follows the Xcode String Catalog format.
- Contains `sourceLanguage`, `strings`, and `version`.
- Each string entry has an `extractionState` and optional `localizations` with translations.

### CSV
- First row: Empty cell followed by language codes (e.g., `,en,fr`).
- Subsequent rows: Key in the first column, followed by translations for each language.

## Error Handling
- Missing or invalid files: Returns an error message (e.g., "error opening JSON file").
- Duplicate keys in CSV: Prevents invalid JSON generation.
- Empty or malformed input: Provides clear error messages.

## Building the Executable
To create a standalone binary:
```bash
go build -o localization_converter localization_converter.go
```
Then use it like:
```bash
./localization_converter -mode json2csv -input Localizable.xcstrings -output translations.csv
```

## License

This project is licensed under the MIT License. See below for details:

```
MIT License

Copyright (c) 2025 Muhammad Ahmad

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## Contributing
Feel free to submit issues or pull requests to improve this tool. Suggestions for additional features (e.g., support for more formats) are welcome!

## Contact
For questions or support, please open an issue on this repository.


---