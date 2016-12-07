package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/infrakit/pkg/cli"
	instance_plugin "github.com/docker/infrakit/pkg/rpc/instance"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:   os.Args[0],
		Short: "GCE instance plugin",
	}

	name := cmd.Flags().String("name", "instance-gce", "Plugin name to advertise for discovery")
	logLevel := cmd.Flags().Int("log", cli.DefaultLogLevel, "Logging level. 0 is least verbose. Max is 5")
	project := cmd.Flags().String("project", "code-story-blog", "Google Cloud project")
	zone := cmd.Flags().String("zone", "europe-west1-d", "Google Cloud zone")

	cmd.Run = func(c *cobra.Command, args []string) {
		cli.SetLogLevel(*logLevel)
		cli.RunPlugin(*name, instance_plugin.PluginServer(NewGCEInstancePlugin(*project, *zone)))
	}

	cmd.AddCommand(cli.VersionCommand())

	err := cmd.Execute()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
