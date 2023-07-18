/*
Copyright © 2023 HAMID ARIA hamidaria.70@gmail.com
*/
package cmd

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "topwatcher",
	Short: "Topwatcher is a useful and handy golang code which is dedicated to Kubernetes clusters as a Cronjob.",
	Run: func(cmd *cobra.Command, args []string) {
		gin.SetMode(gin.DebugMode)
		router := gin.Default()

		router.GET("/apiv1/version", getVersion)
		router.Run()
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

func getVersion(c *gin.Context) {
	version := "v0.2.0"

	c.IndentedJSON(http.StatusOK, gin.H{
		"version": version,
	})
}
