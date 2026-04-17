package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ogormans-deptstack/kubectl-generate/cmd/kubectl-schemagen/manifest"
	"github.com/ogormans-deptstack/kubectl-generate/cmd/kubectl-schemagen/migrate"
	"github.com/ogormans-deptstack/kubectl-generate/cmd/kubectl-schemagen/scaffold"
)

var version = "dev"

func main() {
	if err := newRootCommand().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubectl-schemagen",
		Short: "OpenAPI schema-powered Kubernetes tools",
		Long: `kubectl-schemagen provides a suite of tools that leverage the cluster's
OpenAPI v3 schema for manifest generation, API migration, and scaffolding.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
	}

	cmd.AddCommand(manifest.NewCommand())
	cmd.AddCommand(migrate.NewCommand())
	cmd.AddCommand(scaffold.NewCommand())

	return cmd
}
