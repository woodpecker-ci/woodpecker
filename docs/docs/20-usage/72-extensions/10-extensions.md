# Extensions

Woodpecker allows you to replace internal logic with external extensions by using pre-defined http endpoints. For example you can preprocess pipeline configurations by letting Woodpecker call an external service you developed.

## Security

To prevent your extensions service from leaking data to other systems rather than the Woodpecker server, Woodpecker is signing http-requests using [http signatures](https://tools.ietf.org/html/draft-cavage-http-signatures). Woodpecker therefore generates a public-private ed25519 key pair at the first server start. To verify the request send to your extension webservice you have to verify the signature using the Woodpecker public key with some library like [httpsig](https://github.com/go-fed/httpsig). You can get the public Woodpecker key by opening `http://my-woodpecker.tld/api/signature/public-key` or by visiting the Woodpecker UI, going to you repo settings and opening the extensions page.
