#!/usr/bin/env python3
"""
Symphonix-Health Google Tink AEAD Keyset Generator
Generates a cryptographically secure AEAD keyset JSON (AES256-GCM) compliant with
Google Tink specification and Woodpecker CI database secret encryption.
"""

import argparse
import base64
import json
import os
import secrets
import sys

def generate_tink_aes256_gcm_keyset() -> dict:
    # 32 cryptographically random bytes for AES-256
    raw_key = os.urandom(32)
    # Protobuf serialization of google.crypto.tink.AesGcmKey:
    # tag 1 (version=0): 0x08, 0x00
    # tag 3 (key_value): 0x1A, 0x20 (length 32), followed by raw_key
    proto_payload = b"\x08\x00\x1a\x20" + raw_key
    b64_key_data = base64.b64encode(proto_payload).decode("ascii")

    # Positive 31-bit integer for keyId (standard Tink compatible)
    key_id = secrets.randbits(31)

    return {
        "primaryKeyId": key_id,
        "key": [
            {
                "keyData": {
                    "typeUrl": "type.googleapis.com/google.crypto.tink.AesGcmKey",
                    "value": b64_key_data,
                    "keyMaterialType": "SYMMETRIC",
                },
                "status": "ENABLED",
                "keyId": key_id,
                "outputPrefixType": "TINK",
            }
        ],
    }

def main():
    parser = argparse.ArgumentParser(
        description="Generate Google Tink AEAD Keyset for Symphonix HealthCI"
    )
    parser.add_argument(
        "-o",
        "--output",
        default="tink-keyset.json",
        help="Target output file path (default: tink-keyset.json)",
    )
    args = parser.parse_args()

    keyset = generate_tink_aes256_gcm_keyset()
    with open(args.output, "w", encoding="utf-8") as f:
        json.dump(keyset, f, indent=2)

    print(f"[OK] Generated secure Tink AEAD keyset at: {args.output}")
    print(f"     Primary Key ID: {keyset['primaryKeyId']}")
    print(f"     Algorithm: AES256-GCM (type.googleapis.com/google.crypto.tink.AesGcmKey)")

if __name__ == "__main__":
    main()
