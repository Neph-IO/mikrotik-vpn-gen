package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Neph-IO/mikrotik-vpn-gen/internal/api"
	"github.com/Neph-IO/mikrotik-vpn-gen/internal/config"
	"github.com/Neph-IO/mikrotik-vpn-gen/internal/mikrotik"
)

func main() {
	checkTempFolder()
	//Conf Loading
	fmt.Println("[CONF]Loading Config")
	err := config.Load("config.yaml")
	if err != nil {
		panic(fmt.Sprintf("[CONF]%v", err))
	}
	fmt.Println("[CONF]Succesfuly loaded")

	//Test Mk communication
	resp := mikrotik.RunMk("/system/identity/print")
	for _, re := range resp.Re {
		for key, val := range re.Map {
			fmt.Printf("[MIKROTIK]%s: %s\n", key, val)
		}
	}

	//API HANDLER DEFINITION
	http.HandleFunc("/api/createvpn", api.SecureOrigin(api.CreateVPNHandler))
	http.HandleFunc("/api/deletevpn", api.SecureOrigin(api.DeleteVPNHandler))
	//Lancement du serveur
	fmt.Printf("[API]Listening on port  %s\n", config.Conf.GlobalConf.ApiPort)
	err = http.ListenAndServe(":"+config.Conf.GlobalConf.ApiPort, nil)
	if err != nil {
		panic(err)
	}
}

func checkTempFolder() {
	err := os.MkdirAll("temp", 0755)
	if err != nil {
		fmt.Printf("[CONF]Can't create temp/ folder: %v \n", err)
	}
	fmt.Println("[CONF]Temp/ folder ready")
}
