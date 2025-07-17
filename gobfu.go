/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gobfuscator

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

//go:embed sql/postgres-template.gsql
var sqlTemplate []byte

//go:embed data/default-data.yaml
var defaultDataYaml []byte

type Item struct {
	table     string
	column    string
	sqlType   string
	generator string
}

type Formatter interface {
	Format(item *Item) string
}

type SQLFormat func(item *Item) string
type SQLFormatter struct {
	PhoneNumber SQLFormat `table:"phone-numbers"`
	Email       SQLFormat `table:"email"`
	Word        SQLFormat `table:"words"`
	Null        SQLFormat `table:"null"`
	Address     SQLFormat `table:"addresses-1"`
	Business    SQLFormat `table:"businesses"`
	Default     SQLFormat `table:"__gob_default__"`
}

// this needs to be extracted.
// and loaded as a map.
// the yaml file will dictate where each function comes from and loaded
// by the matching name. Not a struct

func (f *SQLFormatter) Format(item *Item) string {
	var format SQLFormat

	switch item.generator {
	case "phone-numbers":
		format = f.PhoneNumber
	case "email":
		format = f.Email
	case "words":
		format = f.Word
	case "null":
		format = f.Null
	case "addresses-1":
		format = f.Address
	case "businesses":
		format = f.Business
	default:
		format = f.Default
	}

	return format(item)
}

func GenerateUpdates(config string, formatter Formatter) (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current directory: %v", err)
	}

	var buffer bytes.Buffer

	yamlFile := filepath.Join(currentDir, config)
	items, err := readItems(yamlFile)
	if err != nil {
		return "", fmt.Errorf("error reading mapping file %s: %v", config, err)
	}

	grouped := make(map[string][]Item)
	for _, item := range items {
		grouped[item.table] = append(grouped[item.table], item)
	}

	tables := make([]string, 0, len(grouped))
	for table := range grouped {
		tables = append(tables, table)
	}
	sort.Strings(tables)

	for _, table := range tables {
		group := grouped[table]

		sort.Slice(group, func(i, j int) bool {
			return group[i].column < group[j].column
		})

		var updates []string
		for _, item := range group {
			updates = append(updates, formatter.Format(&item))
		}
		_, _ = fmt.Fprintf(&buffer, "update \"%s\" set %s;\n", table, strings.Join(updates, ",\n  "))
	}

	return buffer.String(), nil
}

func GenerateInserts(custom string) (string, error) {
	allData := make(map[string][]string)

	// embedded default from build
	if err := toYaml(defaultDataYaml, allData); err != nil {
		return "", fmt.Errorf("error reading embedded default YAML data: %v", err)
	}

	// If custom is specified, read the single YAML file
	if custom != "" {
		if err := appendFromFile(custom, allData); err != nil {
			return "", fmt.Errorf("error reading custom YAML file %s: %v", custom, err)
		}
	}

	var insertBuffer bytes.Buffer
	for kind, values := range allData {
		for index, value := range values {
			escapedValue := strings.ReplaceAll(value, "'", "''")
			// Fprintf to a buffer doesn't return errors.. just ignoring the error explicitly
			_, _ = fmt.Fprintf(&insertBuffer, "insert into gobfuscator_anon_data (idx, kind, value) values (%d, '%s', '%s');\n",
				index, kind, escapedValue)
		}
	}

	return insertBuffer.String(), nil
}

func ApplyTemplate(inserts string, updates string) (string, error) {

	tmpl, err := template.New("postgres-template").Parse(string(sqlTemplate))

	if err != nil {
		return "", fmt.Errorf("error parsing template: %v", err)
	}

	data := struct {
		Inserts string
		Updates string
	}{
		Inserts: inserts,
		Updates: updates,
	}

	var resultBuffer bytes.Buffer
	if err := tmpl.Execute(&resultBuffer, data); err != nil {
		return "", fmt.Errorf("error executing template: %v", err)
	}

	return resultBuffer.String(), nil
}

// WriteToFile writes the provided content to a file at the specified path.
// It handles file existence checks and user confirmation for overwriting.
func WriteToFile(path string, content string) (int, error) {
	output, err := attemptFile(path)
	if err != nil {
		return 0, err
	}

	defer output.Close()

	len, err := output.WriteString(content)
	if err != nil {
		return 0, fmt.Errorf("error writing to file: %v", err)
	}

	log.Printf("Successfully wrote: %s", path)
	return len, nil
}

// attemptFile creates a file at the specified path. If the file already exists,
// it prompts the user for confirmation before overwriting.
//
// Parameters:
//   - path: The file path where the file should be created
//
// Returns:
//   - *os.File: A file handle to the newly created file
//   - error: Any error encountered during file creation
func attemptFile(path string) (*os.File, error) {
	var output *os.File
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("File %s already exists. Overwrite? (y/n): ", path)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("error reading input: %v", err)
		}
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			return nil, fmt.Errorf("operation cancelled by user")
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("error creating output directory: %v", err)
	}

	output, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("error creating output file: %v", err)
	}

	return output, nil
}

type YAMLMapping map[string]map[string]struct {
	Type   string `yaml:"type"`
	Source string `yaml:"source"`
}

func readItems(yamlPath string) ([]Item, error) {
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("error reading YAML file: %v", err)
	}

	var mapping YAMLMapping
	if err := yaml.Unmarshal(data, &mapping); err != nil {
		return nil, fmt.Errorf("error unmarshaling YAML: %v", err)
	}

	var items []Item
	for table, columns := range mapping {
		for column, details := range columns {
			generator := details.Source
			if generator == "" || generator == "null" {
				generator = "null"
			}

			items = append(items, Item{
				table:     table,
				column:    column,
				sqlType:   details.Type,
				generator: generator,
			})
		}
	}

	return items, nil
}

func appendFromFile(path string, data map[string][]string) error {
	fileData, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return toYaml(fileData, data)
}

func toYaml(fileData []byte, data map[string][]string) error {
	var yamlData map[string][]string
	if err := yaml.Unmarshal(fileData, &yamlData); err != nil {
		return err
	}

	for kind, values := range yamlData {
		data[kind] = append(data[kind], values...)
	}

	return nil
}
