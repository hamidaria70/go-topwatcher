package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ashwanthkumar/slack-go-webhook"
	"gopkg.in/yaml.v2"
)

func MergePodMetricMaps(podDetailList []map[string]string, podMetricsDetailList []map[string]string) []map[string]string {
	for i := range podDetailList {
		podName1 := podDetailList[i]["name"]
		for a := range podMetricsDetailList {
			podName2 := podMetricsDetailList[a]["name"]

			if podName1 == podName2 {
				podDetailList[i]["ram"] = podMetricsDetailList[a]["ram"]
			}
		}
	}
	return podDetailList
}

func CheckPodRamUsage(configFile Configuration, podInfo []map[string]string) ([]string, []string) {
	deploymentList := make([]string, 0)
	allkeys := make(map[string]bool)
	target := make([]string, 0)
	keys := make(map[string]int)
	list := make([]string, 0)
	alerts := make([]string, 0)

	for element := range podInfo {
		ramValue, _ := strconv.Atoi(podInfo[element]["ram"])
		if ramValue > configFile.Kubernetes.Threshold.Ram {
			alert := fmt.Sprintf("Pod %v from deployment %v has high ram usage. current ram usage is %v",
				podInfo[element]["name"], podInfo[element]["deployment"], podInfo[element]["ram"])
			deploymentList = append(deploymentList, podInfo[element]["deployment"])
			alerts = append(alerts, alert)
		}
	}

	if len(deploymentList) > 0 {
		for _, item := range deploymentList {
			if _, value := allkeys[item]; !value {
				allkeys[item] = true
				list = append(list, item)
			}
		}

		exeptions := configFile.Kubernetes.Exeptions.Deployments
		var newExeptions []string

		for _, item := range exeptions {
			for _, element := range list {
				if item == element {
					newExeptions = append(newExeptions, item)
				}
			}
		}

		list = append(list, newExeptions...)

		for _, entry := range list {
			keys[entry]++
		}
		for k, v := range keys {
			if v == 1 {
				target = append(target, k)
			}
		}
		if len(target) == 0 {
			fmt.Println("targets were eliminated by exeptions!!!")
		}
	} else {
		fmt.Println("there is nothing to do!!!")
	}

	return alerts, target
}

func SendSlackPayload(configFile Configuration, alerts []string) {

	for _, alert := range alerts {
		fmt.Println(alert)
		webhookUrl := configFile.Slack.WebhookUrl
		payload := slack.Payload{
			Text:     alert,
			Channel:  "#" + configFile.Slack.Channel,
			Username: configFile.Slack.UserName,
		}
		errorSendSlack := slack.Send(webhookUrl, "", payload)
		if len(errorSendSlack) > 0 {
			fmt.Printf("error: %s\n", errorSendSlack)
		}
	}
}

func readFile(configFile *Configuration) {
	file, err := os.Open("config.yaml")

	if err != nil {
		processError(err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(configFile)

	if err != nil {
		processError(err)
	}
}
