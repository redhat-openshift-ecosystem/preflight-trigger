## preflight-trigger

Create on-demand preflight jobs in openshift-ci system

### Options

```
      --asset-type string                   Type of asset to trigger
      --dry-run                             Do perform any actions, but do not actually trigger the job
      --gpg-decryption-private-key string   GPG private key to use for decryption
      --gpg-decryption-public-key string    GPG public key to use for decryption
      --gpg-encryption-private-key string   GPG private key to use for encryption
      --gpg-encryption-public-key string    GPG public key to use for encryption
  -h, --help                                help for preflight-trigger
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

* [preflight-trigger artifacts](preflight-trigger_artifacts.md)	 - Get artifacts from a given openshift-ci job
* [preflight-trigger checkhealth](preflight-trigger_checkhealth.md)	 - Verify cluster under test is ready.
* [preflight-trigger completion](preflight-trigger_completion.md)	 - Generate the autocompletion script for the specified shell
* [preflight-trigger create](preflight-trigger_create.md)	 - Create contains subcommands for creating jobs and documentation.
* [preflight-trigger decode](preflight-trigger_decode.md)	 - Decode a value or local file; value or file location is required
* [preflight-trigger decrypt](preflight-trigger_decrypt.md)	 - Decrypt a GPG encrypted file or arbitrary data from stdin
* [preflight-trigger encode](preflight-trigger_encode.md)	 - Encode a value or local file; value or file location is required
* [preflight-trigger encrypt](preflight-trigger_encrypt.md)	 - Encrypt a file or arbitrary data from stdin

