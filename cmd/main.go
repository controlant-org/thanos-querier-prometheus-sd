package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	// Setup our CLI arguments.
	flag.String("output-file", "/tmp/tqsd/result.yaml", "The location of the output file.")
	flag.Int("interval", 10000, "The number of milliseconds to wait between discovery cycles.")

	// Bring in any flags set on the command line.
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

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

		err = ioutil.WriteFile(viper.GetString("output-file"), result, 0644)
		if err != nil {
			log.Fatal(err)
		}

		// Wait for the next cycle.
		time.Sleep(time.Duration(viper.GetInt("interval")) * time.Millisecond)
	}
}
