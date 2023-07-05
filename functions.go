package main

import (
	"fmt"
	"strconv"

	"topwatcher/pkg/reader"

	"github.com/ashwanthkumar/slack-go-webhook"
)

func CheckPodRamUsage(configFile *reader.Configuration, podInfo []Info) ([]string, []string) {
	deploymentList := make([]string, 0)
	allkeys := make(map[string]bool)
	list := make([]string, 0)
	alerts := make([]string, 0)

	for each := range podInfo {
		for c := range podInfo[each].Pods {
			ramValueNew, _ := strconv.Atoi(podInfo[each].Pods[c]["ram"])
			if ramValueNew > configFile.Kubernetes.Threshold.Ram &&
				IsException(podInfo[each].Deployment, podInfo[each].Pods[c]["name"], exceptions) &&
				podInfo[each].Kind == "ReplicaSet" {
				alert := fmt.Sprintf("Pod %v from deployment %v has high ram usage. current ram usage is %v",
					podInfo[each].Pods[c]["name"], podInfo[each].Deployment, podInfo[each].Pods[c]["ram"])
				deploymentList = append(deploymentList, podInfo[each].Deployment)
				alerts = append(alerts, alert)
			}

		}
	}

	if len(deploymentList) > 0 {
		for _, item := range deploymentList {
			if _, value := allkeys[item]; !value {
				allkeys[item] = true
				list = append(list, item)
			}
		}

	} else {
		InfoLogger.Println("There is nothing to do!!!")
	}

	return alerts, list
}

func SendSlackPayload(configFile *reader.Configuration, alerts []string) {

	for _, alert := range alerts {
		InfoLogger.Println(alert)
		webhookUrl := configFile.Slack.WebhookUrl
		payload := slack.Payload{
			Text:     alert,
			Channel:  "#" + configFile.Slack.Channel,
			Username: configFile.Slack.UserName,
		}
		errorSendSlack := slack.Send(webhookUrl, "", payload)
		if len(errorSendSlack) > 0 {
			ErrorLogger.Printf("error: %s\n", errorSendSlack)
		}
	}
}

func IsException(deployment string, podName string, exception []string) bool {

	for _, name := range exception {
		if deployment == name {
			WarningLogger.Printf("'%v' was eliminated by exceptions!!!\n", podName)
			return false
		}
	}
	return true
}
