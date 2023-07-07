/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"flag"
	"log"
	"os"
	k8s "topwatcher/pkg/kubernetes"
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

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "topwatcher",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	var alerts []string
	var target []string

	clientSet, config := k8s.GetClusterAccess(&configFile)

	if len(configFile.Kubernetes.Namespaces) > 0 {
		if k8s.Contain(configFile.Kubernetes.Namespaces, clientSet) {
			podInfo := k8s.GetPodInfo(clientSet, &configFile, config)
			if configFile.Logging.Debug {
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
		k8s.RestartDeployment(clientSet, target)
	}

	if configFile.Slack.Notify && len(configFile.Slack.Channel) > 0 {
		SendSlackPayload(&configFile, alerts)
	} else {
		for _, alert := range alerts {
			InfoLogger.Println(alert)
		}
	}
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	var flags int

	flag.Parse()

	configFile = reader.ReadFile()
	if configFile.Logging.Debug {
		flags = log.Ldate | log.Ltime | log.Lshortfile
		DebugLogger = log.New(os.Stdout, "DEBUG ", flags)
	} else {
		flags = log.Ldate | log.Ltime
	}
	InfoLogger = log.New(os.Stdout, "INFO ", flags)
	WarningLogger = log.New(os.Stdout, "WARNING ", flags)
	ErrorLogger = log.New(os.Stdout, "ERROR ", flags)

	InfoLogger.Println("Starting topwatcher...")

	if configFile.Logging.Debug {
		DebugLogger.Println("Reading Configuration file...")
	}

	allkeys := make(map[string]bool)

	for _, item := range configFile.Kubernetes.Exceptions.Deployments {
		if _, value := allkeys[item]; !value {
			allkeys[item] = true
			exceptions = append(exceptions, item)
		}
	}
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.topwatcher.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
