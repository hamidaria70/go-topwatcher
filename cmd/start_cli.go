/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"os"
	"topwatcher/pkg/reader"

	"github.com/spf13/cobra"
)

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
	DebugLogger   *log.Logger
	exceptions    []string
	configFile    reader.Configuration
	configPath    string
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var flags int
		var alerts []string
		var target []string
		allkeys := make(map[string]bool)
		debugMode, _ := cmd.Flags().GetBool("debug")
		configPath, _ = cmd.Flags().GetString("config")

		configFile = reader.ReadFile(configPath)

		for _, item := range configFile.Kubernetes.Exceptions.Deployments {
			if _, value := allkeys[item]; !value {
				allkeys[item] = true
				exceptions = append(exceptions, item)
			}
		}

		if debugMode || configFile.Logging.Debug {
			flags = log.Ldate | log.Ltime | log.Lshortfile
			DebugLogger = log.New(os.Stdout, "DEBUG ", flags)
		} else {
			flags = log.Ldate | log.Ltime
		}
		InfoLogger = log.New(os.Stdout, "INFO ", flags)
		WarningLogger = log.New(os.Stdout, "WARNING ", flags)
		ErrorLogger = log.New(os.Stdout, "ERROR ", flags)

		InfoLogger.Println("Starting topwatcher...")

		clientSet, config := GetClusterAccess(&configFile, debugMode)

		if len(configFile.Kubernetes.Namespaces) > 0 {
			if Contain(configFile.Kubernetes.Namespaces, clientSet, debugMode) {
				podInfo := GetPodInfo(clientSet, &configFile, config, debugMode)
				if debugMode || configFile.Logging.Debug {
					DebugLogger.Printf("Pods information list is: %v", podInfo)
				}
				if configFile.Kubernetes.Threshold.Ram > 0 {
					alerts, target = CheckPodRamUsage(&configFile, podInfo)
				} else {
					ErrorLogger.Println("Ram value is not defined in configuration file")
					os.Exit(1)
				}
			} else {
				WarningLogger.Printf("'%v' namespace is not in the cluster!!", configFile.Kubernetes.Namespaces)
			}
		} else {
			ErrorLogger.Println("Namespace is not defined")
			os.Exit(1)
		}

		if len(target) > 0 && configFile.Kubernetes.PodRestart {
			RestartDeployment(clientSet, target, debugMode)
		}

		if configFile.Slack.Notify && len(configFile.Slack.Channel) > 0 {
			SendSlackPayload(&configFile, alerts)
		} else {
			for _, alert := range alerts {
				InfoLogger.Println(alert)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().BoolP("debug", "d", false, "Turn on debug mode")
	startCmd.Flags().StringP("config", "c", "./config.yaml", "Config file address")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
