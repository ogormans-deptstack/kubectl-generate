package scaffold

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scaffold RESOURCE_TYPE",
		Short: "Generate kustomize bases or Helm charts from cluster resources",
		Long: `Reverse-engineers running cluster resources into kustomize base directories
or Helm chart skeletons, using the cluster's OpenAPI schema for type-correct
field generation.`,
		Example: `  kubectl schemagen scaffold deployment --name=web -o kustomize
  kubectl schemagen scaffold --namespace=default -o helm`,
		Aliases: []string{"sc"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("scaffold is not yet implemented")
		},
	}

	return cmd
}
