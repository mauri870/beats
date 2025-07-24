// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package cmd

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/elastic/beats/v7/libbeat/cmd/instance"
	"github.com/elastic/gojsonschema"
)

// GenValidateCmd generates the command version for a Beat.
func GenValidateCmd(settings instance.Settings) *cobra.Command {
	var printSchema bool
	var writeReference string
	command := &cobra.Command{
		Use:   "validate",
		Short: "Validate the provided configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			schema := settings.Schema
			if printSchema {
				fmt.Println(string(schema))
				return nil
			}

			if writeReference != "" {
				if err := writeConfigReference(settings, writeReference, schema); err != nil {
					return fmt.Errorf("failed to write reference config: %w", err)
				}

				fmt.Printf("%s created\n", writeReference)
				return validateConfigFile(writeReference, schema, cmd.ErrOrStderr())
			}

			configFile, err := cmd.Flags().GetString("c")
			if err != nil {
				return err
			}
			if configFile == "" {
				return fmt.Errorf("config file must be provided with -c")
			}

			if err := validateConfigFile(configFile, schema, cmd.ErrOrStderr()); err != nil {
				return fmt.Errorf("failed to validate config file %s: %w", configFile, err)
			}

			return nil
		},
	}

	command.Flags().BoolVarP(&printSchema, "print-schema", "p", false, "Print the validation schema and exit")
	command.Flags().StringVar(&writeReference, "write-reference", "", "Generate a reference configuration file based on the schema")
	return command
}

// validateConfigFile validates the provided configuration file against the schema.
// Any validation errors found will be printed to w.
func validateConfigFile(configFile string, schema []byte, w io.Writer) error {
	userCfg, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}

	jsonSchema, err := yaml.YAMLToJSON(schema)
	if err != nil {
		return fmt.Errorf("failed to parse schema: %w", err)
	}
	jsonUserCfg, err := yaml.YAMLToJSON(userCfg)
	if err != nil {
		return fmt.Errorf("failed to parse user config: %w", err)
	}

	result, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(jsonSchema),
		gojsonschema.NewBytesLoader(jsonUserCfg),
	)
	if err != nil {
		return fmt.Errorf("failed to validate user config %s: %w", configFile, err)
	}

	if !result.Valid() {
		for _, desc := range result.Errors() {
			fmt.Fprintln(w, desc)
		}

		fmt.Printf("%s is invalid\n", configFile)
		os.Exit(1)
		return nil
	}

	fmt.Printf("%s is valid\n", configFile)
	return nil
}

func writeConfigReference(settings instance.Settings, referenceFilename string, schema []byte) error {
	// Parse the schema YAML
	var schemaMap map[string]interface{}
	if err := yaml.Unmarshal(schema, &schemaMap); err != nil {
		return fmt.Errorf("failed to parse schema YAML: %w", err)
	}

	// Generate reference config from schema dynamically
	configContent := generateDynamicReferenceConfig(schemaMap, settings.Name)

	// Write to file
	if err := os.WriteFile(referenceFilename, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write reference config to %s: %w", referenceFilename, err)
	}

	return nil
}

// generateDynamicReferenceConfig generates a reference configuration dynamically from schema
func generateDynamicReferenceConfig(schema map[string]interface{}, beatName string) string {
	var result strings.Builder

	// Add header
	result.WriteString(fmt.Sprintf("########################## %s Configuration ###########################\n\n", strings.ToUpper(beatName)))
	result.WriteString("# This file is a full configuration example documenting all non-deprecated\n")
	result.WriteString("# options in comments. For a shorter configuration example, that contains only\n")
	result.WriteString(fmt.Sprintf("# the most common options, please see %s.yml in the same directory.\n", beatName))
	result.WriteString("#\n")
	result.WriteString("# You can find the full configuration reference here:\n")
	result.WriteString(fmt.Sprintf("# https://www.elastic.co/guide/en/beats/%s/index.html\n\n", beatName))

	// Get properties from schema
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		return result.String()
	}

	// Sort property keys for consistent output
	var keys []string
	for key := range properties {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Generate each property dynamically
	for _, key := range keys {
		if prop, exists := properties[key]; exists {
			result.WriteString(generatePropertyFromSchema(key, prop))
			result.WriteString("\n")
		}
	}

	return result.String()
}

// generatePropertyFromSchema generates property configuration from schema metadata
func generatePropertyFromSchema(key string, prop interface{}) string {
	var result strings.Builder

	propMap, ok := prop.(map[string]interface{})
	if !ok {
		return ""
	}

	// Add property description as comment if available
	if description, hasDescription := propMap["description"].(string); hasDescription {
		result.WriteString("# " + strings.ReplaceAll(description, "\n", "\n# ") + "\n")
	}

	// Generate the property configuration
	propType, _ := propMap["type"].(string)
	defaultValue, hasDefault := propMap["default"]

	// If property has no defined type, comment it out to avoid validation issues
	if propType == "" {
		result.WriteString(fmt.Sprintf("#%s:\n", key))
		return result.String()
	}

	switch propType {
	case "object":
		result.WriteString(generateObjectProperty(key, propMap, hasDefault, defaultValue))
	case "array":
		result.WriteString(generateArrayProperty(key, propMap, hasDefault, defaultValue))
	default:
		result.WriteString(generateScalarProperty(key, propMap, hasDefault, defaultValue))
	}

	return result.String()
}

func generateObjectProperty(key string, propMap map[string]interface{}, hasDefault bool, defaultValue interface{}) string {
	// Handle object properties
	if properties, hasProps := propMap["properties"].(map[string]interface{}); hasProps {
		// Check if any sub-property has a default value or should be enabled
		hasEnabledSubProps := false
		for _, subProp := range properties {
			if subPropMap, ok := subProp.(map[string]interface{}); ok {
				if _, hasSubDefault := subPropMap["default"]; hasSubDefault {
					hasEnabledSubProps = true
					break
				}
			}
		}

		// If no sub-properties have defaults, comment out the entire object
		if !hasEnabledSubProps && !hasDefault {
			return fmt.Sprintf("#%s:\n", key)
		}

		var result strings.Builder
		result.WriteString(key + ":\n")

		// Sort keys for consistent output
		var keys []string
		for k := range properties {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			if subProp, ok := properties[k].(map[string]interface{}); ok {
				// Add sub-property description as comment
				if description, hasDescription := subProp["description"].(string); hasDescription {
					result.WriteString(fmt.Sprintf("  # %s\n", description))
				}

				// Add sub-property value
				if subDefault, hasSubDefault := subProp["default"]; hasSubDefault {
					result.WriteString(fmt.Sprintf("  %s: %s\n", k, formatSchemaValue(subDefault)))
				} else {
					subType, _ := subProp["type"].(string)
					if subType == "object" {
						// For object types without defaults, just comment out the key without a value
						result.WriteString(fmt.Sprintf("  #%s:\n", k))
					} else {
						result.WriteString(fmt.Sprintf("  #%s: %s\n", k, getExampleValue(subType, k)))
					}
				}
			}
		}
		return result.String()
	} else if hasDefault {
		// Object has a default value
		return fmt.Sprintf("%s: %s\n", key, formatSchemaValue(defaultValue))
	} else {
		// Empty object, comment it out to avoid validation errors
		return fmt.Sprintf("#%s:\n", key)
	}
}

func generateArrayProperty(key string, propMap map[string]interface{}, hasDefault bool, defaultValue interface{}) string {
	var result strings.Builder

	if hasDefault {
		result.WriteString(fmt.Sprintf("%s: %s\n", key, formatSchemaValue(defaultValue)))
	} else {
		result.WriteString(fmt.Sprintf("#%s:\n", key))

		// Add array item example if items schema is available
		if items, hasItems := propMap["items"].(map[string]interface{}); hasItems {
			if itemType, hasItemType := items["type"].(string); hasItemType {
				if itemType == "object" {
					result.WriteString("#  - {}\n")
				} else {
					result.WriteString(fmt.Sprintf("#  - %s\n", getExampleValue(itemType, key)))
				}
			}
		}
	}

	return result.String()
}

func generateScalarProperty(key string, propMap map[string]interface{}, hasDefault bool, defaultValue interface{}) string {
	if hasDefault {
		return fmt.Sprintf("%s: %s\n", key, formatSchemaValue(defaultValue))
	} else {
		propType, _ := propMap["type"].(string)
		return fmt.Sprintf("#%s: %s\n", key, getExampleValue(propType, key))
	}
}

func formatSchemaValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		if v == "" {
			return `""`
		}
		// Don't quote values that look like variable substitutions or are already formatted
		if strings.Contains(v, "${") || strings.Contains(v, ":") || strings.Contains(v, "/") {
			return v
		}
		return fmt.Sprintf(`"%s"`, v)
	case bool:
		return fmt.Sprintf("%v", v)
	case int, int64, float64:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func getExampleValue(propType, key string) string {
	switch propType {
	case "string":
		if strings.Contains(key, "host") {
			return "localhost"
		} else if strings.Contains(key, "path") {
			return "/path/to/file"
		} else if strings.Contains(key, "url") {
			return "http://localhost"
		} else if strings.Contains(key, "index") {
			return "metricbeat-%{[agent.version]}"
		}
		return `""`
	case "boolean":
		return "false"
	case "integer":
		return "0"
	case "number":
		return "0.0"
	case "array":
		return "[]"
	case "object":
		return "{}"
	default:
		return `""`
	}
}
