package start

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/adrg/xdg"
)

const appName = "openslides"

// StartOpenSlides runs docker-compose up to start all OpenSlides services. It
// also creates initial data if the database is empty.
func StartOpenSlides(ctx context.Context) error {
	dcp := dockerComposeYmlPath()
	dcArgs := []string{"--file", dcp, "up", "--build"}
	dc := exec.CommandContext(ctx, "docker-compose", dcArgs...)

	if dcp == "-" {
		appHome := path.Join(xdg.DataHome, appName)
		var w strings.Builder
		if err := constructDockerComposeYML(&w, true, appHome, true); err != nil {
			return fmt.Errorf("constructing Docker Compose configuration: %w", err)
		}

		r := strings.NewReader(w.String())
		if err := createSecrets(appHome); err != nil {
			return fmt.Errorf("creating secrets: %w", err)
		}
		dc.Stdin = r
	}
	dc.Stdout = os.Stdout
	dc.Stderr = os.Stderr

	if err := dc.Run(); err != nil {
		return fmt.Errorf("running docker-compose with `%s`: %w", dc.String(), err)
	}
	// TODO:
	// - Use Goroutine for the Run() call or use --detach flag.
	// - Run initial-data.
	// - Provide graceful stop with docker-compose stop and kill after second Ctrl-C. Aternative: Provide stop command.

	return nil
}

// fileExists checks if the file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
