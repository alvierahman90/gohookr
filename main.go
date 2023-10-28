package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"git.alv.cx/alvierahman90/gohookr/config"
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

	raw_config, err := os.ReadFile(config_filename)
	if err != nil {
		panic(err.Error())
	}

	if err := json.Unmarshal(raw_config, &c); err != nil {
		panic(err.Error())
	}

	if err := c.Validate(); err != nil {
		panic(err.Error())
	}

	log.Fatal(http.ListenAndServe(c.ListenAddress, r))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	// Check what service is specified in URL (/webhooks/{service}) and if it exists
	serviceName := string(mux.Vars(r)["service"])
	service, ok := c.Services[serviceName]
	if !ok {
		writeResponse(w, 404, "Service Not Found")
		fmt.Printf("Service not found: %v\n", serviceName)
		return
	}
	fmt.Printf("Got webhook for: %v\n", serviceName)

	// Read payload or return 500 if that doesn't work out
	payload := ""
	if p, err := io.ReadAll(r.Body); err != nil {
		writeResponse(w, 500, "Internal Server Error: Could not read payload")
		fmt.Println("Error: Could not read payload")
		return
	} else {
		payload = string(p)
	}

	// Verify that signature provided matches signature calculated using secretsss
	signature := r.Header.Get(service.SignatureHeader)
	calculatedSignature := fmt.Sprintf(
		"%v%v",
		service.SignaturePrefix,
		getSha256HMACSignature([]byte(service.Secret), payload),
	)
	fmt.Printf("signature          = %v\n", signature)
	fmt.Printf("calcuatedSignature = %v\n", calculatedSignature)
	if checkSignature && !service.DisableSignatureVerification && signature != calculatedSignature {
		writeResponse(w, 400, "Bad Request: Signatures do not match")
		fmt.Println("Signatures do not match!")
		return
	}

	// Run tests and script as goroutine to prevent timing out
	go func(){
		// Run tests, immediately stop if one fails
		for _, test := range service.Tests {
			if _, err := test.Execute(payload); err != nil {
				fmt.Printf("Test failed(%v) for service %v\n", test, serviceName)
				return
			}
		}
		stdout, err := service.Script.Execute(payload)
		fmt.Println(string(stdout))
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	writeResponse(w, 200, "OK")
	return
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
