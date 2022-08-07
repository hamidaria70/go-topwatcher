package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

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

func CheckPodRamUsage(podInfo []map[string]string) {
	for element := range podInfo {
		ramValue, _ := strconv.Atoi(podInfo[element]["ram"])
		if ramValue > 4 {
			fmt.Printf("Pod %v from deployment %v has hich ram usage. current ram usage is %v\n",
				podInfo[element]["name"], podInfo[element]["deployment"], podInfo[element]["ram"])
		} else {
			os.Exit(1)
		}
	}
}

func main() {
	podDetailList := make([](map[string]string), 0)
	podMetricsDetailList := make([](map[string]string), 0)

	clientSet, config := GetClusterAccess()

	pods, err := clientSet.CoreV1().Pods("default").List(context.Background(), v1.ListOptions{})
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

	podMetricsList, err := metricsClientset.MetricsV1beta1().PodMetricses("default").List(context.TODO(), v1.ListOptions{})
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
	CheckPodRamUsage(podInfo)
}
