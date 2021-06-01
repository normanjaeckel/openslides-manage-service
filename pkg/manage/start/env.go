package start

import (
	"bytes"
	_ "embed" // Blank import required to use go directive.
	"fmt"
	"io/fs"
	"os"
	"path"
)

const envFileName = "services.env"

// defaultServiesEnv contains all default environment variables to be
// used in docker-compose.yml template.
//go:embed default_services.env
var defaultServicesEnv []byte

// servicesEnv returns a formated string (which cares of YAML syntax and
// indetion) with all default environment variables to be used in
// docker-compose.yml template.
func servicesEnv() string {
	res := ""
	env := bytes.Split(defaultServicesEnv, []byte("\n"))
	for _, e := range env {
		if len(e) == 0 {
			continue
		}
		res += "\n      - " + string(e)
	}
	return res
}

// createEnvFile creates a services.env file in the given directory.
func createEnvFile(d string) error {
	p := path.Join(d, envFileName)

	if fileExists(p) {
		fmt.Printf("File %s does already exist. Skip this step.\n", p)
		return nil
	}

	if err := os.WriteFile(p, defaultServicesEnv, fs.ModePerm); err != nil {
		return fmt.Errorf("write services file at %s: %w", p, err)
	}

	fmt.Printf("Successfully created file %s.\n", p)

	return nil
}
