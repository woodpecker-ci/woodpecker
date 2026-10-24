# Symphonix HealthCI: Phase 1 Hardening & PostgreSQL Foundation

This module contains the Phase 1 hardening assets and operational configurations for adapting Woodpecker CI to Symphonix-Health's air-gapped clinical AI environment.

---

## 1. Phase 1 Code Modifications Completed

1. **Google Tink AEAD Secret Encryption Service Activated**:
   - **File**: [`server/model/secret.go`](file:///C:/Users/Joshua%20Tedam/github/woodpecker/server/model/secret.go#L33-L44)
     - Standardized `model.SecretStore` to use `*model.ListOptionsWithAll` across all listing methods.
   - **File**: [`server/services/encryption/wrapper/store/secret_store.go`](file:///C:/Users/Joshua%20Tedam/github/woodpecker/server/services/encryption/wrapper/store/secret_store.go)
     - Updated `EncryptedSecretStore` method signatures to match `model.SecretStore` with `*model.ListOptionsWithAll`.
   - **File**: [`server/services/secret/db.go`](file:///C:/Users/Joshua%20Tedam/github/woodpecker/server/services/secret/db.go#L24-L32)
     - Updated `db` service struct to accept `model.SecretStore`, allowing transparent wrapping with `EncryptedSecretStore`.
   - **File**: [`server/services/setup.go`](file:///C:/Users/Joshua%20Tedam/github/woodpecker/server/services/setup.go#L59-L76) & [`server/services/manager.go`](file:///C:/Users/Joshua%20Tedam/github/woodpecker/server/services/manager.go#L85-L98)
     - Resolved `TODO(1544)`: wired `setupSecretService` to instantiate `EncryptedSecretStore` and configure `encryption.Encryption(c, store)` when `WOODPECKER_ENCRYPTION_TINK_KEYSET_FILE` or raw key flags are set.

2. **Kubernetes Runner Security Hardening (Defense-in-Depth)**:
   - **File**: [`pipeline/backend/kubernetes/pod.go`](file:///C:/Users/Joshua%20Tedam/github/woodpecker/pipeline/backend/kubernetes/pod.go#L710-L745)
     - Added strict security defaults for unprivileged steps: automatically sets `AllowPrivilegeEscalation: false` and drops all Linux capabilities (`Drop: ["ALL"]`) by default, preventing container breakout and ensuring NHS DCB0160 compliance.

---

## 2. Quickstart: Running HealthCI with PostgreSQL & Tink AEAD

### Step 1: Generate a Tink AEAD Keyset
Run the included cryptographic keyset generator:

```bash
python generate_tink_keyset.py -o tink-keyset.json
```

This creates an AES256-GCM keyset compatible with Google Tink.

### Step 2: Start the Hardened Stack
Deploy PostgreSQL 16, Woodpecker Server (with Tink AEAD active), and the Hardened Agent:

```bash
docker compose -f docker-compose.symphonix-postgres.yml up -d
```

### Step 3: Verify Initialization
1. Check that PostgreSQL is healthy:
   ```bash
   docker exec -it symphonix-healthci-postgres pg_isready -U woodpecker -d woodpecker
   ```
2. Verify server log confirms Tink AEAD encryption is active:
   ```bash
   docker logs symphonix-healthci-server | grep "database secret store encryption configured successfully"
   ```
3. Access the HealthCI Web UI at `http://localhost:8000`.
4. Inspect Prometheus step-level metrics at `http://localhost:9001/metrics`.

---

## 3. Production Deployment Notes (NHS Trust Air-Gap)

- **Database Connection Strings**: In hospital production environments, configure SSL encryption (`sslmode=verify-full&sslrootcert=/etc/ssl/certs/hospital-ca.crt`).
- **Keyset Storage**: Store `tink-keyset.json` in a secure Kubernetes Secret or mount it from OpenBao/HashiCorp Vault using CSI Secret Store driver.
- **Zero Internet Access**: Pre-load worker images into an internal air-gapped container registry (Harbor/Forgejo).
