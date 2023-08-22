# Secrets encryption

By default, Woodpecker does not encrypt secrets in its database. You can enable encryption 
using a simple AES key.

:::caution
Secrets encryption is experimental.
Currently encryption is unrevertable (do backups) 
and requires empty `secrets` table (can be evaluated in fresh installation or delete all secrets and create new after enabling encryption).

Check the [current state](https://github.com/woodpecker-ci/woodpecker/issues/1541)
:::

## Common

### Enabling secrets encryption

To enable secrets encryption set
`WOODPECKER_ENCRYPTION_KEY` or `WOODPECKER_ENCRYPTION_KEY_FILE` environment 
variable.

After encryption is enabled you will be unable to start Woodpecker server without providing valid encryption key!

## AES
Simple AES encryption.

### Configuration
You can manage encryption on server using these environment variables:
- `WOODPECKER_ENCRYPTION_KEY` - encryption key
- `WOODPECKER_ENCRYPTION_KEY_FILE` - file to read encryption key from

One option to generate encryption key is to use OpenSSL, but any password generator can also be used. Recommended key length is at least 32 bytes:
```shell
$ openssl rand -base64 32
GjVHT007c4x3N+YPbsZld+hifba1enXkOzIb/0h6oW8=
```

If we run the server with `WOODPECKER_ENCRYPTION_KEY='GjVHT007c4x3N+YPbsZld+hifba1enXkOzIb/0h6oW8='`, and try to create a secret like `some_secret:super-secret-value` 
then we'll get messages in the log similar to:
```log
{"level":"debug","id":0,"name":"some_secret","time":"2023-08-20T19:37:42Z","caller":"/woodpecker/server/plugins/secrets/encrypted.go:219","message":"encryption"}
{"level":"debug","id":9,"name":"some_secret","time":"2023-08-20T19:37:42Z","caller":"/woodpecker/server/plugins/secrets/encrypted.go:230","message":"decryption"}
```
and a row in the database similar to:
```psql
woodpecker=# select secret_id, secret_name, secret_value from secrets;
 secret_id | secret_name |                          secret_value
-----------+-------------+----------------------------------------------------------------
         9 | some_secret | PUattjAz6EOP28sbJOEaDSZyXRDrPxGQv9EyQHQPimrWLQELr59WYp83DUNQ6w
(1 row)
```

:::note
You won't get exactly the same secret's encrypted value, because a random nonce is used.
:::
