package cmd

import (
	"fmt"

	"github.com/k0sproject/k0sctl/analytics"
	"github.com/k0sproject/k0sctl/phase"
	"github.com/k0sproject/k0sctl/pkg/apis/k0sctl.k0sproject.io/v1beta1"
	"github.com/k0sproject/k0sctl/pkg/apis/k0sctl.k0sproject.io/v1beta1/cluster"
	"github.com/urfave/cli/v2"
)

var kubeconfigCommand = &cli.Command{
	Name:  "kubeconfig",
	Usage: "Output the admin kubeconfig of the cluster",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "address",
			Usage: "Set kubernetes API address (default: auto-detect)",
			Value: "",
		},
		configFlag,
		debugFlag,
		traceFlag,
		redactFlag,
		analyticsFlag,
	},
	Before: actions(initSilentLogging, initConfig, initAnalytics),
	After: func(ctx *cli.Context) error {
		analytics.Client.Close()
		return nil
	},
	Action: func(ctx *cli.Context) error {
		c := ctx.Context.Value(ctxConfigKey{}).(*v1beta1.Cluster)

		// Change so that the internal config has only single controller host as we
		// do not need to connect to all nodes
		c.Spec.Hosts = cluster.Hosts{c.Spec.K0sLeader()}
		manager := phase.Manager{Config: c}

		manager.AddPhase(
			&phase.Connect{},
			&phase.DetectOS{},
			&phase.GetKubeconfig{APIAddress: ctx.String("address")},
			&phase.Disconnect{},
		)

		if err := manager.Run(); err != nil {
			return err
		}

		fmt.Println(c.Metadata.Kubeconfig)

		return nil
	},
}
