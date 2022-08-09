package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/ashwanthkumar/slack-go-webhook"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Configuration struct {
	Kubernetes struct {
		InsideCluster bool   `yaml:"inside-cluster"`
		Namespaces    string `yaml:"namespaces"`
		Threshold     struct {
			Ram int `yaml:"ram"`
		} `yaml:"threshold"`
	} `yaml:"kubernetes"`
	Slack struct {
		WebhookUrl string `yaml:"webhookurl"`
		Notify     bool   `yaml:"notify"`
		Channel    string `yaml:"channel"`
		UserName   string `yaml:"username"`
	} `yaml:"slack"`
}

func GetClusterAccess() (*kubernetes.Clientset, *rest.Config) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error Getting Kubernetes clientset: %v\n", err)
		os.Exit(1)
	}
	return clientSet, config
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
	for element := range podInfo {
		ramValue, _ := strconv.Atoi(podInfo[element]["ram"])
		if ramValue > configFile.Kubernetes.Threshold.Ram {
			alert = fmt.Sprintf("Pod %v from deployment %v has hich ram usage. current ram usage is %v",
				podInfo[element]["name"], podInfo[element]["deployment"], podInfo[element]["ram"])
			fmt.Println(alert)
			SendSlackPayload(configFile, alert)
		}
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

func main() {
	podDetailList := make([](map[string]string), 0)
	podMetricsDetailList := make([](map[string]string), 0)
	var configFile Configuration

	readFile(&configFile)
	clientSet, config := GetClusterAccess()
	pods, err := clientSet.CoreV1().Pods(configFile.Kubernetes.Namespaces).List(context.Background(), v1.ListOptions{})
	if err != nil {
		fmt.Printf("Error Getting Pods: %v\n", err)
		os.Exit(1)
	}
	for _, pod := range pods.Items {
		podDetail := map[string]string{
			"name":       pod.Name,
			"deployment": pod.Labels["app"],
		}
		podDetailList = append(podDetailList, podDetail)
	}

	metricsClientset, err := metricsv.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	podMetricsList, err := metricsClientset.MetricsV1beta1().PodMetricses(configFile.Kubernetes.Namespaces).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for _, v := range podMetricsList.Items {
		podMetricsDetail := map[string]string{
			"name": v.GetName(),
			"ram":  fmt.Sprintf("%v", v.Containers[0].Usage.Memory().Value()/(1024*1024)),
		}
		podMetricsDetailList = append(podMetricsDetailList, podMetricsDetail)
	}
	podInfo := MergePodMetricMaps(podDetailList, podMetricsDetailList)
	CheckPodRamUsage(configFile, podInfo)
}
