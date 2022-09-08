## preflight-trigger checkhealth

Verify cluster under test is ready.

### Synopsis

Check that the /readyz endpoint is healthy. Also check that
the cluster is not in a degraded state. We do this by checking the
Degraded, Available, and Progressing status' of clusteroperator
resources. We also check the clusterversion resource to ensure its
Progrssing status is not true.'

```
preflight-trigger checkhealth [flags]
```

### Options

```
  -h, --help   help for checkhealth
```

### Options inherited from parent commands

```
      --asset-type string                   Type of asset to trigger
      --dry-run                             Do perform any actions, but do not actually trigger the job
      --gpg-decryption-private-key string   GPG private key to use for decryption
      --gpg-decryption-public-key string    GPG public key to use for decryption
      --gpg-encryption-private-key string   GPG private key to use for encryption
      --gpg-encryption-public-key string    GPG public key to use for encryption
      --hidden                              Hide job in the list of jobs visible by deck
      --job-name string                     Name of the job to trigger
      --job-suffix string                   Suffix to append to the job name
      --ocp-version string                  Version of OCP to use
      --output-path string                  Path to output the job to
      --pflt-artifacts string               Path to artifacts to use for preflight (default "artifacts")
      --pflt-docker-config string           Docker config to use for preflight
      --pflt-index-image string             Index image to use for preflight
      --pflt-log-file string                Path to log file to use for preflight
      --pflt-log-level string               Level of logging to use for preflight (default "trace")
      --pflt-namespace string               Namespace to use for preflight
      --pflt-service-account string         Service account to use for preflight
      --release-image-ref string            Release image reference to use for preflight
      --test-asset string                   Test asset to use for preflight
```

### SEE ALSO

* [preflight-trigger](preflight-trigger.md)	 - Create on-demand preflight jobs in openshift-ci system

