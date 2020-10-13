package provision

import (
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func UpdateProEndpoints(destAddr string) {

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	namespace := "default"
	endpointClient := clientset.CoreV1().Endpoints(namespace)

	endpoints, getError := endpointClient.Get("provision", metav1.GetOptions{})
	if getError != nil {
		mt.Println("Failed to get %s endpoints error: %s", namespace, getError)
	} 

	if getError == nil {
		for index, item := range endpoints.Subsets {
			newAddreses := []v1.EndpointAddress{}
			var address v1.EndpointAddress
			address.IP = destAddr
			fmt.Printf("Added endpoint address: %s\n", destAddr)
			newAddreses = append(newAddreses, address)
			//item.Addresses = append(item.Addresses, address)
			item.Addresses = append(newAddreses, item.Addresses...)
			endpoints.Subsets[index].Addresses = item.Addresses
		}
		/* Update */
		_, updateError := endpointClient.Update(endpoints)
		if updateError != nil {
			fmt.Println("Failed to update %s endpoints error: %s", namespace, updateError)
		}
		time.Sleep(1 * time.Second)
	} else {
		/* create Adresses */
		newAddreses := []v1.EndpointAddress{}
		var address v1.EndpointAddress
		address.IP = destAddr
		newAddreses = append(newAddreses, address)
		/* create Port */
		newPort := []v1.EndpointPort{}
		var port1, port2 v1.EndpointPort
		port1.Port = 80
		port1.Protocol = "TCP"
		port1.Name = "tcp-port1"
		port2.Port = 8000
		port2.Protocol = "TCP"
		port2.Name = "tcp-port2"
		newPort = append(newPort, port1, port2)

		/* create subset */
		newSubset := []v1.EndpointSubset{}
		var subset v1.EndpointSubset
		subset.Addresses = newAddreses
		subset.Ports = newPort
		newSubset = append(newSubset, subset)
		/* create endpoint */
		var endpoint v1.Endpoints
		endpoint.Subsets = newSubset
		endpoint.ObjectMeta.Name = "provision"

		/* Create client endpoints */
		_, createError := endpointClient.Create(&endpoint)
		if createError != nil {
			fmt.Println("Failed to create %s endpoints error: %s", namespace, createError)
		} else {
			fmt.Println("Success to create %s endpoints", namespace)
		}
		time.Sleep(1 * time.Second)
	}
	endpoints, _ = endpointClient.Get("provision", metav1.GetOptions{})

	for i, item := range endpoints.Subsets {
		for j, item2 := range item.Addresses {
			fmt.Printf("%d-%d Updated Name: %s, address: %s \n", i, j, namespace, item2.IP)
		}
	}
}

func RemoveProEndpoints(destAddr string) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	namespace := "default"
	endpointClient := clientset.CoreV1().Endpoints(namespace)

	endpoints, getError := endpointClient.Get("provision", metav1.GetOptions{})
	if getError != nil {
		fmt.Println("Failed to get %s endpoints error: %s", namespace, getError)
	} else {
		for i, item := range endpoints.Subsets {
			if len(item.Addresses) > 1 {
				for j := 0; j < len(item.Addresses); j++ {
					if item.Addresses[j].IP == destAddr {
						item.Addresses = append(item.Addresses[:j], item.Addresses[j+1:]...)
						j-- // -1 as the slice just go
					}
				}
				endpoints.Subsets[i].Addresses = item.Addresses
				fmt.Println( "Remove %s space %s endpoint", namespace, destAddr)
				/* Update */
				_, updateError := endpointClient.Update(endpoints)
				if updateError != nil {
					fmt.Println("Failed to update %s endpoints error: %s", namespace, updateError)
				}
				endpoints, _ = endpointClient.Get("provision", metav1.GetOptions{})
				for i, item := range endpoints.Subsets {
					for j, item2 := range item.Addresses {
						fmt.Printf("%d-%d Updated Name: %s, address: %s \n", i, j, namespace, item2.IP)
					}
				}

			} else if len(item.Addresses) == 1 {
				if item.Addresses[0].IP == destAddr {
					tlog.L(tlog.Prov, "Remove space %s only one endpoint %s", namespace, destAddr)
					/* Delete */
					deleteError := endpointClient.Delete("provision", &metav1.DeleteOptions{})
					if deleteError != nil {
						fmt.Println("Failed to delete %s endpoints error: %s", namespace, deleteError)
					}
				}
			}
		}

	}
}

