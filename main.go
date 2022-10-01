package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	jobName := flag.String("jobname", "test-job", "Name of the Job")
	containerImage := flag.String("image", "ubuntu:latest", "Name of the container image")
	entryCommand := flag.String("command", "ls", "The command to run inside the container")
	kubeConfig := flag.String("config", "config", "The kubeconfig file")

	flag.Parse()

	fmt.Printf("Args: %s %s %s %s", *jobName, *containerImage, *entryCommand, *kubeConfig)

	clientset := connectToK8s(*kubeConfig)
	fmt.Printf("%v\n", clientset)
}

func connectToK8s(kubeConfig string) *kubernetes.Clientset {
	home, exists := os.LookupEnv("HOME")
	if !exists {
		home = "/root"
	}

	configPath := filepath.Join(home, ".kube", kubeConfig)
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		log.Panicln("failed to create config")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Panicln("failed to create clientset")
	}

	return clientset
}
