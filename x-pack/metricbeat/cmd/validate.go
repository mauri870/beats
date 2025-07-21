package cmd

import (
	_ "embed"
	"fmt"
	"os"

	cmd "github.com/elastic/beats/v7/libbeat/cmd"
	"github.com/elastic/gojsonschema"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

//go:embed metricbeat-schema.yaml
var schema []byte

func addValidateCommand(command *cmd.BeatsRootCmd) {
	command.AddCommand(validateCmd())
}

func validateCmd() *cobra.Command {
	var printSchema bool
	command := &cobra.Command{
		Use:   "validate",
		Short: "Validate the provided configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
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
					fmt.Fprintf(cmd.ErrOrStderr(), "%s\n", desc)
				}

				return fmt.Errorf("%s is invalid", configFile)
			}

			fmt.Printf("%s is valid\n", configFile)
			return nil
		},
	}

	command.Flags().BoolVarP(&printSchema, "print-schema", "p", false, "Print the validation schema and exit")
	return command
}
