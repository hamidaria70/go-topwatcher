package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ashwanthkumar/slack-go-webhook"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetClusterAccess() (*kubernetes.Clientset, *rest.Config) {
	userHomeDir, err := os.UserHomeDir()

	if err != nil {
		panic(err.Error())
	}
	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)

	if err != nil {
		panic(err.Error())
	}
	clientSet, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		fmt.Printf("Error Getting Kubernetes clientset: %v\n", err)
		os.Exit(1)
	}
	return clientSet, kubeConfig
}

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

func CheckPodRamUsage(configFile Configuration, podInfo []map[string]string) {
	alert := ""
	deploymentList := make([]string, 0)
	allkeys := make(map[string]bool)
	deploymentListPurified := make([]string, 0)
	for element := range podInfo {
		ramValue, _ := strconv.Atoi(podInfo[element]["ram"])
		if ramValue > configFile.Kubernetes.Threshold.Ram {
			alert = fmt.Sprintf("Pod %v from deployment %v has high ram usage. current ram usage is %v",
				podInfo[element]["name"], podInfo[element]["deployment"], podInfo[element]["ram"])
			deploymentList = append(deploymentList, podInfo[element]["deployment"])
			fmt.Println(alert)
			if configFile.Slack.Notify {
				SendSlackPayload(configFile, alert)
			}
		}
	}
	for _, item := range deploymentList {
		if _, value := allkeys[item]; !value {
			allkeys[item] = true
			deploymentListPurified = append(deploymentListPurified, item)
		}
	}

	for _, deploymentName := range deploymentListPurified {
		fmt.Printf("Restarting deployment %v\n", deploymentName)
		fmt.Println("************************************")
		fmt.Println(v1.DeploymentList{})
		fmt.Println("************************************")

	}
}

func SendSlackPayload(configFile Configuration, alert string) {

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

func processError(err error) {
	fmt.Println(err)
	os.Exit(2)
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
