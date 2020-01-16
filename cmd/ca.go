package main

import (
	"os"

	"github.com/fergusn/muzeum/internal/pki"
	"github.com/spf13/cobra"
)

func init() {
	var cn string

	cmd := &cobra.Command{
		Use:   "ca",
		Short: "Generate CA certificate",

		Run: func(c *cobra.Command, args []string) {
			pki.Generate(os.Stdout, cn)
		},
	}

	cmd.PersistentFlags().StringVarP(&cn, "common-name", "", "muzeum", "--common-name muzeum")

	cli.AddCommand(cmd)
}
