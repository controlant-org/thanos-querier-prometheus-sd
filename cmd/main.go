package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	// Get our Kubernetes auth by assuming we're in a cluster and using our service account.
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Create our Kubernetes clientset.
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	for {
		// Get a list of all Prometheus instances managed by the Prometheus operator.
		services, err := clientset.CoreV1().Services("").List(context.TODO(), v1.ListOptions{
			// NOTE: this is hardcoded as there is currently no way to set annotations / labels on the service generated
			// 			 by the Prometheus operator.
			// TODO: Ideally, we'd just watch the Prometheus custom resource instances with specific annotation(s) and derive
			//			 our service discovery file from them, but this approach will do for now.
			LabelSelector: "operated-prometheus=true",
		})
		if err != nil {
			log.Fatal(err)
		}

		// Generate a map for the service discovery.
		discovery := []map[string]([]string){}
		for _, service := range services.Items {
			record := fmt.Sprintf("dnssrv+_grpc._tcp.%s.%s.svc.cluster.local", service.Name, service.Namespace)
			log.Printf("adding service: %s", record)
			discovery = append(discovery, map[string]([]string){"targets": {record}})
		}

		// Write out the map as YAML into a file.
		result, err := yaml.Marshal(&discovery)
		if err != nil {
			log.Fatal(err)
		}

		// TODO: configurable service discovery output file.
		err = ioutil.WriteFile("/tmp/sd/result.yaml", result, 0644)
		if err != nil {
			log.Fatal(err)
		}

		// Wait for the next cycle.
		time.Sleep(10 * time.Second)
	}
}
