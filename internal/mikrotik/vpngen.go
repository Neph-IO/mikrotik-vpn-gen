package mikrotik

import (
	"archive/zip"

	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/Neph-IO/mikrotik-vpn-gen/internal/config"
	"github.com/Neph-IO/mikrotik-vpn-gen/internal/sftp"
)

func MakeVpn(nom string, prenom string, password string, profile string) (string, error) {
	if nom == "" || prenom == "" || password == "" || profile == "" {
		return "", fmt.Errorf("[VPN]error in makevpn args \n")
	}
	nom = normalizeName(nom)
	prenom = normalizeName(prenom)
	pshortname, ok := config.Conf.Vpncreator.ProfileMap[profile]
	if !ok {
		return "", fmt.Errorf("[VPN]'%s' Profile unknown", profile)
	}
	pshortname = strings.ToUpper(pshortname)
	certName := formatNames(pshortname, nom, prenom)

	//Si le certificat existe on ne recréé pas juste in réexporte les certificat avec le nouveau password
	res := RunMk("/certificate/print")
	for _, cert := range res.Re {
		if cert.Map["name"] == certName {
			fmt.Println("Le certificat existe")
			certID := cert.Map[".id"]
			exportCert(certID, password, certName)
			downloadCerts(certName)
			createOpenvpnFile(certName)
			zipOpenVPNConfig(certName)
			moveToNginx(certName)
			cleanTempFolder(certName)
			return certName, nil
		}
	}
	//Liste des instruction fragmentées en fonctions
	makeSecret(certName, password, profile)
	certid := makeCert(certName)
	signCert(certid)
	exportCert(certid, password, certName)
	downloadCerts(certName)
	createOpenvpnFile(certName)
	zipOpenVPNConfig(certName)
	moveToNginx(certName)
	cleanTempFolder(certName)
	return certName, nil
}

func normalizeName(s string) string {
	if s == "" {
		return s
	}
	r := []rune(strings.ToLower(s))
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// Here is the logic for naming
func formatNames(initialeSct string, nom string, prenom string) string {
	rPrenom := []rune(prenom)
	if len(rPrenom) == 0 {
		return ""
	}
	// Name in lowercase, initals and lastname in caps
	return fmt.Sprintf("%s%s%s", initialeSct, nom, strings.ToUpper(string(rPrenom[0])))
}

// Cree le secret sur le MK  ppp/secret
func makeSecret(certName string, password string, profile string) {
	args := []string{
		fmt.Sprintf("=name=%s", certName),
		fmt.Sprintf("=password=%s", password),
		"=service=ovpn",
		fmt.Sprintf("=profile=%s", profile),
	}
	fmt.Printf("/ppp/secret/add %v\n", args)
	RunMk("/ppp/secret/add", args...)
}

func makeCert(certName string) string {
	args := []string{
		fmt.Sprintf("=name=%s", certName),
		fmt.Sprintf("=common-name=%s", certName),
		fmt.Sprintf("=country=%s", config.Conf.Vpncreator.Countrycode),
		fmt.Sprintf("=days-valid=%s", config.Conf.Vpncreator.ValidTime),
		fmt.Sprintf("=key-size=%s", config.Conf.Vpncreator.Keysize),
		fmt.Sprintf("=key-usage=tls-client"),
	}
	fmt.Println("[VPN] Creation du cert")
	RunMk("/certificate/add", args...)
	res := RunMk("/certificate/print", "?name="+certName)
	return res.Re[0].Map[".id"]

}

func signCert(certID string) {
	if certID == "" {
		fmt.Println("No Cert ID given to sign")
	}
	args := []string{
		fmt.Sprintf("=.id=%s", certID),
		fmt.Sprintf("=ca=%s", config.Conf.Vpncreator.CaName),
	}
	fmt.Println("[VPN] Certificate signing")
	RunMk("/certificate/sign", args...)
}

func exportCert(certID string, password string, certname string) {
	fmt.Printf("[VPN] Certificate export %s\n", certID)
	args := []string{
		fmt.Sprintf("=.id=%s", certID),
		fmt.Sprintf("=export-passphrase=%s", password),
		fmt.Sprintf("=file-name=%s", certname),
	}
	//fmt.Println(RunMk("/certificate/print", "?name="+certname))
	RunMk("/certificate/export-certificate", args...)
}

func downloadCerts(certname string) (certfile string, keyfile string) {
	err := sftp.DownloadFileFromMK(certname+".crt", "temp/"+certname+".crt")
	if err != nil {
		fmt.Print("[SFTP]", err)
	}
	err = sftp.DownloadFileFromMK(certname+".key", "temp/"+certname+".key")
	if err != nil {
		fmt.Print("[SFTP]", err)
	}
	return certname + ".crt", certname + ".key"
}

func createOpenvpnFile(certname string) error {
	data, err := os.ReadFile("template/Client.ovpn")
	if err != nil {
		return err
	}
	content := string(data)

	//Replacing the content of the .ovpn file by the ones created by the backend
	content = strings.ReplaceAll(content, "cert cert.crt", fmt.Sprintf("cert %s.crt", certname))
	content = strings.ReplaceAll(content, "key key.key", fmt.Sprintf("key %s.key", certname))
	err = os.MkdirAll("temp/"+certname, 0755)
	if err != nil {
		return err
	}
	ovpnPath := filepath.Join("temp/"+certname, "Client.ovpn")

	err = os.Rename("temp/"+certname+".crt", "temp/"+certname+"/"+certname+".crt")
	if err != nil {
		return err
	}
	err = os.Rename("temp/"+certname+".key", "temp/"+certname+"/"+certname+".key")
	if err != nil {
		return err
	}
	caFilename := config.Conf.Vpncreator.CaName + ".crt"
	caPath := filepath.Join("template", caFilename)

	input, err := os.ReadFile(caPath)
	if err != nil {
		return fmt.Errorf("[VPN]CA reading failed: %w", err)
	}

	outputPath := filepath.Join("temp", certname, caFilename)
	err = os.WriteFile(outputPath, input, 0644)
	if err != nil {
		return fmt.Errorf("[VPN]CA writing failed: %w", err)
	}

	err = os.WriteFile(ovpnPath, []byte(content), 0644)
	if err != nil {
		return err
	}

	secretPath := filepath.Join("temp", certname, "secret")
	secretContent := fmt.Sprintf("%s\n", certname)

	err = os.WriteFile(secretPath, []byte(secretContent), 0644)
	if err != nil {
		return err
	}
	fmt.Println("[VPN]Files Created")

	return nil
}

func zipOpenVPNConfig(certname string) error {
	sourceDir := filepath.Join("temp", certname)
	zipPath := filepath.Join("temp", certname+".zip")

	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("[VPN]Zip failed : %w", err)
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil //only files
		}

		// add relative file path
		relPath, err := filepath.Rel("temp", path) //important : base = temp/
		if err != nil {
			return err
		}

		dest, err := archive.Create(relPath)
		if err != nil {
			return err
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		_, err = io.Copy(dest, srcFile)
		return err
	})

	if err != nil {
		return fmt.Errorf("[VPN]Error during zip : %w", err)
	}

	fmt.Println("[VPN]Config file successfully ziped")
	return nil
}

func moveToNginx(certName string) {
	err := os.Rename("temp/"+certName+".zip", config.Conf.Vpncreator.Nginxfolder+certName+".zip")
	if err != nil {
		log.Fatal(err)
	}
}

func cleanTempFolder(certName string) {
	tempFolder := filepath.Join("temp", certName)

	err := os.RemoveAll(tempFolder)
	if err != nil {
		println(err)
	}
}

func DeleteGeneratedVPN(certname string) error {

	zipPath := filepath.Join(config.Conf.Vpncreator.Nginxfolder, certname)
	fmt.Println("[VPN]Trying to delete :", zipPath)

	err := os.Remove(zipPath)
	if err != nil {
		return fmt.Errorf("[VPN]Can't delete %s : %w", certname, err)
	}

	fmt.Println("[VPN]Deleted :", zipPath)
	return nil
}
