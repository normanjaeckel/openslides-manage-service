package manage_start

import (
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/adrg/xdg"
)

// SetupOpenSlidesFiles creates the required files to run OpenSlides with Docker
// Compose.
func SetupOpenSlidesFiles(useXdg, remote bool) error {
	var dataPath string
	if useXdg {
		dataPath = path.Join(xdg.DataHome, appName)
	}
	if dataPath == "" {
		p, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current directory: %w", err)
		}
		dataPath = p
	}

	if err := os.MkdirAll(dataPath, fs.ModePerm); err != nil {
		return fmt.Errorf("creating directory `%s`: %w", dataPath, err)
	}

	if err := createDockerComposeYML(dataPath, remote); err != nil {
		return fmt.Errorf("creating Docker Compose YML: %w", err)
	}

	if err := createEnvFile(dataPath); err != nil {
		return fmt.Errorf("creating .env file: %w", err)
	}

	if err := createSecrets(dataPath); err != nil {
		return fmt.Errorf("creating secrets: %w", err)
	}

	return nil
}
