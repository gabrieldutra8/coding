package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/hashicorp/consul/api"
)

const serviceName = "example-service"
const servicePort = 8080

func main() {
	consulClientConfig := api.DefaultConfig()
	consulClient, err := api.NewClient(consulClientConfig)
	if err != nil {
		log.Fatal(err)
	}

	registerService(consulClient, serviceName, servicePort)
	defer deregisterService(consulClient, serviceName)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("health"))
	})

	go func() {
		http.ListenAndServe(fmt.Sprintf(":%d", servicePort), nil)
	}()
}

func deregisterService(client *api.Client, serviceName string) error {
	return client.Agent().ServiceDeregister(serviceName)
}

func registerService(client *api.Client, serviceName string, servicePort int) error {
	registration := &api.AgentServiceRegistration{
		Name:    serviceName,
		ID:      serviceName,
		Port:    servicePort,
		Address: getOutBoundIP(),
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d/health", getOutBoundIP(), servicePort),
			Interval: "30s",
		},
	}

	return client.Agent().ServiceRegister(registration)
}

func getOutBoundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	addr := conn.LocalAddr().(*net.UDPAddr)
	return addr.IP.String()
}
