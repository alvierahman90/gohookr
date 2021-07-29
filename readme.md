# gohookr

A _really_ simple webhook receiver, which listens at `0.0.0.0:<port>/webhooks/<webhook-name>`.

Default port is 80 and can be overriden by setting the environment variable `PORT`.

Default config path is `/etc/gohookr.conf` and can be overriden by setting environment variable
`CONFIG`.

Check below for an example configuration.

## Installation

After you [install go](https://golang.org/doc/install):

```
go mod tidy
go build
cp gohookr /usr/local/bin/
cp gohookr.service /usr/lib/systemd/system/
systemctl daemon-reload
systemctl enable --now gohookr
```

## Signature Verification

Signature verificaiton is done using SHA256 HMACs.
You **must** set which HTTP header gohookr will receive a signature from using the `SignatureHeader`
key for each service.
You should also specify a shared secret in the `Secret` key.

### Disable Signature Verification

You can disable signature verification altogether by setting environment variable
`NO_SIGNATURE_VERIFICATION` to `true`.

## Tests

gohookr can run test before running your script.
Tests must be in the form of bash scripts.
A non-zero return code is considered a fail and gohookr will run no further tests and will not
deploy.

Tests are run in the order they're listed so any actions that need to be done before
real tests are run can simply be put before the tests.

## Example Config

An example config file can be found [here](./config.json) but also below:

```json
{
  "Services": {
    "test": {
      "Script": "./example.sh",
      "Secret": "THISISVERYSECRET",
      "SignatureHeader": "X-Gitea-Signature",
      "Tests": [
        {
          "Command": "git",
          "Arguments": [ "pull" ]
        }
      ]
    }
  }
}
```
