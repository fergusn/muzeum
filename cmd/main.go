package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var cli = &cobra.Command{
	Use:   "muzeum",
	Short: "Artifact Storage",
	Long:  `Artifact Storage`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

func main() {
	if err := cli.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
