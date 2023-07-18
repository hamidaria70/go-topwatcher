/*
Copyright © 2023 HAMID ARIA hamidaria.70@gmail.com
*/
package cmd

import (
	"log"
	"net/http"
	"os"
	"topwatcher/pkg/reader"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "topwatcher",
	Short: "Topwatcher is a useful and handy golang code which is dedicated to Kubernetes clusters as a Cronjob.",
	Run: func(cmd *cobra.Command, args []string) {

		var flags int
		inputConfigPath, _ := cmd.Flags().GetString("config")
		inputExceptions, _ := cmd.Flags().GetStringSlice("exceptions")
		isDebugMode, _ := cmd.Flags().GetBool("debug")
		inputKubeConfig, _ := cmd.Flags().GetString("kubeconfig")

		gin.SetMode(gin.DebugMode)
		router := gin.Default()

		if _, err := os.Stat(inputConfigPath); err != nil {
			log.New(os.Stdout, "ERROR ", log.Ldate|log.Ltime).Println("Try again using switches , Run 'topwatcher start -h'")
			os.Exit(1)
		} else {
			configFile = reader.ReadFile(inputConfigPath)
		}

		if isDebugMode || configFile.Logging.Debug {
			flags = log.Ldate | log.Ltime | log.Lshortfile
			DebugLogger = log.New(os.Stdout, "DEBUG ", flags)
		} else {
			flags = log.Ldate | log.Ltime
		}

		InfoLogger = log.New(os.Stdout, "INFO ", flags)
		WarningLogger = log.New(os.Stdout, "WARNING ", flags)
		ErrorLogger = log.New(os.Stdout, "ERROR ", flags)

		router.GET("/apiv1/version", getVersion)
		router.Handle("LIST", "/apiv1/exceptions", func(c *gin.Context) {

			var exceptionsList []string
			allkeys := make(map[string]bool)

			if len(inputExceptions) > 0 {
				exceptionsList = inputExceptions
			} else {
				exceptionsList = configFile.Kubernetes.Exceptions.Deployments
			}

			for _, item := range exceptionsList {
				if _, value := allkeys[item]; !value {
					allkeys[item] = true
					exceptions = append(exceptions, item)
				}
			}

			c.IndentedJSON(http.StatusOK, exceptions)
			exceptions = nil
		})
		clientSet, config := GetClusterAccess(&configFile, isDebugMode, inputKubeConfig)
		router.GET("/apiv1/podinfo", func(c *gin.Context) {
			nameSpace := "default"
			podInfo := GetPodInfo(clientSet, &configFile, config, isDebugMode, nameSpace)

			c.IndentedJSON(http.StatusOK, podInfo)
		})
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
