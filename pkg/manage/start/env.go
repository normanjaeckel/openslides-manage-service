package manage_start

import _ "embed" // Blank import required to use go directive.

// defaultServiesEnv contains all default environment variables to be
// used in docker-compose.yml template.
//go:embed default_services.env
var defaultServiesEnv []byte

// servicesEnv returns a formated string (which cares of YAML syntax and
// indetion) with all default environment variables to be used in
// docker-compose.yml template.
func servicesEnv() string {
	// TODO: Read defaultServicesEnv and parse it to formatted string.
	return "\n    - FOOOOO=barrrr"
}
