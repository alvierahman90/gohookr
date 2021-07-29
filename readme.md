# gohookr

A _really_ simple webhook receiver.

Check config.json for an example configuration.

Default config path is `/etc/ghookr.conf`, can be overriden with `CONFIG` environment variable.

## Signature Verification

Signature verificaiton is done using SHA256 HMACs.
You **must** set which HTTP header gohookr will receive a signature from using the `SignatureHeader`
key for each service.
You should also specify a shared secret in the `Secret` key.

### Disable Signature Verification

You can disable signature verification altogether by setting environment variable `NO_SIGNATURE_VERIFICATION`
to `true`.

## Tests

gohookr can run test before running your script.
Tests must be in the form of bash scripts.
A non-zero return code is considered a fail and gohookr will run no further tests and will not
deploy.

Tests are run in the order they're listed so any actions that need to be done before
tests are run can simply be put before the tests.

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
