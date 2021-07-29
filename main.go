package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/gorilla/mux"
)

var config_filename = "/etc/ghookr.json"
var noSignatureCheck = false

func main() {
	// Used for testing purposes... generates hmac string
	if os.Getenv("HMACGEN") == "true" {
		input, err := ioutil.ReadAll(os.Stdin)
		secret := os.Getenv("SECRET")
		if err != nil {
			panic(err.Error())
		}
		fmt.Println(getSha256HMACSignature([]byte(secret), string(input)))
		return
	}
	r := mux.NewRouter()
	r.HandleFunc("/webhook/{service}", webhook)

	port := ":80"
	if p, ok := os.LookupEnv("PORT"); ok {
		port = fmt.Sprintf(":%v", p)
	}

	if p, ok := os.LookupEnv("CONFIG"); ok {
		config_filename = p
	}

	if p, ok := os.LookupEnv("NO_SIGNATURE_CHECK"); ok {
		noSignatureCheck = p == "true"
	}

	log.Fatal(http.ListenAndServe(port, r))
}

func webhook(w http.ResponseWriter, r *http.Request) {
	// TODO run any specified tests before running script

	service_name := mux.Vars(r)["service"]

	payload := ""
	if p, err := ioutil.ReadAll(r.Body); err != nil {
		writeResponse(w, 500, "Internal Server Error: Could not read payload")
		return
	} else {
		payload = string(p)
	}

	raw_config, err := ioutil.ReadFile(config_filename)
	if err != nil {
		writeResponse(w, 500, "Internal Server Error: Could not open config file")
		return
	}
	config := Config{}
	json.Unmarshal(raw_config, &config)

	var service = Service{}
	if val, ok := config.Services[string(service_name)]; !ok {
		writeResponse(w, 404, "Service Not Found")
		return
	} else {
		service = val
	}

	// Verify that signature provided matches signature calculated using secretsss
	signature := r.Header.Get(service.SignatureHeader)
	calculatedSignature := getSha256HMACSignature([]byte(service.Secret), payload)
	if noSignatureCheck || signature == calculatedSignature {
		writeResponse(w, 400, "Bad Request: Signatures do not match")
		return
	}

	if stdout, err := exec.Command(service.Script, payload).Output(); err != nil {
		writeResponse(w, 500, err.Error())
		return
	} else {
		writeResponse(w, 200, string(stdout))
		return
	}
}

func writeResponse(w http.ResponseWriter, responseCode int, responseString string) {
	w.WriteHeader(responseCode)
	w.Write([]byte(fmt.Sprintf("%v %v", responseCode, responseString)))
}

func getSha256HMACSignature(secret []byte, data string) string {
	h := hmac.New(sha256.New, secret)
	io.WriteString(h, data)
	return hex.EncodeToString(h.Sum(nil))
}

type Service struct {
	Gitea           bool
	Script          string
	Secret          string
	SignatureHeader string
	Tests           []string
}

type Config struct {
	Services map[string]Service
}
