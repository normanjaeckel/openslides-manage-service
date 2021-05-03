package manage_start

import (
	_ "embed" // Blank import required to use go directive.
	"fmt"
	"html/template"
	"io"
	"path"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
)

// Constants to be used in docker-compose.yml template.
const (
	DockerRegistry     = "ghcr.io/normanjaeckel/openslides"
	OpenSlidesTag      = "4.0.0-dev"
	ExternalHTTPPort   = "8000"
	ExternalManagePort = "9008"
)

// dockerComposeYmlPath returns the path to docker-compose.yml file if it exists
// in current directory or in XGD Data directory. If the file does not exist, it
// returns "-" so that Docker Compose can read config from stdin.
func dockerComposeYmlPath() string {
	p, err := xdg.SearchDataFile(path.Join(appName, "docker-compose.yml"))
	if err != nil {
		return "-"
	}
	return p
}

// servicesYML contains the definitions for all OpenSlides services.
//go:embed services.yml
var servicesYML string

// service holds the metadata of an OpenSlides service to be used in
// docker-compose.yml template.
type service struct {
	Image string
	Path  string
	Args  struct {
		MODULE string
		PORT   string
	}
	Environment map[string]string
}

// services provides a map with all OpenSlides services.
func services() (map[string]service, error) {
	var s map[string]service
	if err := yaml.Unmarshal([]byte(servicesYML), &s); err != nil {
		return nil, fmt.Errorf("unmarshalling servivesYML: %w", err)
	}
	return s, nil
}

// defaultDockerCompose contains the template of the default docker-compose.yml.
//go:embed docker-compose.yml.tpl
var defaultDockerCompose string

// tplData holds the data used to execute the docker-compose.yml template.
type tplData struct {
	ExternalHTTPPort   string
	ExternalManagePort string
	Service            map[string]string
	Secret             map[string]string
}

// constructDockerComposeYML writes the populated template to the given writer.
// If remote is true it uses GitHub URIs for the build context. Else it uses
// relative paths to local code as provided in OpenSlides main repository.
func constructDockerComposeYML(w io.Writer, remote bool, secretsParentPath string) error {
	composeTPL, err := template.New("compose").Parse(defaultDockerCompose)
	if err != nil {
		return fmt.Errorf("creating Docker Compose template: %w", err)
	}
	composeTPL.Option("missingkey=error")

	td := tplData{
		ExternalHTTPPort:   ExternalHTTPPort,
		ExternalManagePort: ExternalManagePort,
	}

	if err := populateServices(&td, remote); err != nil {
		return fmt.Errorf("populating services to template data: %w", err)
	}

	populateSecrets(&td, secretsParentPath)

	if err := composeTPL.Execute(w, td); err != nil {
		return fmt.Errorf("writing Docker Compose file: %w", err)
	}

	return nil
}

// populateServices is a small helper function that populates service metadata
// to the given template data.
func populateServices(td *tplData, remote bool) error {
	services, err := services()
	if err != nil {
		return fmt.Errorf("getting services: %w", err)
	}

	td.Service = make(map[string]string, len(services))

	if remote {
		// Remote case with image URL to Docker Registry.
		for name, service := range services {
			td.Service[name] = fmt.Sprintf(
				"image: %s/%s:%s\n    environment:%s",
				DockerRegistry,
				service.Image,
				OpenSlidesTag,
				servicesEnv(),
			)
		}
		return nil
	}

	// Local case with build context pointing to relativ path.
	fragment := `env_file: services.env
    image: %s
    build:
      context: ./%s`

	fragmentSuffix := `
      args:
        MODULE: %s
        PORT: %s`

	for name, service := range services {
		s := fmt.Sprintf(
			fragment,
			fmt.Sprintf("%s:%s", service.Image, OpenSlidesTag),
			service.Path,
		)
		if service.Args.MODULE != "" || service.Args.PORT != "" {
			s += fmt.Sprintf(
				fragmentSuffix,
				service.Args.MODULE,
				service.Args.PORT,
			)
		}
		if len(service.Environment) > 0 {
			s += "\n    environment:"
			for k, v := range service.Environment {
				s += fmt.Sprintf("\n      - %s=%s", k, v)
			}
		}
		td.Service[name] = s
	}
	return nil
}
