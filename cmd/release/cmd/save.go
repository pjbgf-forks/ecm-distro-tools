package cmd

import (
	"errors"

	"github.com/rancher/ecm-distro-tools/release/images"
	"github.com/spf13/cobra"
)

var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "save k3s images to a tarball",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("expected at least two arguments: [k3s-version] [output.tar]")
		}
		version, output := args[0], args[1]
		_, found := rootConfig.K3s.Versions[version]
		if !found {
			return errors.New("verify your config file, version not found: " + version)
		}

		return images.Save(version, output)
	},
}

func init() {
	rootCmd.AddCommand(saveCmd)
}
