# gohookr

A _really_ simple webhook receiver, which listens at `/webhooks/<webhook-name>`.

## Installation

After you [install go](https://golang.org/doc/install):

```
make
```

## Configuration

Default config path is `/etc/gohookr.json`.
It can be overriden by setting environment variable `CONFIG`.

Check below for an example configuration, which should tell you most of the things you need to know
to configure gohookr.

Currently gohookr must be restarted after config changes.

### Signature Verification

Signature verificaiton is done using SHA256 HMACs.
You **must** set which HTTP header gohookr will receive a signature from using the `SignatureHeader`
key for each service.
You should also specify a shared secret in the `Secret` key.

You may also need to specify a `SignaturePrefix`.
For GitHub it would be `sha256=`.

#### Disable Signature Verification

You can disable signature verification by setting `DisableSignatureVerification` for a service to `true`.

You can disable signature verification for all services by setting environment variable
`NO_SIGNATURE_VERIFICATION` to `true`.

### Writing Commands

gohookr doesn't care what the command is as long as the `Program` is executable.
You can specify extra arguments with the `Arguments` field.
You can ask it to put the payload as the last (or second to last if `AppendHeaders` is set) argument by setting `AppendPayload` to true.
You can ask it to put the request headers as the last argument by setting `AppendHeaders` to true.

### Writing Tests

gohookr can run test before running your script.
Tests must be in the form of bash scripts.
A non-zero return code is considered a fail and gohookr will run no further tests and will not
deploy.

Tests are run in the order they're listed so any actions that need to be done before
tests are run can simply be put in this section before the tests.

### Example Config

Required config keys are `ListenAddress` and `Services`.

Requried keys per service are `Script.Program`, `Secret`, `SignatureHeader`.

An example config file can be found [here](./config.json) but also below:

```json
{
  "ListenAddress": "127.0.0.1:8654",
  "Services": {
    "test": {
      "Script": {
          "Program": "./example.sh",
          "AppendPayload": true,
          "AppendHeaders": true
      },
      "DisableSignatureVerification": true,
      "Tests": [
        {
          "Program": "echo",
          "Arguments": [ "test" ]
        }
      ]
    },
    "fails_tests": {
      "Script": {
          "Program": "./example.sh",
          "AppendPayload": true
      },
      "Secret": "who_cares",
      "SignatureHeader": "X-Hub-Signature-256",
      "SignaturePrefix": "sha256=",
      "Tests": [
        {
          "Program": "false",
          "Arguments": []
        }
      ]
    }
  }
}
```
