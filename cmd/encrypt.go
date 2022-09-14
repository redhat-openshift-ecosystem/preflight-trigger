/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/gobuffalo/envy"
	"github.com/spf13/cobra"
	"log"
	"os"
)

// encryptCmd represents the encrypt command
var encryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt a file or arbitrary data from stdin",
	Long: `Primarily used for encrypting files and dockerconfigjson to be decrypted by
by preflight-trigger in openshift-ci and hosted certification pipeline.`,
	PreRun: encryptPreRun,
	Run:    encryptRun,
}

func init() {
	rootCmd.AddCommand(encryptCmd)
	encryptCmd.Flags().StringVarP(&CommandFlags.FileToEncrypt, "file", "", "", "File to encrypt")
}

func encryptPreRun(cmd *cobra.Command, args []string) {
	CommandFlags.GPGPassphrase = envy.Get("GPG_PASSPHRASE", "")
	if CommandFlags.GPGEncryptionPublicKey == "" || CommandFlags.GPGEncryptionPrivateKey == "" {
		log.Fatalf("GPG_ENCRYPTION_PUBLIC_KEY and GPG_ENCRYPTION_PRIVATE_KEY must be set")
	}
}

func encryptRun(cmd *cobra.Command, args []string) {
	encryptionpublickey, err := os.ReadFile(CommandFlags.GPGEncryptionPublicKey)
	if err != nil {
		return
	}

	encryptionprivatekey, err := os.ReadFile(CommandFlags.GPGEncryptionPrivateKey)
	if err != nil {
		return
	}

	var fcontent []byte

	if CommandFlags.PfltDockerConfig == "" && CommandFlags.FileToEncrypt == "" {
		fcontent, err = os.ReadFile(os.Stdin.Name())
		if err != nil {
			log.Printf("Error reading from stdin: %s", err)
		}
	} else {
		if CommandFlags.PfltDockerConfig != "" {
			fcontent, err = os.ReadFile(CommandFlags.PfltDockerConfig)
			if err != nil {
				log.Printf("Error reading dockerconfigjson: %s", err)
			}
		} else {
			fcontent, err = os.ReadFile(CommandFlags.FileToEncrypt)
			if err != nil {
				log.Printf("Error reading file: %s", err)
			}
		}
	}

	msg := crypto.NewPlainMessage(fcontent)
	var publickeyobj *crypto.Key
	var publickeyring *crypto.KeyRing
	publickeyobj, err = crypto.NewKeyFromArmored(string(encryptionpublickey))
	publickeyring, err = crypto.NewKeyRing(publickeyobj)

	var privatekeyobj *crypto.Key
	var privatekeyring *crypto.KeyRing
	privatekeyobj, err = crypto.NewKeyFromArmored(string(encryptionprivatekey))
	privatekeyring, err = crypto.NewKeyRing(privatekeyobj)

	var encryptedmsg *crypto.PGPMessage
	encryptedmsg, err = publickeyring.Encrypt(msg, privatekeyring)
	if err != nil {
		log.Fatalf("Error encrypting message: %v", err)
	}

	armor, err := encryptedmsg.GetArmored()

	if CommandFlags.OutputPath == "" {
		_, err = os.Stdout.Write([]byte(armor))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err = os.WriteFile(CommandFlags.OutputPath, []byte(armor), 0644)
		if err != nil {
			log.Fatalf("Unable to write to %s: %s", CommandFlags.OutputPath, err)
		}
	}
}
