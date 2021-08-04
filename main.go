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
var config Config

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
	config = Config{}
	json.Unmarshal(raw_config, &config)
	if err := config.Validate(); err != nil {
		panic(err.Error())
	}


	log.Fatal(http.ListenAndServe(config.ListenAddress, r))
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

func (c Config) Validate() error {
	if c.ListenAddress == "" {
		return requiredFieldError{"ListenAddress", ""}
	}

	jsonbytes, _ := json.MarshalIndent(c, "", "  ")
	fmt.Println(string(jsonbytes))

	for serviceName, service := range c.Services {
		if service.Script.Program == "" {
			return requiredFieldError{"Script.Program", serviceName}
		}
		if service.SignatureHeader == "" {
			return requiredFieldError{"SignatureHeader", serviceName}
		}
		if service.Secret == "" {
			return requiredFieldError{"Secret", serviceName}
		}
	}

	return nil
}

func (c Command) Execute(payload string) ([]byte, error) {
	arguments := make([]string, 0)
	copy(c.Arguments, arguments)
	if c.AppendPayload {
		arguments = append(arguments, payload)
	}

	return exec.Command(c.Program, arguments...).Output()
}

type Command struct {
	Program       string
	Arguments     []string
	AppendPayload bool
}

type Service struct {
	Script          Command
	Secret          string
	SignatureHeader string
	Tests           []Command
}

type Config struct {
	ListenAddress string
	Services      map[string]Service
}

type requiredFieldError struct {
	fieldName   string
	serviceName string
}

func (e requiredFieldError) Error() string {
	return fmt.Sprintf("%v cannot be empty (%v)", e.fieldName, e.serviceName)
}
