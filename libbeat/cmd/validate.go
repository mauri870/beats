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
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/elastic/beats/v7/libbeat/cmd/instance"
	"github.com/elastic/gojsonschema"
)

// GenValidateCmd generates the command version for a Beat.
func GenValidateCmd(settings instance.Settings) *cobra.Command {
	var printSchema bool
	command := &cobra.Command{
		Use:   "validate",
		Short: "Validate the provided configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			schema := settings.Schema
			if printSchema {
				fmt.Println(string(schema))
				return nil
			}

			configFile, err := cmd.Flags().GetString("c")
			if err != nil {
				return err
			}
			if configFile == "" {
				return fmt.Errorf("config file must be provided with -c")
			}

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
					fmt.Fprintln(cmd.ErrOrStderr(), desc)
				}

				fmt.Printf("%s is invalid\n", configFile)
				os.Exit(1)
				return nil
			}

			fmt.Printf("%s is valid\n", configFile)
			return nil
		},
	}

	command.Flags().BoolVarP(&printSchema, "print-schema", "p", false, "Print the validation schema and exit")
	return command
}
