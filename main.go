package main

import (
	"context"
	"fmt"
	"os"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

func main() {
	podDetailList := make([](map[string]string), 0)
	podMetricsDetailList := make([](map[string]string), 0)

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error Getting Kubernetes clientset: %v\n", err)
		os.Exit(1)
	}

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
			"ram":  fmt.Sprintf("%vMi", v.Containers[0].Usage.Memory().Value()/(1024*1024)),
		}
		podMetricsDetailList = append(podMetricsDetailList, podMetricsDetail)
	}
	fmt.Println(podDetailList)
	fmt.Println(podMetricsDetailList)
}
