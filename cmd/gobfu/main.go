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

// Binary gobfu is a tool for database obfuscation.
package main

import (
	"flag"
	"fmt"
	"github.com/acidbluebriggs/gobfuscator"
	"log"
	"os"
)

type configuration struct {
	//mode       string
	config     string
	outputFile string
	dialect    string
	custom     string
	formatter  gobfuscator.Formatter
}

func main() {

	c, err := loadConfig()
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	var inserts string
	var updates string

	inserts, err = gobfuscator.GenerateInserts(c.custom)
	if err != nil {
		log.Fatalf("Error generating inserts: %v", err)
	}

	updates, err = gobfuscator.GenerateUpdates(c.config, c.formatter)
	if err != nil {
		log.Fatalf("Error generating updates: %v", err)
	}

	result, err := gobfuscator.ApplyTemplate(inserts, updates)
	if err != nil {
		log.Fatalf("Error applying template: %v", err)
	}

	file, err := gobfuscator.WriteToFile(c.outputFile, result)
	if err != nil {
		log.Fatalf("Error writing file: %v", err)
	}

	log.Printf("Wrote %d bytes", file)
}

func loadConfig() (*configuration, error) {

	c := configuration{}
	//flag.StringVar(&c.mode, "mode", "", "GenerateInserts obfuscation data or scripts [data, obfuscate]")
	flag.StringVar(&c.config, "config", "", "Path to the yaml database table/colum configuration")
	flag.StringVar(&c.outputFile, "output", "", "The path of the output file to write to")
	flag.StringVar(&c.dialect, "dialect", "postgresql", "Database dialect (currently only supports postgresql)")
	// TODO next pass
	//flag.StringVar(&c.custom, "custom", "", "File containing custom entries for obfuscation data")
	flag.Parse()

	if c.config == "" || c.outputFile == "" {
		return nil, fmt.Errorf("usage: %s -config <config> -output <output_file>", os.Args[0])
	}

	// TODO yes this is a hack, need cobra
	var formatter gobfuscator.Formatter
	if f, exists := formatters[c.dialect]; !exists {
		return nil, fmt.Errorf("dialect '%s' does not exist", c.dialect)
	} else {
		formatter = f
	}
	c.formatter = formatter

	return &c, nil
}

var formatters = map[string]gobfuscator.Formatter{
	"postgresql": &gobfuscator.Postgres,
}
