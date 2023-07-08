package reader

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Configuration struct {
	Kubernetes struct {
		Kubeconfig string `yaml:"kubeconfig"`
		Namespaces string `yaml:"namespaces"`
		Threshold  struct {
			Ram int `yaml:"ram"`
		} `yaml:"threshold"`
		Exceptions struct {
			Deployments []string `yaml:"deployments,flow"`
		} `yaml:"exceptions"`
		PodRestart bool `yaml:"podrestart"`
	} `yaml:"kubernetes"`
	Slack struct {
		WebhookUrl string `yaml:"webhookurl"`
		Notify     bool   `yaml:"notify"`
		Channel    string `yaml:"channel"`
		UserName   string `yaml:"username"`
	} `yaml:"slack"`
	Logging struct {
		Debug bool `yaml:"debug"`
	} `yaml:"logging"`
}

var (
	configFile Configuration
)

func ReadFile(configPath string) Configuration {
	var file *os.File
	var err error

	file, err = os.Open(configPath)
	if err != nil {
		log.New(os.Stdout, "ERROR ", log.Ldate|log.Ltime).Println(err)
		os.Exit(1)
	}

	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&configFile)

	if err != nil {
		log.New(os.Stdout, "ERROR ", log.Ldate|log.Ltime).Println(err)
		os.Exit(1)
	}

	return configFile
}
