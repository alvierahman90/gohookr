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

var config_filename = "/etc/gohookr.json"
var checkSignature = true

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/webhooks/{service}", webhookHandler)

	port := ":80"
	if p, ok := os.LookupEnv("PORT"); ok {
		port = fmt.Sprintf(":%v", p)
	}

	if p, ok := os.LookupEnv("CONFIG"); ok {
		config_filename = p
	}

	if p, ok := os.LookupEnv("NO_SIGNATURE_CHECK"); ok {
		checkSignature = p != "true"
	}

	log.Fatal(http.ListenAndServe(port, r))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
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

	// check what service is specified in URL (/webhooks/{service}) and if it exists
	service, ok := config.Services[string(mux.Vars(r)["service"])]
	if !ok {
		writeResponse(w, 404, "Service Not Found")
		return
	}

	// Verify that signature provided matches signature calculated using secretsss
	signature := r.Header.Get(service.SignatureHeader)
	calculatedSignature := getSha256HMACSignature([]byte(service.Secret), payload)
	fmt.Printf("signature = %v\n", signature)
	fmt.Printf("calcuatedSignature = %v\n", signature)
	if signature != calculatedSignature  && checkSignature{
		writeResponse(w, 400, "Bad Request: Signatures do not match")
		return
	}

	// Run tests, immediately stop if one fails
	for _, test := range service.Tests {
		if _, err := exec.Command(test.Command, test.Arguments...).Output(); err != nil {
			writeResponse(w, 409,
				fmt.Sprintf("409 Conflict: Test failed: %v", err.Error()),
			)
			return
		}
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

type Test struct {
	Command string
	Arguments []string
}

type Service struct {
	Gitea           bool
	Script          string
	Secret          string
	SignatureHeader string
	Tests           []Test
}

type Config struct {
	Services map[string]Service
}
