package resolve

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mutagen-io/mutagen/cmd"
	"github.com/mutagen-io/mutagen/pkg/service/synchronization"
)

// resolveMain is the entry point for the resolve command.
func resolveMain(command *cobra.Command, arguments []string) error {
	// Validate arguments.
	if len(arguments) != 2 {
		return fmt.Errorf("resolve requires exactly two arguments: session and resolution")
	}

	sessionID := arguments[0]
	resolution := arguments[1]

	if resolution != "alpha" && resolution != "beta" {
		return fmt.Errorf("resolution must be either 'alpha' or 'beta'")
	}

	// Create a session selection specification.
	selection := &synchronization.Selection{
		Specifications: []string{sessionID},
	}

	// Connect to the daemon and defer closure of the connection.
	daemonConnection, err := cmd.CreateDaemonClientConnection()
	if err != nil {
		return fmt.Errorf("unable to connect to daemon: %w", err)
	}
	defer daemonConnection.Close()

	// Create a synchronization service client.
	synchronizationService := synchronization.NewSynchronizationClient(daemonConnection)

	// Invoke the resolution.
	request := &synchronization.ResolveConflictsRequest{
		Selection:  selection,
		FavorAlpha: resolution == "alpha",
	}
	if _, err := synchronizationService.ResolveConflicts(context.Background(), request); err != nil {
		return fmt.Errorf("unable to resolve conflicts: %w", err)
	}

	// Success.
	fmt.Println("Conflicts resolved successfully.")
	return nil
}

// ResolveCommand is the resolve command.
var ResolveCommand = &cobra.Command{
	Use:          "resolve <session> <alpha|beta>",
	Short:        "Resolve conflicts in a synchronization session",
	RunE:         resolveMain,
	SilenceUsage: true,
}
