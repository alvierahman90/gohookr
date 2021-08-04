# gohookr

A _really_ simple webhook receiver, which listens at `/webhooks/<webhook-name>`.

Default config path is `/etc/gohookr.conf` and can be overriden by setting environment variable
`CONFIG`.

Check below for an example configuration.

## Installation

After you [install go](https://golang.org/doc/install):

```
make
```

## Signature Verification

Signature verificaiton is done using SHA256 HMACs.
You **must** set which HTTP header gohookr will receive a signature from using the `SignatureHeader`
key for each service.
You should also specify a shared secret in the `Secret` key.

You may also need to specify a `SignaturePrefix`.
For GitHub it would be `sha256=`.

### Disable Signature Verification

You can disable signature verification altogether by setting environment variable
`NO_SIGNATURE_VERIFICATION` to `true`.

## Writing Commands

gohookr doesn't care what the command is as long as the `Program` is executable.
You can specify extra arguments with the `Arguments` field.
You can ask it to put the payload as the last argument by setting `AppendPayload` to true.

## Writing Tests

gohookr can run test before running your script.
Tests must be in the form of bash scripts.
A non-zero return code is considered a fail and gohookr will run no further tests and will not
deploy.

Tests are run in the order they're listed so any actions that need to be done before
tests are run can simply be put in this section before the tests.

## Example Config

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
          "AppendPayload": true
      },
      "Secret": "THISISVERYSECRET",
      "SignatureHeader": "X-Hub-Signature",
      "SignaturePrefix": "sha256=",
      "Tests": [
        {
          "Program": "echo",
          "Arguments": [ "test" ]
        }
      ]
    }
  }
}
```
