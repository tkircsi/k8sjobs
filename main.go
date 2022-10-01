package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	jsonserializer "k8s.io/apimachinery/pkg/runtime/serializer/json"
)

func main() {
	jobName := flag.String("jobname", "test-job", "Name of the Job")
	containerImage := flag.String("image", "ubuntu:latest", "Name of the container image")
	entryCommand := flag.String("command", "ls", "The command to run inside the container")
	kubeConfig := flag.String("config", "config", "The kubeconfig file")

	flag.Parse()

	log.Printf("Args: %s %s %s %s\n", *jobName, *containerImage, *entryCommand, *kubeConfig)

	clientset := connectToK8s(*kubeConfig)
	launchK8sJob(clientset, jobName, containerImage, entryCommand)
	log.Println("job successfully created")
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

func launchK8sJob(clientset *kubernetes.Clientset, jobName, image, cmd *string) {
	jobs := clientset.BatchV1().Jobs("default")
	var backOffLimit int32 = 0

	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      *jobName,
			Namespace: "default",
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:    *jobName,
							Image:   *image,
							Command: strings.Split(*cmd, " "),
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
			BackoffLimit: &backOffLimit,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	job, err := jobs.Create(ctx, jobSpec, metav1.CreateOptions{})
	if err != nil {

		log.Fatalf("failed to create job: %q\n", *jobName)
	}
	cancel()

	// Serializer = Decoder + Encoder.
	serializer := jsonserializer.NewSerializerWithOptions(
		nil,
		nil,
		nil,
		jsonserializer.SerializerOptions{
			Yaml:   true,
			Pretty: false,
			Strict: false,
		},
	)

	yaml, err := runtime.Encode(serializer, job)
	if err != nil {
		log.Fatalln("can not serialize job into yaml")
	}

	log.Println(string(yaml))
}
