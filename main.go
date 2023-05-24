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

	readFile(&configFile)

	clientSet, config := GetClusterAccess()
	podDetailList, podMetricsDetailList := GetPodInfo(clientSet, configFile, config)

	podInfo := MergePodMetricMaps(podDetailList, podMetricsDetailList)
	alerts, target := CheckPodRamUsage(configFile, podInfo)

	if len(target) > 0 {
		RestartDeployment(clientSet, target)
	}

	if configFile.Slack.Notify {
		SendSlackPayload(configFile, alerts)
	} else {
		for _, alert := range alerts {
			fmt.Println(alert)
		}
	}
}
