/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
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
		var nameSpace string
		var exceptionsList []string
		var ramThreshold int
		var alerts []string
		var target []string
		allkeys := make(map[string]bool)
		isDebugMode, _ := cmd.Flags().GetBool("debug")
		isPodRestart, _ := cmd.Flags().GetBool("restart-pod")
		configPath, _ := cmd.Flags().GetString("config")
		namespace, _ := cmd.Flags().GetString("namespace")
		inputKubeConfig, _ := cmd.Flags().GetString("kubeconfig")
		inputRam, _ := cmd.Flags().GetInt("ram-threshold")
		inputExceptions, _ := cmd.Flags().GetStringSlice("exceptions")

		configFile = reader.ReadFile(configPath)
		fmt.Println(inputExceptions)
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

		if isDebugMode || configFile.Logging.Debug {
			flags = log.Ldate | log.Ltime | log.Lshortfile
			DebugLogger = log.New(os.Stdout, "DEBUG ", flags)
		} else {
			flags = log.Ldate | log.Ltime
		}
		InfoLogger = log.New(os.Stdout, "INFO ", flags)
		WarningLogger = log.New(os.Stdout, "WARNING ", flags)
		ErrorLogger = log.New(os.Stdout, "ERROR ", flags)

		InfoLogger.Println("Starting topwatcher...")

		clientSet, config := GetClusterAccess(&configFile, isDebugMode, inputKubeConfig)

		if len(namespace) > 0 {
			nameSpace = namespace
		} else {
			nameSpace = configFile.Kubernetes.Namespaces
		}

		if inputRam > 0 {
			ramThreshold = inputRam
		} else {
			ramThreshold = configFile.Kubernetes.Threshold.Ram
		}

		if len(nameSpace) > 0 {
			if Contain(nameSpace, clientSet, isDebugMode) {
				podInfo := GetPodInfo(clientSet, &configFile, config, isDebugMode, nameSpace)
				if isDebugMode || configFile.Logging.Debug {
					DebugLogger.Printf("Pods information list is: %v", podInfo)
				}
				if ramThreshold > 0 {
					alerts, target = CheckPodRamUsage(&configFile, podInfo)
				} else {
					ErrorLogger.Println("Ram value is not defined in configuration file")
					os.Exit(1)
				}
			} else {
				WarningLogger.Printf("'%v' namespace is not in the cluster!!", nameSpace)
			}
		} else {
			ErrorLogger.Println("Namespace is not defined")
			os.Exit(1)
		}

		if len(target) > 0 && (configFile.Kubernetes.PodRestart || isPodRestart) {
			RestartDeployment(clientSet, target, isDebugMode, nameSpace)
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
	startCmd.Flags().BoolP("restart-pod", "R", false, "Trigger pod restart")
	startCmd.Flags().StringP("config", "c", "./config.yaml", "Config file address")
	startCmd.Flags().StringP("namespace", "n", "", "Target namespace")
	startCmd.Flags().StringP("kubeconfig", "k", "", "Path to cluster kubeconfig")
	startCmd.Flags().IntP("ram-threshold", "r", 0, "Ram threshold")
	startCmd.Flags().StringSliceP("exceptions", "e", []string{}, "List of exception to prevent restarting	Note: comma separated without spaces --> deployment1,deployment2,deployment3")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
