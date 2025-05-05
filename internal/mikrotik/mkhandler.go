package mikrotik

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/Neph-IO/mikrotik-vpn-gen/internal/config"
	"github.com/go-routeros/routeros/v3"
)

func initMK() (*routeros.Client, error) {
	cfg := config.Conf.Mikrotik

	if cfg.Tls {
		caCert, _ := os.ReadFile(cfg.CertName)
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(caCert)
		tlsCfg := &tls.Config{
			RootCAs: certPool,
		}

		return routeros.DialTLS(cfg.Address+":"+cfg.Port, cfg.Username, cfg.Password, tlsCfg)
	}
	return routeros.Dial(cfg.Address+":"+cfg.Port, cfg.Username, cfg.Password)
}

func RunMk(command string, args ...string) *routeros.Reply {
	client, err := initMK()
	if err != nil {
		fmt.Println(err)
	}
	defer client.Close()
	onlyOneString := append([]string{command}, args...)
	reponse, err := client.Run(onlyOneString...)
	if err != nil {
		fmt.Println(err)
	}
	return reponse
}
