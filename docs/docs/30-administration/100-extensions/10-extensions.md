# Extensions

Woodpecker also you to replace internal logic with external extensions by using pre-defined http endpoints. For example you can preprocess pipeline configurations by letting Woodpecker call an external service you developed.

## Security

To prevent your extensions service from leaking data to other system rather than the Woodpecker server, Woodpecker is signing http-request based on [http signatures](https://tools.ietf.org/html/draft-cavage-http-signatures). Woodpecker therefore generates a public-private ed25519 key pair at the first server start. To verify the request send to your extension webservice you have to verify the signature using the Woodpecker public key with some library like [httpsig](https://github.com/go-fed/httpsig). You can get the public Woodpecker key by opening `http://my-woodpecker.tld/api/service-key`. For a reference implementation of a configuration webservice checkout our [example-config-service](https://github.com/woodpecker-ci/example-config-service) repository.

using a private key generated on the first start of the woodpecker server. You can get the public key for verification by opening you UI and going to ... This way the external api can verify the authenticity request from the Woodpecker instance.
