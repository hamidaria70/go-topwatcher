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
	pod_detail_list := make([](map[string]string), 0)

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error Getting Kubernetes clientset: %v\n", err)
		os.Exit(1)
	}

	pods, err := clientset.CoreV1().Pods("default").List(context.Background(), v1.ListOptions{})
	if err != nil {
		fmt.Printf("Error Getting Pods: %v\n", err)
		os.Exit(1)
	}
	for _, pod := range pods.Items {
		pod_detail := map[string]string{
			"name":       pod.Name,
			"deployment": pod.Labels["app"],
		}
		pod_detail_list = append(pod_detail_list, pod_detail)
	}
	fmt.Println(pod_detail_list)

	metricsClientset, err := metricsv.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	podMetricsList, err := metricsClientset.MetricsV1beta1().PodMetricses("default").List(context.TODO(), v1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for _, v := range podMetricsList.Items {
		fmt.Printf("Pod Name: %s , RAM Usage: %vMi\n", v.GetName(), v.Containers[0].Usage.Memory().Value()/(1024*1024))
	}
}
