package manage_start

import (
	_ "embed" // Blank import required to use go directive.
	"fmt"
	"html/template"
	"io"
	"os"
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
func constructDockerComposeYML(w io.Writer, remote bool, secretsParentPath string, fullEnv bool) error {
	composeTPL, err := template.New("compose").Parse(defaultDockerCompose)
	if err != nil {
		return fmt.Errorf("creating Docker Compose template: %w", err)
	}
	composeTPL.Option("missingkey=error")

	td := tplData{
		ExternalHTTPPort:   ExternalHTTPPort,
		ExternalManagePort: ExternalManagePort,
	}

	if err := populateServices(&td, remote, fullEnv); err != nil {
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
func populateServices(td *tplData, remote, fullEnv bool) error {
	services, err := services()
	if err != nil {
		return fmt.Errorf("getting services: %w", err)
	}

	td.Service = make(map[string]string, len(services))
	for name, service := range services {
		if remote {
			td.Service[name] = populateServiceRemoteCase(service.Image, fullEnv) // TODO: Use flag.
		} else {
			td.Service[name] = populateServiceLocalCase(service)
		}
	}
	return nil
}

// populateServiceRemoteCase is a small helper function that populates service
// metadata to the given template data for one service in remote case. If
// fullEnv is true all environment variables are inserted, else we use the
// services.env file.
func populateServiceRemoteCase(img string, fullEnv bool) string {
	s := fmt.Sprintf(
		"image: %s/%s:%s",
		DockerRegistry,
		img,
		OpenSlidesTag,
	)
	if fullEnv {
		return s + "\n    environment:" + servicesEnv()
	}
	return s + "\n    env_file: " + envFileName
}

// populateServiceLocalCase is a small helper function that populates service metadata
// to the given template data for one service in local case.
func populateServiceLocalCase(service service) string {
	fragment := `env_file: %s
    image: %s
    build:
      context: ./%s`
	fragmentSuffix := `
      args:
        MODULE: %s
        PORT: %s`

	s := fmt.Sprintf(
		fragment,
		envFileName,
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

	return s
}

// createDockerComposeYML creates a docker-compose.yml file in the given
// directory using a template. In remote mode it uses URLs to GitHub Container
// Registry of all services. Else it uses relative paths to local code as
// provided in OpenSlides main repository.
func createDockerComposeYML(d string, remote bool) error {
	p := path.Join(d, "docker-compose.yml")

	if fileExists(p) {
		fmt.Printf("File %s does already exist. Skip this step.\n", p)
		return nil
	}

	w, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("creating file `%s`: %w", p, err)
	}
	defer w.Close()

	if err := constructDockerComposeYML(w, remote, d, false); err != nil {
		return fmt.Errorf("writing content to file `%s`: %w", p, err)
	}

	fmt.Printf("Successfully created file %s.\n", p)

	return nil
}
