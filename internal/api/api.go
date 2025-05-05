package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Neph-IO/mikrotik-vpn-gen/internal/config"
	"github.com/Neph-IO/mikrotik-vpn-gen/internal/mikrotik"
)

func CreateVPNHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Reçu requête %s sur %s\n", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		http.Error(w, "Wrong Method", http.StatusMethodNotAllowed)
		return
	}

	type RequestData struct {
		Nom      string `json:"nom"`
		Prenom   string `json:"prenom"`
		Profile  string `json:"profile"`
		Password string `json:"password"`
	}

	var data RequestData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	certName, err := mikrotik.MakeVpn(data.Nom, data.Prenom, data.Password, data.Profile)
	if err != nil {
		fmt.Print(err)
		return
	}

	url := fmt.Sprintf("http://%s/vpn/%s.zip", r.Host, certName) //Should i make it https ?
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": url})
}

func DeleteVPNHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	type RequestData struct {
		Certname string `json:"certname"`
	}

	var data RequestData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil || data.Certname == "" {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	err = mikrotik.DeleteGeneratedVPN(data.Certname)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// a little strict rule for API calls.. just basics
func SecureOrigin(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allowedOrigin := config.Conf.GlobalConf.AllowedOrigin

		origin := r.Header.Get("Origin")
		referer := r.Header.Get("Referer")

		if origin != "" && origin != allowedOrigin {
			http.Error(w, "Forbidden - invalid origin", http.StatusForbidden)
			return
		}

		if referer != "" && !strings.HasPrefix(referer, allowedOrigin) {
			http.Error(w, "Forbidden - invalid referer", http.StatusForbidden)
			return
		}

		if origin == "" && referer == "" {
			http.Error(w, "Forbidden - missing origin/referer", http.StatusForbidden)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		handler(w, r)
	}
}
