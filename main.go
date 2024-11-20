package main

import (
  "os"
	"log"
	"github.com/ryanvillarreal/krouter/pkg/cmd"
)

func main() {
   // root needed to start dhcp/dns server and modify ifaces
   if os.Geteuid() != 0 {
    log.Fatal("This program must be run as root/administrator")
   }
	 if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	 }
}

