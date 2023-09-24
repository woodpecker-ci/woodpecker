# Secrets encryption

By default, Woodpecker does not encrypt secrets in its database. You can enable encryption using a simple AES key.

:::caution
Secrets encryption is experimental.
Check the [current state](https://github.com/woodpecker-ci/woodpecker/issues/1541)
:::

## Common

### Enabling secrets encryption

To enable secrets encryption set `WOODPECKER_SECRETS_ENCRYPTION_MODE` environment variable to the one of:
- `Disabled` (default) - use plain text secrets;
- `Enabled` - use encryption without migration (encryption) of already existing secrets;
- `EnabledAndEncrypt` - use encryption and  encrypt already existing secrets;
- `DisabledAndDecrypt` - use plain text secrets and run decryption of existing secrets.

:::caution
After migration, don't forget to switch the mode: `EnabledAndEncrypt` -> `Enabled`, `DisabledAndDecrypt` -> `Disabled`.
After encryption is enabled you will be unable to start Woodpecker server without providing valid encryption key!
:::

## AES
Simple AES encryption.

### Configuration
You can manage encryption on server using these environment variables:
- `WOODPECKER_ENCRYPTION_AES_KEY` - encryption key
- `WOODPECKER_ENCRYPTION_AES_KEY_FILE` - file to read encryption key from

One option to generate encryption key is to use OpenSSL, but any password generator can also be used. Recommended key length is at least 32 bytes:
```shell
$ openssl rand -base64 32
GjVHT007c4x3N+YPbsZld+hifba1enXkOzIb/0h6oW8=
```

If we run the server with `WOODPECKER_ENCRYPTION_AES_KEY='GjVHT007c4x3N+YPbsZld+hifba1enXkOzIb/0h6oW8='`, and try to create a secret like `some_secret:super-secret-value` 
then we'll get messages in the log similar to:
```log
{"level":"debug","id":1,"name":"s-name","time":"2023-09-24T10:49:21Z","caller":"/woodpecker/server/plugins/secrets/encrypted.go:48","message":"encryption"}
```
and a row in the database similar to:
```psql
woodpecker=# select secret_id, secret_name, secret_value from secrets;
 secret_id | secret_name |                          secret_value
-----------+-------------+----------------------------------------------------------------
         1 | some_secret | _aes_PUattjAz6EOP28sbJOEaDSZyXRDrPxGQv9EyQHQPimrWLQELr59WYp83DUNQ6w
(1 row)
```

:::note
You won't get exactly the same secret's encrypted value, because a random nonce is used.
:::
