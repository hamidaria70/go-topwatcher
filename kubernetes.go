package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

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
			"kind":       pod.OwnerReferences[0].Kind,
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

func Contain(nominated string, clientSet *kubernetes.Clientset) bool {
	var namespaceList []string

	namespace, err := clientSet.CoreV1().Namespaces().List(context.Background(), v1.ListOptions{})
	if err != nil {
		processError(err)
	}
	for _, namespace := range namespace.Items {
		namespaceList = append(namespaceList, namespace.Name)
	}

	for _, item := range namespaceList {
		if item == nominated {
			return true
		}
	}
	return false
}
