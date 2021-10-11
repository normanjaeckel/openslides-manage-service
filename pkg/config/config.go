package config

import (
	"bytes"
	_ "embed" // Blank import required to use go directive.
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/OpenSlides/openslides-manage-service/pkg/shared"
	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
	"github.com/spf13/cobra"
)

//go:embed default-docker-compose.yml
var defaultDockerComposeYml []byte

//go:embed default-config.yml
var defaultConfig []byte

const (
	// ConfigHelp contains the short help text for the command.
	ConfigHelp = "Rebuilds the YAML file for using Docker Compose or Docker Swarm"

	// ConfigHelpExtra contains the long help text for the command without the headline.
	ConfigHelpExtra = `This command (re)creates a YAML file in the given directory.`
)

// Cmd returns the setup subcommand.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config directory",
		Short: ConfigHelp,
		Long:  ConfigHelp + "\n\n" + ConfigHelpExtra,
		Args:  cobra.ExactArgs(1),
	}

	tplFile := FlagTpl(cmd)
	configFiles := FlagConfig(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		dir := args[0]

		var tpl []byte
		if *tplFile != "" {
			fc, err := os.ReadFile(*tplFile)
			if err != nil {
				return fmt.Errorf("reading file %q: %w", *tplFile, err)
			}
			tpl = fc
		}

		var config [][]byte
		if len(*configFiles) > 0 {
			for _, configFile := range *configFiles {
				fc, err := os.ReadFile(configFile)
				if err != nil {
					return fmt.Errorf("reading file %q: %w", configFile, err)
				}
				config = append(config, fc)
			}
		}

		if err := Config(dir, tpl, config); err != nil {
			return fmt.Errorf("running Config(): %w", err)
		}
		return nil
	}
	return cmd
}

// FlagTpl setups the template flag to the given cobra command.
func FlagTpl(cmd *cobra.Command) *string {
	return cmd.Flags().StringP("template", "t", "", "custom YAML template file")
}

// FlagConfig setups the config flag to the given cobra command.
func FlagConfig(cmd *cobra.Command) *[]string {
	return cmd.Flags().StringArrayP("config", "c", nil, "custom YAML config file, can be use more then once, ordering is important")
}

// Config rebuilds the YAML file for using Docker Compose or Docker Swarm.
//
// A custom template for the YAML file and YAML configs can be provided.
func Config(dir string, tplContent []byte, cfgContent [][]byte) error {
	// Create directory
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("creating directory at %q: %w", dir, err)
	}

	// Create YAML file
	if err := CreateYmlFile(dir, true, tplContent, cfgContent); err != nil {
		return fmt.Errorf("creating YAML file at %q: %w", dir, err)
	}

	return nil
}

// CreateYmlFile builds the YAML file at the given directory. Use a truthy value for force
// to override an existing file.
func CreateYmlFile(dir string, force bool, tplContent []byte, cfgContent [][]byte) error {
	if tplContent == nil {
		tplContent = defaultDockerComposeYml
	}

	cfg, err := newYmlConfig(cfgContent)
	if err != nil {
		return fmt.Errorf("creating new YML config object: %w", err)
	}

	marshalContentFunc := func(v interface{}) (string, error) {
		y, err := yaml.Marshal(v)
		if err != nil {
			return "", err
		}
		result := "\n"
		for _, line := range strings.Split(string(y), "\n") {
			if len(line) != 0 {
				result += fmt.Sprintf("    %s\n", line)
			}
		}
		result = strings.TrimRight(result, "\n")
		return result, nil
	}
	funcMap := template.FuncMap{}
	funcMap["marshalContent"] = marshalContentFunc

	tmpl, err := template.New("YAML File").Option("missingkey=error").Funcs(funcMap).Parse(string(tplContent))
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	var res bytes.Buffer
	if err := tmpl.Execute(&res, cfg); err != nil {
		return fmt.Errorf("executing template %v: %w", tmpl, err)
	}

	if err := shared.CreateFile(dir, force, cfg.Filename, res.Bytes(), false); err != nil {
		return fmt.Errorf("creating YAML file at %q: %w", dir, err)
	}

	return nil
}

type ymlConfig struct {
	Filename string `yaml:"filename"`

	Host string `yaml:"host"`
	Port string `yaml:"port"`

	// TODO: Remove these two fields.
	ManageHost string `yaml:"manageHost"`
	ManagePort string `yaml:"managePort"`

	DisablePostgres  bool `yaml:"disablePostgres"`
	DisableDependsOn bool `yaml:"disableDependsOn"`

	Defaults struct {
		ContainerRegistry string `yaml:"containerRegistry"`
		Tag               string `yaml:"tag"`
	} `yaml:"defaults"`

	DefaultEnvironment map[string]string `yaml:"defaultEnvironment"`

	Services map[string]service `yaml:"services"`
}

type service struct {
	ContainerRegistry string          `yaml:"containerRegistry"`
	Tag               string          `yaml:"tag"`
	AdditionalContent json.RawMessage `yaml:"additionalContent"`
}

func newYmlConfig(configFiles [][]byte) (*ymlConfig, error) {
	// Reverse config files
	for i := len(configFiles)/2 - 1; i >= 0; i-- {
		opp := len(configFiles) - 1 - i
		configFiles[i], configFiles[opp] = configFiles[opp], configFiles[i]
	}

	// Append default config
	configFiles = append(configFiles, defaultConfig)

	// Unmarshal and merge them all
	config := new(ymlConfig)
	for _, configFile := range configFiles {
		c := new(ymlConfig)
		if err := yaml.Unmarshal(configFile, c); err != nil {
			return nil, fmt.Errorf("unmarshaling YAML: %w", err)
		}
		if err := mergo.Merge(config, c); err != nil {
			return nil, fmt.Errorf("merging config files: %w", err)
		}
	}

	// Fill services
	allServices := []string{
		"proxy",
		"client",
		"backend",
		"datastoreReader",
		"datastoreWriter",
		"postgres",
		"autoupdate",
		"auth",
		"redis",
		"media",
		"icc",
		"manage",
	}
	if len(config.Services) == 0 {
		config.Services = make(map[string]service, len(allServices))
	}

	for _, name := range allServices {
		_, ok := config.Services[name]
		if !ok {
			config.Services[name] = service{}
		}
		s := config.Services[name]

		if s.ContainerRegistry == "" {
			s.ContainerRegistry = config.Defaults.ContainerRegistry
		}
		if s.Tag == "" {
			s.Tag = config.Defaults.Tag
		}

		config.Services[name] = s
	}

	return config, nil
}