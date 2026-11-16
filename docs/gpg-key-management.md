# GPG Key Management and Rotation

This document describes how GPG keys are used in `preflight-trigger`, how to generate a new key pair, where keys are stored, and how to rotate them.

## Overview

`preflight-trigger` uses OpenPGP to encrypt artifacts (such as `dockerconfigjson`) on the submitter side and decrypt them in-cluster.

* Encryption: `preflight-trigger encrypt` encrypts with the recipient public key and signs with the sender private key (`ProtonMail/gopenpgp` `KeyRing.Encrypt`).
* Decryption: `preflight-trigger decrypt` decrypts with the recipient private key and verifies the signature with the sender public key (`KeyRing.Decrypt`).

In the deployed pipeline:

* The private key for encryption is stored in Google Secret Manager and is pulled at encrypt time.
* The public key for decryption is stored in the OpenShift cluster that runs the Prow job and is used to decrypt and verify artifacts.

Manage these secrets via the OpenShift CI Secret Manager CLI (`hack/secret-manager.sh` in `openshift/release`, aliased `sm` in this doc). It is the primary interface for CI secrets and is backed by Google Secret Manager. See [Using the Secret Manager CLI](https://docs.ci.openshift.org/architecture/cli-secret-manager/) for full documentation.
Only CI infrastructure can read secret values; Rover group members can create/update/delete but cannot read back values.

> Code references: `cmd/encrypt.go:36-92`, `cmd/decrypt.go:33-92`, `cmd/root.go:34-38` for flag definitions. Flags accept file paths to armored keys, not inline data.

## Flags and Environment

| Flag | Env var | Used by | Description |
|------|---------|---------|-------------|
| `--gpg-encryption-public-key` | `GPG_ENCRYPTION_PUBLIC_KEY` | `encrypt` | Path to armored public key file |
| `--gpg-encryption-private-key` | `GPG_ENCRYPTION_PRIVATE_KEY` | `encrypt` | Path to armored private key file |
| `--gpg-decryption-public-key` | `GPG_DECRYPTION_PUBLIC_KEY` | `decrypt` | Path to armored public key file |
| `--gpg-decryption-private-key` | `GPG_DECRYPTION_PRIVATE_KEY` | `decrypt` | Path to armored private key file |

`GPG_PASSPHRASE` is read from the environment in `encryptPreRun` and `decryptPreRun`. If keys are passphrase-protected, set this variable on both sides. Current implementation reads it but does not explicitly unlock passphrase-protected keys via `crypto` API; use unprotected keys or extend `cmd/encrypt.go` and `cmd/decrypt.go` to call `Key.Unlock`.

## Key Storage

### Secret Manager / Google Secret Manager (encryption side)

The underlying storage is Google Secret Manager, but CI teams interact with it through the Secret Manager CLI. This enforces collection/group/field hierarchy and Rover group access control.

* Secret holds the armored private key (and, where the implementation requires it, the public key). Update the secret version; do not delete old versions until rotation is verified.
* Consumers reference secrets in CI config via a `credentials` stanza. Every secret referenced this way is mounted as a file before the job step runs. Example:
  ```yaml
  credentials:
    - collection: my-collection
      group: preflight-gpg
      mount_path: /tmp/preflight-gpg
  # Mounts /tmp/preflight-gpg/private.asc, /tmp/preflight-gpg/public.asc etc.
  ```
  To mount a single field: add `field: private.asc` to the entry.
* At step runtime, invoke `preflight-trigger encrypt --gpg-encryption-private-key /tmp/preflight-gpg/private.asc --gpg-encryption-public-key /tmp/preflight-gpg/public.asc`.
* Restrict IAM to the service account that performs encryption.

#### Prerequisites

* `gcloud` CLI and `podman` (or `docker` with `CONTAINER_ENGINE=docker`) installed.
* Clone `openshift/release` and use `hack/secret-manager.sh`:
  ```bash
  git clone https://github.com/openshift/release.git
  alias sm="./release/hack/secret-manager.sh"
  sm login   # Authenticates via Red Hat SSO (browser flow, caches locally)
  sm clean   # To clear cached credentials
  ```
* Rover group must have access to the collection. Collections are granted in `core-services/sync-rover-groups/_config.yaml` under `groups.<rover-group>.secret_collections`. After merging the PR, the postsubmit `branch-ci-openshift-release-main-gsm-secrets-reconciler` syncs access (under a minute).
* Verify access:
  ```bash
  sm list -c my-collection                          # fields in collection
  sm list --rover-group my-rover-group              # collections for your group
  sm describe -c my-collection preflight-gpg/private.asc  # metadata only, never value
  ```

Update these values for your deployment (collection/group/field naming):

```
COLLECTION=my-collection
GROUP=preflight-gpg
FIELD_PRIVATE=private.asc
FIELD_PUBLIC=public.asc
```

Examples via Secret Manager CLI (backed by GSM):

```bash
# Create (first time)
sm create -c ${COLLECTION} ${GROUP}/${FIELD_PRIVATE} --from-file ./private.asc
sm create -c ${COLLECTION} ${GROUP}/${FIELD_PUBLIC}  --from-file ./public.asc
# Prompts for: Jira project, Rotation information, Request information (enter none if not applicable)

# Update (rotation)
sm update -c ${COLLECTION} ${GROUP}/${FIELD_PRIVATE} --from-file ./private.asc
sm update -c ${COLLECTION} ${GROUP}/${FIELD_PUBLIC}  --from-file ./public.asc

# Literal variant
sm create -c ${COLLECTION} ${GROUP}/${FIELD_PUBLIC} --from-literal "$(cat ./public.asc)"
```

Notes:
* `describe` never shows the value; previous versions are retained in storage but `sm` always serves the latest.
* For non-CI callers that use `gcloud` directly, the equivalent is `gcloud secrets versions add ${SECRET_ID} --project=${PROJECT_ID} --data-file=./private.asc`.

### OpenShift Cluster (decryption side)

* The public key is stored as a Secret or ConfigMap in the cluster namespace that runs ProwJobs (check `config.ProwJobNamespace` and `PFLT_NAMESPACE`), or mounted via the same Secret Manager `credentials` stanza if the decrypt step runs in CI.
* The Prow job mounts the key or injects it via environment and passes `--gpg-decryption-public-key` / `--gpg-decryption-private-key` to the decrypt step. When using Secret Manager, mount the collection/group:
  ```yaml
  credentials:
    - collection: my-collection
      group: preflight-gpg
      mount_path: /tmp/preflight-gpg
  ```
  Then: `preflight-trigger decrypt --gpg-decryption-private-key /tmp/preflight-gpg/private.asc --gpg-decryption-public-key /tmp/preflight-gpg/public.asc`.

Update these values for your deployment:

```
NAMESPACE=preflight
SECRET_NAME=preflight-gpg-public-key
KEY_FILE=public.asc
```

Example update when stored as an OpenShift Secret directly:

```bash
oc create secret generic ${SECRET_NAME} \
  --from-file=public.asc=./public.asc \
  --from-file=private.asc=./private.asc \
  -n ${NAMESPACE} --dry-run=client -o yaml | oc apply -f -

oc rollout restart deployment/<decrypting-component> -n ${NAMESPACE}  # if keys are mounted as env/volumes
```

> Replace `SECRET_NAME`, `NAMESPACE`, and deployment names with the values used in your `openshift/release` `02_config` / periodic definitions.

## Generating a New Key Pair

`preflight-trigger` uses `github.com/ProtonMail/gopenpgp/v2`. Any OpenPGP-compliant generator that produces ASCII-armored keys works. `gpg` (GnuPG) is recommended.

### Option A: GnuPG

```bash
# 1. Generate key (RSA 4096, no expiry or set expiry per policy)
gpg --batch --generate-key <<'KEYCFG'
Key-Type: RSA
Key-Length: 4096
Subkey-Type: RSA
Subkey-Length: 4096
Name-Real: preflight-trigger
Name-Email: preflight@redhat.com
Expire-Date: 0
%no-protection
%commit
KEYCFG

# For passphrase-protected key, omit %no-protection and set Passphrase.

# 2. Export armored keys
gpg --armor --export preflight@redhat.com > public.asc
gpg --armor --export-secret-keys preflight@redhat.com > private.asc

# 3. Verify
gpg --list-keys preflight@redhat.com
gpg --list-secret-keys preflight@redhat.com
```

For a passphrase-protected key:

```bash
gpg --armor --export-secret-keys preflight@redhat.com > private.asc
# Save passphrase in GSM alongside the key or in a separate secret.
```

### Option B: ProtonMail/gopenpgp (programmatic)

```go
package main

import (
    "fmt"
    "github.com/ProtonMail/gopenpgp/v2/crypto"
)

func main() {
    key, _ := crypto.GenerateKey("preflight-trigger", "preflight@redhat.com", "rsa", 4096)
    pub, _ := key.GetArmoredPublicKey()
    priv, _ := key.Armor()
    fmt.Println(pub)
    fmt.Println(priv)
}
```

Save output to `public.asc` and `private.asc`.

## Rotating Keys

Rotation replaces both the GSM private key and the cluster public key atomically. Old artifacts encrypted with the previous key remain decryptable only if the old key is retained.

### Prerequisites

* New `public.asc` and `private.asc` generated and verified.
* Access to update Google Secret Manager and the target OpenShift cluster.
* Maintenance window or dual-key support if zero-downtime is required.

### Steps

1. Generate new key pair and verify armoring:

   ```bash
   gpg --armor --export preflight@redhat.com > public.asc
   gpg --armor --export-secret-keys preflight@redhat.com > private.asc
   head -n 5 public.asc   # -----BEGIN PGP PUBLIC KEY BLOCK-----
   head -n 5 private.asc  # -----BEGIN PGP PRIVATE KEY BLOCK-----
   ```

2. Test encrypt/decrypt locally before publishing:

   ```bash
   echo "test payload" | preflight-trigger encrypt \
     --gpg-encryption-public-key ./public.asc \
     --gpg-encryption-private-key ./private.asc > /tmp/test.asc

   cat /tmp/test.asc | preflight-trigger decrypt \
     --gpg-decryption-public-key ./public.asc \
     --gpg-decryption-private-key ./private.asc

   # Should output: test payload
   ```

   For dockerconfig:

   ```bash
   preflight-trigger encrypt --pflt-docker-config ./config.json \
     --gpg-encryption-public-key ./public.asc \
     --gpg-encryption-private-key ./private.asc --output-path /tmp/config.asc
   preflight-trigger decrypt --file /tmp/config.asc \
     --gpg-decryption-public-key ./public.asc \
     --gpg-decryption-private-key ./private.asc
   ```

3. Publish private key to Secret Manager (backed by Google Secret Manager):

   ```bash
   # Ensure you are logged in: sm login
   sm update -c ${COLLECTION} ${GROUP}/${FIELD_PRIVATE} --from-file ./private.asc
   sm update -c ${COLLECTION} ${GROUP}/${FIELD_PUBLIC}  --from-file ./public.asc
   # Provide Jira project / rotation info / request info when prompted if creating for the first time.
   # If you operate outside CI infra, raw GSM equivalent:
   # gcloud secrets versions add ${SECRET_ID} --project=${PROJECT_ID} --data-file=./private.asc
   ```

4. Publish public key to OpenShift:

   ```bash
   oc create secret generic ${SECRET_NAME} \
     --from-file=public.asc=./public.asc \
     --from-file=private.asc=./private.asc \
     -n ${NAMESPACE} --dry-run=client -o yaml | oc apply -f -
   ```

   If the decrypt job requires both keys (current code reads both `GPGDecryptionPublicKey` and `GPGDecryptionPrivateKey`), publish both files.
   If the decrypt step runs in CI, mount the same Secret Manager collection/group via the `credentials` stanza instead of a raw OpenShift Secret.

5. Verify in-cluster:

   ```bash
   sm list -c ${COLLECTION}                          # confirm fields updated
   sm describe -c ${COLLECTION} ${GROUP}/${FIELD_PUBLIC}  # metadata / creation date
   oc get secret ${SECRET_NAME} -n ${NAMESPACE} -o jsonpath='{.data.public\.asc}' | base64 -d | head
   oc debug job/<recent-preflight-job> -- cat /path/to/mounted/public.asc  # adjust path
   ```

6. Run a canary ProwJob with `--dry-run` or a real submission:

   ```bash
   preflight-trigger create job --ocp-version 4.18 --pflt-index-image <image> \
     --test-asset <asset> --dry-run | grep -i gpg
   ```

   Submit a test job and confirm decrypt succeeds and artifacts are produced.

7. Retire old key version after verification:

   Previous versions are retained automatically in storage, but the CLI always serves the latest. There is no `sm` rollback-to-version; if needed, re-apply the old file with `sm update`. For raw GSM:
   ```bash
   gcloud secrets versions disable <old-version> --secret=${SECRET_ID} --project=${PROJECT_ID}
   ```
   Keep the old key for the retention window needed to decrypt in-flight artifacts, then destroy.

### Dual-Key / Zero-Downtime Rotation

Current `encrypt`/`decrypt` implementations support a single key pair. For zero downtime:

* Deploy the decrypt side to accept both old and new keys (requires code change to try multiple `KeyRing`s).
* Alternatively, coordinate a short window: publish new keys to both GSM and cluster simultaneously, and re-encrypt any artifacts created during the window if decryption fails.

### Rollback

If decryption fails after rotation:

```bash
sm update -c ${COLLECTION} ${GROUP}/${FIELD_PRIVATE} --from-file ./private.asc.prev
sm update -c ${COLLECTION} ${GROUP}/${FIELD_PUBLIC}  --from-file ./public.asc.prev
# Raw GSM fallback:
# gcloud secrets versions access <previous-version> --secret=${SECRET_ID} --project=${PROJECT_ID} > ./private.asc.prev
oc create secret generic ${SECRET_NAME} --from-file=public.asc=./public.asc.prev -n ${NAMESPACE} --dry-run=client -o yaml | oc apply -f -
```

Re-publish the previous `public.asc` and `private.asc` to both stores.

## Verification Checklist

* [ ] `public.asc` and `private.asc` are ASCII-armored and match (`gpg --list-packets` shows same key ID).
* [ ] Local `encrypt` then `decrypt` round-trip succeeds for both stdin and file modes.
* [ ] Secret Manager fields updated (`sm list -c ${COLLECTION}` shows new creation date via `sm describe`).
* [ ] GSM backing secret new version is `ENABLED` and accessible to the encryption service account (if applicable).
* [ ] OpenShift Secret in `preflight` namespace contains updated keys.
* [ ] Canary ProwJob decrypts successfully.
* [ ] Old key version retained until no in-flight artifacts remain.

## Troubleshooting

* `GPG_ENCRYPTION_PUBLIC_KEY and GPG_ENCRYPTION_PRIVATE_KEY must be set`: flags or env vars not set; `encrypt` requires both encryption keys.
* `Unable to read decryption private key`: file path wrong or Secret not mounted.
* `Error creating new public key from armored`: file is not armored or contains extra whitespace; re-export with `gpg --armor`.
* `Error decrypting message`: key mismatch, passphrase not set via `GPG_PASSPHRASE`, or message was encrypted with a different key pair. Verify same pair is used on both sides.
* No GSM access: check `gcloud auth` and IAM `secretmanager.secretAccessor` on the secret, or `sm login` / Rover group membership (`sm list --rover-group my-rover-group`, `core-services/sync-rover-groups/_config.yaml`, postsubmit `branch-ci-openshift-release-main-gsm-secrets-reconciler`).
* `You don't have permission to access secrets in collection ...`: Rover group does not have access; add to `sync-rover-groups/_config.yaml` and wait for reconciler.
* `Secret is in inconsistent state`: run `sm delete -c $COLLECTION $GROUP/$FIELD` then `sm create` again.
* `sm get-sa fails`: collection has no `updater_service_accounts` entry; request one if automating rotation.

## References

* Implementation: `cmd/encrypt.go`, `cmd/decrypt.go`, `cmd/root.go`
* Library: `github.com/ProtonMail/gopenpgp/v2`
* GnuPG: https://gnupg.org/documentation/
* OpenShift CI Secret Manager CLI: https://docs.ci.openshift.org/architecture/cli-secret-manager/
* Secret collections config: `openshift/release:core-services/sync-rover-groups/_config.yaml`
