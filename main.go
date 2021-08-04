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

	"git.alra.uk/alvierahman90/gohookr/config"
	"github.com/gorilla/mux"
)

var config_filename = "/etc/gohookr.json"
var checkSignature = true
var c config.Config

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/webhooks/{service}", webhookHandler)

	if p, ok := os.LookupEnv("CONFIG"); ok {
		config_filename = p
	}

	if p, ok := os.LookupEnv("NO_SIGNATURE_CHECK"); ok {
		checkSignature = p != "true"
	}

	raw_config, err := ioutil.ReadFile(config_filename)
	if err != nil {
		panic(err.Error())
	}

	json.Unmarshal(raw_config, &c)
	if err := c.Validate(); err != nil {
		panic(err.Error())
	}

	log.Fatal(http.ListenAndServe(c.ListenAddress, r))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	payload := ""
	if p, err := ioutil.ReadAll(r.Body); err != nil {
		writeResponse(w, 500, "Internal Server Error: Could not read payload")
		return
	} else {
		payload = string(p)
	}

	// check what service is specified in URL (/webhooks/{service}) and if it exists
	service, ok := c.Services[string(mux.Vars(r)["service"])]
	if !ok {
		writeResponse(w, 404, "Service Not Found")
		return
	}

	// Verify that signature provided matches signature calculated using secretsss
	signature := r.Header.Get(service.SignatureHeader)
	calculatedSignature := getSha256HMACSignature([]byte(service.Secret), payload)
	fmt.Printf("signature = %v\n", signature)
	fmt.Printf("calcuatedSignature = %v\n", signature)
	if signature != calculatedSignature && checkSignature {
		writeResponse(w, 400, "Bad Request: Signatures do not match")
		return
	}

	// Run tests, immediately stop if one fails
	for _, test := range service.Tests {
		if _, err := test.Execute(payload); err != nil {
			writeResponse(w, 409,
				fmt.Sprintf("Conflict: Test failed: %v", err.Error()),
			)
			return
		}
	}

	if stdout, err := service.Script.Execute(payload); err != nil {
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
