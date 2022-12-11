# Secrets encryption

By default, Woodpecker does not encrypt secrets in its database. You can enable encryption 
using [Google TINK](https://developers.google.com/tink) encryption.

## Configuration

You can manage encryption on server using these two environment variables:
- `WOODPECKER_SECRETS_ENCRYPTION_KEYSET_FILE` - filepath of keyset which will be used to encrypt and decrypt your 
secrets in runtime 
- `WOODPECKER_SECRETS_DECRYPT_ALL_KEYSET_FILE` - filepath of the same keyset used to fully decrypt all the secrets on 
server startup and permanently disable encryption

## Encryption keys

You will need plaintext AES256_SIV Google TINK keyset to encrypt your data.

To generate it and then rotate keys, if needed, install `tinkey`([installation guide](https://developers.google.com/tink/install-tinkey))

Keyset contains one or more keys, used to encrypt or decrypt your data, and primary key ID, used to determine which key 
to use while encrypting new data.

New encryption keyset generation example:
```shell
tinkey create-keyset --key-template AES256_SIV --out-format json --out keyset.json`
```

Existing keyset key rotation example:
```shell
tinkey rotate-keyset —in keyset_v1.json —out keyset_v2.json —key-template AES256_SIV
```

## Server encryption lifecycle

### 1.Enabling secrets encryption

To enable secrets encryption and encrypt all existing secrets in database set environment vatiable `WOODPECKER_SECRETS_ENCRYPTION_KEYSET_FILE`.
After encryption is enabled you will be unable to start Woodpecker server without providing valid encryption keyset!

### 2.Encryption keys rotation

To rotate encryption keys you just need to replace keyset file with new one. At the moment server detects new encryption 
keyset it will re-encrypt all existing secrets with new key, so you will be unable to start server with previous 
keyset anymore.

### 3.Disabling encryption and decrypting all secrets

To disable secrets encryption and decrypt all secrets in database you need to start server with 
`WOODPECKER_SECRETS_DECRYPT_ALL_KEYSET_FILE` environment variable set to the latest encryption keyset filepath.

Note that you should not set `WOODPECKER_SECRETS_ENCRYPTION_KEYSET_FILE` in this case. Server will not start with both
encryption and decryption variables set at the same time.

After secrets was decrypted server will proceed working in unencrypted mode. You will not need to use "decrypt all" 
variable to start server anymore.
