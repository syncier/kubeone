package cmd

import (
	"errors"
	"fmt"

	"github.com/kubermatic/kubeone/pkg/installer"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type resetOptions struct {
	globalOptions
	Manifest       string
	DestroyWorkers bool
}

// resetCmd setups reset command
func resetCmd(rootFlags *pflag.FlagSet) *cobra.Command {
	ropts := &resetOptions{}
	cmd := &cobra.Command{
		Use:   "reset <manifest>",
		Short: "Revert changes",
		Long: `Undo all changes done by KubeOne to the configured machines.

This command takes KubeOne manifest which contains information about hosts.
It's possible to source information about hosts from Terraform output, using the '--tfjson' flag.`,
		RunE: func(_ *cobra.Command, args []string) error {
			gopts, err := persistentGlobalOptions(rootFlags)
			if err != nil {
				return err
			}

			logger := initLogger(gopts.Verbose)
			ropts.TerraformState = gopts.TerraformState
			ropts.Verbose = gopts.Verbose

			if len(args) != 1 {
				return errors.New("expected path to a cluster config file as an argument")
			}

			ropts.Manifest = args[0]
			if ropts.Manifest == "" {
				return errors.New("no cluster config file given")
			}

			return runReset(logger, ropts)
		},
	}

	cmd.Flags().BoolVarP(&ropts.DestroyWorkers, "destroy-workers", "", false, "destroy all worker machines before resetting cluster")

	return cmd
}

// runReset resets all machines provisioned by KubeOne
func runReset(logger *logrus.Logger, resetOptions *resetOptions) error {
	if resetOptions.Manifest == "" {
		return errors.New("no cluster config file given")
	}

	cluster, err := loadClusterConfig(resetOptions.Manifest)
	if err != nil {
		return fmt.Errorf("failed to load cluster: %v", err)
	}

	if err = applyTerraform(resetOptions.TerraformState, cluster); err != nil {
		return err
	}

	if err = cluster.DefaultAndValidate(); err != nil {
		return err
	}

	options := &installer.Options{
		Verbose:        resetOptions.Verbose,
		DestroyWorkers: resetOptions.DestroyWorkers,
	}

	return installer.NewInstaller(cluster, logger).Reset(options)
}