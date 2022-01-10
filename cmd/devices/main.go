package main

import (
	"fmt"
	"go.bug.st/serial/enumerator"
)

func main() {
	var err error

	var ports []*enumerator.PortDetails
	ports, err = enumerator.GetDetailedPortsList()
	if err != nil {
		return
	}

	for _, port := range ports {
		fmt.Printf("Name=%#v IsUSB=%#v VID=%#v PID=%#v Product=%#v SerialNumber=%#v\n", port.Name, port.IsUSB, port.VID, port.PID, port.Product, port.SerialNumber)
	}
}
