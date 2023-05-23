package main

import (
	"context"
	"fmt"
	"os"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
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
