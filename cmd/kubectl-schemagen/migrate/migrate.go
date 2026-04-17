package migrate

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate [FILE...]",
		Short: "Detect and rewrite deprecated Kubernetes APIs",
		Long: `Reads YAML manifests and detects deprecated or removed API versions by
comparing against the connected cluster's OpenAPI schema. Optionally rewrites
the manifests in-place to use the current API version.`,
		Example: `  kubectl schemagen migrate deployment.yaml
  kubectl schemagen migrate --dry-run manifests/
  kubectl schemagen migrate --in-place *.yaml`,
		Aliases: []string{"mig"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("migrate is not yet implemented")
		},
	}

	return cmd
}
