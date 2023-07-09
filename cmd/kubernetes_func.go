package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"topwatcher/pkg/reader"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Info struct {
	Deployment string
	Kind       string
	Replicas   int
	Pods       []map[string]string
}

func GetClusterAccess(configFile *reader.Configuration) (*kubernetes.Clientset, *rest.Config) {
	var kubeConfigPath string

	if configFile.Kubernetes.Kubeconfig != "" {
		if configFile.Logging.Debug {
			DebugLogger.Println("Building kubeconfig file from configuration file")
		}
		kubeConfigPath = configFile.Kubernetes.Kubeconfig
	} else {
		if configFile.Logging.Debug {
			DebugLogger.Println("Reading kubeconfig from user home directory")
		}

		userHomeDir, err := os.UserHomeDir()
		if configFile.Logging.Debug {
			DebugLogger.Println("User home directory is: ", userHomeDir)
		}

		if err != nil {
			ErrorLogger.Println(err)
			os.Exit(1)
		}
		kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
		if configFile.Logging.Debug {
			DebugLogger.Println("Kubeconfig path is: ", kubeConfigPath)
		}

	}
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if configFile.Logging.Debug {
		DebugLogger.Println("Building kubeconfig file from the path")
	}

	if err != nil {
		ErrorLogger.Println(err)
		os.Exit(1)
	}

	clientSet, err := kubernetes.NewForConfig(kubeConfig)
	if configFile.Logging.Debug {
		DebugLogger.Println("Getting new clientset")
	}

	if err != nil {
		ErrorLogger.Printf("Error Getting Kubernetes clientset: %v\n", err)
		os.Exit(1)
	}

	return clientSet, kubeConfig
}

func RestartDeployment(clientSet *kubernetes.Clientset, target []string) {

	for _, deploymentName := range target {
		deploymentClient := clientSet.AppsV1().Deployments("default")
		data := fmt.Sprintf(`{"spec": {"template": {"metadata": {"annotations": {"kubectl.kubernetes.io/restartedAt": "%s"}}}}}`, time.Now().Format("20060102150405"))
		_, err := deploymentClient.Patch(context.TODO(), deploymentName, types.StrategicMergePatchType, []byte(data), v1.PatchOptions{})

		if configFile.Logging.Debug {
			DebugLogger.Printf("%v deployment was restarted by patching\n", deploymentName)
		}

		if err != nil {
			ErrorLogger.Println(err)
		}
	}
}

func GetPodInfo(clientSet *kubernetes.Clientset, configFile *reader.Configuration, config *rest.Config) []Info {
	var info Info
	info.Pods = make([]map[string]string, 0)
	var podInfo []Info

	podDetailList := make([](map[string]string), 0)
	podMetricsDetailList := make([](map[string]string), 0)

	pods, err := clientSet.CoreV1().Pods(configFile.Kubernetes.Namespaces).List(context.Background(), v1.ListOptions{})
	if err != nil {
		ErrorLogger.Printf("Error Getting Pods: %v\n", err)
		os.Exit(1)
	}

	for _, pod := range pods.Items {
		if pod.OwnerReferences[0].Kind != "Job" && pod.Status.Phase == "Running" {
			podDetail := map[string]string{
				"name":       pod.Name,
				"deployment": pod.Labels["app"],
				"kind":       pod.OwnerReferences[0].Kind,
			}
			podDetailList = append(podDetailList, podDetail)

		}
	}

	metricsClientset, err := metricsv.NewForConfig(config)
	if err != nil {
		ErrorLogger.Println(err)
		os.Exit(1)
	}

	podMetricsList, err := metricsClientset.MetricsV1beta1().PodMetricses(configFile.Kubernetes.Namespaces).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		ErrorLogger.Println(err)
		os.Exit(1)
	}
	if len(podMetricsList.Items) == len(podDetailList) {
		for v := range podMetricsList.Items {
			podMetricsDetail := map[string]string{
				"name": podMetricsList.Items[v].GetName(),
				"ram":  fmt.Sprintf("%v", podMetricsList.Items[v].Containers[0].Usage.Memory().Value()/(1024*1024)),
			}
			podMetricsDetailList = append(podMetricsDetailList, podMetricsDetail)
		}
	} else {
		if configFile.Logging.Debug {
			DebugLogger.Printf("length of podMetricsList: %v length of podDetailList: %v\n", podMetricsList.Items, podDetailList)
		}
		ErrorLogger.Println("Metrics are not available for some pods")
		os.Exit(1)
	}

	keys := make(map[string]int)
	for _, entry := range podDetailList {
		keys[entry["deployment"]]++
	}
	for j, n := range podDetailList {
		if n["name"] == podMetricsDetailList[j]["name"] {
			if info.Deployment != n["deployment"] && info.Deployment != "" {
				info.Pods = nil
			}
			info.Deployment = n["deployment"]
			info.Kind = n["kind"]
			info.Pods = append(info.Pods, podMetricsDetailList[j])

		}
		if len(info.Pods) == keys[info.Deployment] {
			info.Replicas = keys[info.Deployment]
			podInfo = append(podInfo, info)

		}
	}
	return podInfo
}

func Contain(nominated string, clientSet *kubernetes.Clientset) bool {
	var namespaceList []string

	namespace, err := clientSet.CoreV1().Namespaces().List(context.Background(), v1.ListOptions{})
	if err != nil {
		ErrorLogger.Println(err)
		os.Exit(2)
	}
	for _, namespace := range namespace.Items {
		namespaceList = append(namespaceList, namespace.Name)
	}

	for _, item := range namespaceList {
		if item == nominated {
			if configFile.Logging.Debug {
				DebugLogger.Printf("%v namespace exists inside the cluster\n", nominated)
			}
			return true
		}
	}
	return false
}
