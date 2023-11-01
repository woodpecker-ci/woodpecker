# Secrets encryption

:::danger
Secrets encryption is currently broken and therefore disabled by default. It will be fixed in an upcoming release.

Check:

- <https://github.com/woodpecker-ci/woodpecker/issues/1541> and
- <https://github.com/woodpecker-ci/woodpecker/pull/2300>

:::

By default, Woodpecker does not encrypt secrets in its database. You can enable encryption
using simple AES key or more advanced [Google TINK](https://developers.google.com/tink) encryption.

## Common

### Enabling secrets encryption

To enable secrets encryption and encrypt all existing secrets in database set
`WOODPECKER_ENCRYPTION_KEY`, `WOODPECKER_ENCRYPTION_KEY_FILE` or `WOODPECKER_ENCRYPTION_TINK_KEYSET_PATH` environment
variable depending on encryption method of your choice.

After encryption is enabled you will be unable to start Woodpecker server without providing valid encryption key!

### Disabling encryption and decrypting all secrets

To disable secrets encryption and decrypt database you need to start server with valid
`WOODPECKER_ENCRYPTION_KEY` or `WOODPECKER_ENCRYPTION_TINK_KEYSET_FILE` environment variable set depending on
enabled encryption method, and `WOODPECKER_ENCRYPTION_DISABLE` set to true.

After secrets was decrypted server will proceed working in unencrypted mode. You will not need to use "disable encryption"
variable or encryption keys to start server anymore.

## AES

Simple AES encryption.

### Configuration

You can manage encryption on server using these environment variables:

- `WOODPECKER_ENCRYPTION_KEY` - encryption key
- `WOODPECKER_ENCRYPTION_KEY_FILE` - file to read encryption key from
- `WOODPECKER_ENCRYPTION_DISABLE` - disable encryption flag used to decrypt all data on server

## TINK

TINK uses AEAD encryption instead of simple AES and supports key rotation.

### Configuration

You can manage encryption on server using these two environment variables:

- `WOODPECKER_ENCRYPTION_TINK_KEYSET_FILE` - keyset filepath
- `WOODPECKER_ENCRYPTION_DISABLE` - disable encryption flag used to decrypt all data on server

### Encryption keys

You will need plaintext AEAD-compatible Google TINK keyset to encrypt your data.

To generate it and then rotate keys if needed, install `tinkey`([installation guide](https://developers.google.com/tink/install-tinkey))

Keyset contains one or more keys, used to encrypt or decrypt your data, and primary key ID, used to determine which key
to use while encrypting new data.

Keyset generation example:

```shell
tinkey create-keyset --key-template AES256_GCM --out-format json --out keyset.json
```

### Key rotation

Use `tinkey` to rotate encryption keys in your existing keyset:

```shell
tinkey rotate-keyset --in keyset_v1.json --out keyset_v2.json --key-template AES256_GCM
```

Then you just need to replace server keyset file with the new one. At the moment server detects new encryption
keyset it will re-encrypt all existing secrets with the new key, so you will be unable to start server with previous
keyset anymore.
