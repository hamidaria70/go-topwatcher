package main

import (
	"log"
	"os"
)

type Configuration struct {
	Kubernetes struct {
		Namespaces string `yaml:"namespaces"`
		Threshold  struct {
			Ram int `yaml:"ram"`
		} `yaml:"threshold"`
		Exceptions struct {
			Deployments []string `yaml:"deployments,flow"`
		} `yaml:"exceptions"`
		PodRestart bool `yaml:"podrestart"`
	} `yaml:"kubernetes"`
	Slack struct {
		WebhookUrl string `yaml:"webhookurl"`
		Notify     bool   `yaml:"notify"`
		Channel    string `yaml:"channel"`
		UserName   string `yaml:"username"`
	} `yaml:"slack"`
	Logging struct {
		Debug bool `yaml:"debug"`
	} `yaml:"logging"`
}

type Info struct {
	Deployment string
	Kind       string
	Replicas   int
	Pods       []map[string]string
}

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
	DebugLogger   *log.Logger
	configFile    Configuration
	exceptions    []string
)

func init() {
	var flags int

	readFile(&configFile)
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
}

func main() {
	var alerts []string
	var target []string

	clientSet, config := GetClusterAccess()

	if len(configFile.Kubernetes.Namespaces) > 0 {
		if Contain(configFile.Kubernetes.Namespaces, clientSet) {
			podInfo := GetPodInfo(clientSet, configFile, config)
			if configFile.Logging.Debug {
				DebugLogger.Printf("Pods information list is: %v", podInfo)
			}
			if configFile.Kubernetes.Threshold.Ram > 0 {
				alerts, target = CheckPodRamUsage(configFile, podInfo)
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
		RestartDeployment(clientSet, target)
	}

	if configFile.Slack.Notify && len(configFile.Slack.Channel) > 0 {
		SendSlackPayload(configFile, alerts)
	} else {
		for _, alert := range alerts {
			InfoLogger.Println(alert)
		}
	}
}
