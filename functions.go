package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ashwanthkumar/slack-go-webhook"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
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
	for _, item := range deploymentList {
		if _, value := allkeys[item]; !value {
			allkeys[item] = true
			list = append(list, item)
		}
	}

	exeptions := configFile.Kubernetes.Exeptions.Deployments
	list = append(list, exeptions...)

	for _, entry := range list {
		keys[entry]++
	}
	for k, v := range keys {
		if v == 1 {
			target = append(target, k)
		}
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

func RestartDeployment(clientSet *kubernetes.Clientset, target []string) {

	for _, deploymentName := range target {
		deploymentClient := clientSet.AppsV1().Deployments("default")
		data := fmt.Sprintf(`{"spec": {"template": {"metadata": {"annotations": {"kubectl.kubernetes.io/restartedAt": "%s"}}}}}`, time.Now().Format("20060102150405"))
		_, err := deploymentClient.Patch(context.TODO(), deploymentName, types.StrategicMergePatchType, []byte(data), v1.PatchOptions{})

		if err != nil {
			fmt.Println(err)
		}
	}
}

func GetPodInfo(clientSet *kubernetes.Clientset, configFile Configuration, config *rest.Config) ([]map[string]string, []map[string]string) {

	podDetailList := make([](map[string]string), 0)
	podMetricsDetailList := make([](map[string]string), 0)

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

	return podDetailList, podMetricsDetailList
}
