package main

import (
	"encoding/csv"  // For reading and writing CSV files
	"encoding/json" // For parsing and generating JSON
	"flag"          // For command-line argument parsing
	"fmt"           // For formatted output
	"os"            // For file operations
	"sort"          // For sorting languages and keys
	"strings"       // For string manipulation
)

// Xcstrings represents the top-level structure of an .xcstrings JSON file as used in Xcode String Catalogs.
// It includes a source language, a map of string entries, and a version.
type Xcstrings struct {
	SourceLanguage string                 `json:"sourceLanguage"` // The default language (e.g., "en")
	Strings        map[string]StringEntry `json:"strings"`        // Map of string keys to their entries
	Version        string                 `json:"version"`        // Version of the format (e.g., "1.0")
}

// StringEntry represents an individual string entry within the "strings" map.
// It contains metadata and optional translations.
type StringEntry struct {
	ExtractionState string                  `json:"extractionState"`         // Metadata about how the string was extracted (e.g., "manual")
	Localizations   map[string]Localization `json:"localizations,omitempty"` // Optional translations for different languages
}

// Localization represents a translation for a specific language within a StringEntry.
type Localization struct {
	StringUnit StringUnit `json:"stringUnit"` // The actual translation data
}

// StringUnit holds the translation value and its state (e.g., "translated").
type StringUnit struct {
	State string `json:"state"` // State of the translation (e.g., "translated")
	Value string `json:"value"` // The translated string
}

// readJsonFile reads and parses an .xcstrings JSON file into an Xcstrings struct.
// Args:
//
//	filename: Path to the .xcstrings file.
//
// Returns:
//
//	*Xcstrings: Pointer to the parsed struct, or nil if an error occurs.
//	error: Any error encountered during file opening or JSON decoding.
func readJsonFile(filename string) (*Xcstrings, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening JSON file: %v", err)
	}
	defer file.Close()

	var xc Xcstrings
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&xc); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %v", err)
	}
	return &xc, nil
}

// getSortedLanguages extracts all unique language codes from the Xcstrings struct and sorts them.
// The source language (e.g., "en") is prioritized to appear first.
// Args:
//
//	xc: Pointer to the Xcstrings struct.
//
// Returns:
//
//	[]string: Sorted list of language codes.
func getSortedLanguages(xc *Xcstrings) []string {
	langSet := make(map[string]bool)
	langSet[xc.SourceLanguage] = true // Always include the source language

	// Collect languages from localizations
	for _, entry := range xc.Strings {
		for lang := range entry.Localizations {
			langSet[lang] = true
		}
	}

	languages := make([]string, 0, len(langSet))
	for lang := range langSet {
		languages = append(languages, lang)
	}

	// Sort with source language first
	sort.Slice(languages, func(i, j int) bool {
		if languages[i] == xc.SourceLanguage {
			return true
		}
		if languages[j] == xc.SourceLanguage {
			return false
		}
		return languages[i] < languages[j]
	})
	return languages
}

// createCsvFile generates a CSV file from the parsed .xcstrings data.
// The first column contains string keys, and subsequent columns contain translations for each language.
// Args:
//
//	xc: Pointer to the Xcstrings struct.
//	languages: List of language codes to include as columns.
//	outputFile: Path to the output CSV file.
//
// Returns:
//
//	error: Any error encountered during file creation or writing.
func createCsvFile(xc *Xcstrings, languages []string, outputFile string) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header: empty first cell, then language codes
	header := append([]string{""}, languages...)
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing CSV header: %v", err)
	}

	// Collect and sort keys
	keys := make([]string, 0, len(xc.Strings))
	for key := range xc.Strings {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Write rows: key followed by translations
	for _, key := range keys {
		row := make([]string, len(languages)+1)
		row[0] = key // First column is the key
		entry := xc.Strings[key]
		for i, lang := range languages {
			if lang == xc.SourceLanguage && len(entry.Localizations) == 0 {
				// Use key as the source language value if no localizations exist
				row[i+1] = key
			} else if loc, ok := entry.Localizations[lang]; ok {
				// Use the translated value if available
				row[i+1] = loc.StringUnit.Value
			} else {
				row[i+1] = "" // Empty if no translation
			}
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("error writing CSV row for key '%s': %v", key, err)
		}
	}

	return nil
}

// readCsvFile reads a CSV file and reconstructs an Xcstrings struct.
// The first column is assumed to be keys, and subsequent columns are translations.
// Args:
//
//	filename: Path to the input CSV file.
//
// Returns:
//
//	*Xcstrings: Pointer to the reconstructed Xcstrings struct.
//	[]string: List of language codes from the header.
//	error: Any error encountered during file reading or parsing.
func readCsvFile(filename string) (*Xcstrings, []string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("error opening CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("error reading CSV: %v", err)
	}
	if len(records) < 2 {
		return nil, nil, fmt.Errorf("CSV file must have at least 2 rows")
	}

	languages := records[0][1:] // Languages from second column onward
	if len(languages) == 0 {
		return nil, nil, fmt.Errorf("no languages found in the CSV file")
	}

	sourceLang := languages[0] // First language is the source language
	xc := &Xcstrings{
		SourceLanguage: sourceLang,
		Strings:        make(map[string]StringEntry),
		Version:        "1.0",
	}

	// Process each row
	keySet := make(map[string]bool)
	for _, row := range records[1:] {
		if len(row) == 0 {
			continue
		}
		key := row[0] // Key from first column
		if keySet[key] {
			return nil, nil, fmt.Errorf("duplicate key found: %s", key)
		}
		keySet[key] = true

		entry := StringEntry{
			ExtractionState: "manual", // Default value
			Localizations:   make(map[string]Localization),
		}
		for i, translation := range row[1:] {
			if i >= len(languages) {
				break
			}
			if strings.TrimSpace(translation) != "" {
				entry.Localizations[languages[i]] = Localization{
					StringUnit: StringUnit{
						State: "translated",
						Value: translation,
					},
				}
			}
		}
		xc.Strings[key] = entry
	}

	return xc, languages, nil
}

// writeJsonFile converts an Xcstrings struct to a formatted JSON file and prints it for verification.
// Args:
//
//	xc: Pointer to the Xcstrings struct.
//	outputFile: Path to the output JSON file.
//
// Returns:
//
//	error: Any error encountered during JSON marshalling or file writing.
func writeJsonFile(xc *Xcstrings, outputFile string) error {
	jsonData, err := json.MarshalIndent(xc, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %v", err)
	}

	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		return fmt.Errorf("error writing JSON file: %v", err)
	}

	fmt.Println("Generated JSON structure:")
	fmt.Println(string(jsonData))
	return nil
}

// jsonToCsv converts an .xcstrings JSON file to a CSV file.
// Args:
//
//	inputFile: Path to the input JSON file.
//	outputFile: Path to the output CSV file.
//
// Returns:
//
//	error: Any error encountered during the conversion.
func jsonToCsv(inputFile, outputFile string) error {
	xc, err := readJsonFile(inputFile)
	if err != nil {
		return err
	}

	languages := getSortedLanguages(xc)
	if len(languages) == 0 {
		return fmt.Errorf("no languages found in the JSON file")
	}

	if err := createCsvFile(xc, languages, outputFile); err != nil {
		return err
	}

	fmt.Printf("CSV file '%s' created successfully.\n", outputFile)
	return nil
}

// csvToJson converts a CSV file back to an .xcstrings JSON file.
// Args:
//
//	inputFile: Path to the input CSV file.
//	outputFile: Path to the output JSON file.
//
// Returns:
//
//	error: Any error encountered during the conversion.
func csvToJson(inputFile, outputFile string) error {
	xc, languages, err := readCsvFile(inputFile)
	if err != nil {
		return err
	}

	fmt.Printf("Languages found: %v\n", languages)
	if err := writeJsonFile(xc, outputFile); err != nil {
		return err
	}

	fmt.Printf("JSON file '%s' created successfully.\n", outputFile)
	return nil
}

// main is the entry point of the program, parsing command-line flags and executing the chosen mode.
func main() {
	// Define command-line flags
	mode := flag.String("mode", "", "Operation mode: 'json2csv' or 'csv2json'")
	input := flag.String("input", "", "Input file path")
	output := flag.String("output", "", "Output file path")
	flag.Parse()

	// Validate mode
	if *mode != "json2csv" && *mode != "csv2json" {
		fmt.Println("Error: -mode must be 'json2csv' or 'csv2json'")
		fmt.Println("Usage:")
		fmt.Println("  go run localization_converter.go -mode json2csv -input Localizable.xcstrings -output translations.csv")
		fmt.Println("  go run localization_converter.go -mode csv2json -input translations.csv -output Localizable.xcstrings")
		os.Exit(1)
	}

	// Validate input/output flags
	if *input == "" || *output == "" {
		fmt.Println("Error: -input and -output flags are required")
		fmt.Println("Usage:")
		fmt.Println("  go run localization_converter.go -mode json2csv -input Localizable.xcstrings -output translations.csv")
		fmt.Println("  go run localization_converter.go -mode csv2json -input translations.csv -output Localizable.xcstrings")
		os.Exit(1)
	}

	// Execute the chosen mode
	switch *mode {
	case "json2csv":
		if err := jsonToCsv(*input, *output); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "csv2json":
		if err := csvToJson(*input, *output); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}
