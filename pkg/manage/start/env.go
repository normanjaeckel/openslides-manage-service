package manage_start

import (
	"bytes"
	_ "embed"
) // Blank import required to use go directive.

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
