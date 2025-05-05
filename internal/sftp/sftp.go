package sftp

import (
	"fmt"
	"io"
	"os"

	"github.com/Neph-IO/mikrotik-vpn-gen/internal/config"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func DownloadFileFromMK(source string, dest string) error {
	sshconfig := &ssh.ClientConfig{
		User: config.Conf.Mikrotik.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Conf.Mikrotik.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", config.Conf.Mikrotik.Address+":22", sshconfig)
	if err != nil {
		return err
	}
	defer conn.Close()
	fmt.Println("[SFTP]Connexion Ok")

	sftpClient, err := sftp.NewClient(conn)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	remotefile, err := sftpClient.Open(source)
	if err != nil {
		return err
	}
	defer remotefile.Close()

	localfile, err := os.Create(dest)
	if err != nil {
		return err
	}

	_, err = io.Copy(localfile, remotefile)
	if err != nil {
		return err
	}
	fmt.Printf("[SFTP]Fichier %s téléchargé \n", source)
	localfile.Close()
	return nil

}
