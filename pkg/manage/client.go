package manage

import (
	"context"
	"fmt"
	"time"

	manage_start "github.com/OpenSlides/openslides-manage-service/pkg/manage/start"
	"github.com/OpenSlides/openslides-manage-service/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const rootHelp = `manage is an admin tool to perform manager actions on an OpenSlides instance.`

func cmdRoot(cfg *ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "manage",
		Short:        "Swiss army knife for OpenSlides admins",
		Long:         rootHelp,
		SilenceUsage: true,
	}

	// TODO: Move address and timeout to the commands that need it.
	cmd.PersistentFlags().StringVarP(&cfg.Address, "address", "a", "localhost:9008", "Address of the OpenSlides manage service")
	cmd.PersistentFlags().DurationVarP(&cfg.Timeout, "timeout", "t", 5*time.Second, "Time to wait for the command's response")

	return cmd
}

// RunClient starts the root command.
func RunClient() error {
	cfg := new(ClientConfig)
	cmd := cmdRoot(cfg)
	cmd.AddCommand(
		cmdStart(cfg),
		cmdSetup(cfg),
		//CmdCompose(cfg),
		CmdCheckServer(cfg),
		CmdInitialData(cfg),
		CmdCreateUser(cfg),
		CmdSetPassword(cfg),
		CmdConfig(cfg),
		cmdTunnel(cfg),
	)
	return cmd.Execute()
}

// ClientConfig holds the top level arguments.
type ClientConfig struct {
	Address string
	Timeout time.Duration
}

// Dial creates a grpc connection to the server.
func Dial(ctx context.Context, address string) (proto.ManageClient, func() error, error) {
	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, nil, fmt.Errorf("creating gRPC client connection with grpc.DialContect(): %w", err)
	}
	return proto.NewManageClient(conn), conn.Close, nil
}

const startHelp = `Starts OpenSlides

This starts OpenSlides with the Docker Compose configuration given in current
directory, in XDG Data directory or with default configuration. It also creates
local secrets and initial data if it is the first start.

Eventually the required Docker images are downloaded which may take a while.
`

// cmdStart starts OpenSlides eventually with new secrets and initial-data.
func cmdStart(cfg *ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Starts OpenSlides",
		Long:  startHelp,
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if err := manage_start.StartOpenSlides(ctx); err != nil {
			return fmt.Errorf("starting OpenSlides: %w", err)
		}

		return nil
	}

	return cmd
}

const setupHelp = `Provides configuration files for OpenSlides

This command executes the following steps to install OpenSlides:
- Create a docker-compose.yml in current directory or XDG Data directory
- Create the services.env next to the docker-compose.yml.
- Create local secrets for the auth service and admin password.
`

// cmdSetup creates docker-compose.yml, secrets and services.env.
func cmdSetup(cfg *ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Provides configuration files for OpenSlides",
		Long:  setupHelp,
	}

	cwd := cmd.Flags().Bool("cwd", false, "Create required files in currend working directory.")
	local := cmd.Flags().Bool("local", false, "Use local code to build images instead of URIs to GitHub. This requires --cwd to be set.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if *local && !*cwd {
			return fmt.Errorf("--local requires --cwd to be set")
		}

		if err := manage_start.SetupOpenSlidesFiles(!*cwd, !*local); err != nil {
			return fmt.Errorf("installing OpenSlides: %w", err)
		}

		return nil
	}

	return cmd
}
