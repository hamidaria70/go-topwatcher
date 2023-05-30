package main

import (
	"fmt"
)

type Configuration struct {
	Kubernetes struct {
		Namespaces string `yaml:"namespaces"`
		Threshold  struct {
			Ram int `yaml:"ram"`
		} `yaml:"threshold"`
		Exeptions struct {
			Deployments []string `yaml:"deployments,flow"`
		} `yaml:"exeptions"`
	} `yaml:"kubernetes"`
	Slack struct {
		WebhookUrl string `yaml:"webhookurl"`
		Notify     bool   `yaml:"notify"`
		Channel    string `yaml:"channel"`
		UserName   string `yaml:"username"`
	} `yaml:"slack"`
}

func main() {
	var configFile Configuration
	var alerts []string
	var target []string

	readFile(&configFile)

	clientSet, config := GetClusterAccess()

	if len(configFile.Kubernetes.Namespaces) > 0 {
		if Contain(configFile.Kubernetes.Namespaces, clientSet) {
			podDetailList, podMetricsDetailList := GetPodInfo(clientSet, configFile, config)
			podInfo := MergePodMetricMaps(podDetailList, podMetricsDetailList)
			alerts, target = CheckPodRamUsage(configFile, podInfo)
		} else {
			fmt.Println("nominated namespace is not in the cluster!!")
		}
	} else {
		fmt.Println("namespace is not defined")
	}

	if len(target) > 0 {
		RestartDeployment(clientSet, target)
	}

	if configFile.Slack.Notify && len(configFile.Slack.Channel) > 0 {
		SendSlackPayload(configFile, alerts)
	} else {
		for _, alert := range alerts {
			fmt.Println(alert)
		}
	}
}
