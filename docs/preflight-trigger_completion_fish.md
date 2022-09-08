## preflight-trigger completion fish

Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	preflight-trigger completion fish | source

To load completions for every new session, execute once:

	preflight-trigger completion fish > ~/.config/fish/completions/preflight-trigger.fish

You will need to start a new shell for this setup to take effect.


```
preflight-trigger completion fish [flags]
```

### Options

```
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions
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

* [preflight-trigger completion](preflight-trigger_completion.md)	 - Generate the autocompletion script for the specified shell

