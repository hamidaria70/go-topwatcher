/*
Copyright © 2023 HAMID ARIA hamidaria.70@gmail.com
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "topwatcher",
	Short: "Topwatcher is a useful and handy golang code which is dedicated to Kubernetes clusters as a Cronjob.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Under construction...")
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("debug", "d", false, "Turn on debug mode")
	rootCmd.Flags().StringP("config", "c", "./config.yaml", "Config file address")
}
