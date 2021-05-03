package manage

// import (
// 	"context"
// 	_ "embed" // Neeed for embed. See Docu of Go 1.16
// 	"fmt"
// 	"html/template"
// 	"io"
// 	"io/fs"
// 	"os"
// 	"path"

// 	"gopkg.in/yaml.v3"
// )

// // createDockerComposeYML creates a docker-compose.yml file in the current working directory
// // using a template. In remote mode it uses the GitHub API to fetch the required commit IDs
// // of all services. Else it uses relative paths to local code as provided in OpenSlides
// // main repository.
// func createDockerComposeYML(ctx context.Context, dataPath string, remote bool) error {
// 	p := path.Join(dataPath, "docker-compose.yml")

// 	if fileExists(p) {
// 		return nil
// 	}

// 	w, err := os.Create(p)
// 	if err != nil {
// 		return fmt.Errorf("creating file `%s`: %w", p, err)
// 	}
// 	defer w.Close()

// 	if err := constructDockerComposeYML(ctx, w, remote); err != nil {
// 		return fmt.Errorf("writing content to file `%s`: %w", p, err)
// 	}

// 	return nil
// }

// //go:embed default_services.env
// var defaultServiesEnv []byte

// func createEnvFile(dataPath string) error {
// 	p := path.Join(dataPath, "services.env")

// 	if fileExists(p) {
// 		return nil
// 	}

// 	if err := os.WriteFile(p, defaultServiesEnv, fs.ModePerm); err != nil {
// 		return fmt.Errorf("write services file at %s: %w", p, err)
// 	}
// 	return nil
// }
