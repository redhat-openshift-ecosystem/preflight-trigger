/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"os"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/gobuffalo/envy"
	"github.com/spf13/cobra"
)

// decryptCmd represents the decrypt command
var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Decrypt a GPG encrypted file or arbitrary data from stdin",
	Long: `Primarily used for decrypting files and dockerconfigjson
encrypted with the preflight-trigger tool.`,
	PreRun: decryptPreRun,
	Run:    decryptRun,
}

func init() {
	rootCmd.AddCommand(decryptCmd)
	decryptCmd.Flags().StringVarP(&CommandFlags.FileToDecrypt, "file", "", "", "File to decrypt")
}

func decryptPreRun(cmd *cobra.Command, args []string) {
	CommandFlags.GPGPassphrase = envy.Get("GPG_PASSPHRASE", "")
}

func decryptRun(cmd *cobra.Command, args []string) {
	decryptionprivatekey, err := os.ReadFile(CommandFlags.GPGDecryptionPrivateKey)
	if err != nil {
		log.Fatal(err)
	}

	decryptionpublickey, err := os.ReadFile(CommandFlags.GPGDecryptionPublicKey)
	if err != nil {
		log.Fatal(err)
	}

	var fcontent []byte

	if CommandFlags.PfltDockerConfig == "" && CommandFlags.FileToDecrypt == "" {
		fcontent, err = os.ReadFile(os.Stdin.Name())
		if err != nil {
			log.Printf("Error reading from stdin: %s", err)
		}
	} else {
		if CommandFlags.PfltDockerConfig != "" {
			fcontent = []byte(CommandFlags.PfltDockerConfig)
		} else {
			fcontent, err = os.ReadFile(CommandFlags.FileToDecrypt)
			if err != nil {
				log.Printf("Error reading file: %s", err)
			}
		}
	}

	var publickeyobj *crypto.Key
	var publickeyring *crypto.KeyRing
	publickeyobj, err = crypto.NewKeyFromArmored(string(decryptionpublickey))
	if err != nil {
		log.Fatalf("Error creating new public key from armored: %v", err)
	}
	publickeyring, err = crypto.NewKeyRing(publickeyobj)
	if err != nil {
		log.Fatalf("Error creating new public key ring: %v", err)
	}

	var privatekeyobj *crypto.Key
	var privatekeyring *crypto.KeyRing
	privatekeyobj, err = crypto.NewKeyFromArmored(string(decryptionprivatekey))
	if err != nil {
		log.Fatalf("Error creating new private key from armored: %v", err)
	}
	privatekeyring, err = crypto.NewKeyRing(privatekeyobj)
	if err != nil {
		log.Fatalf("Error creating new private key ring: %v", err)
	}

	var encryptedmsg *crypto.PGPMessage
	encryptedmsg, err = crypto.NewPGPMessageFromArmored(string(fcontent))
	if err != nil {
		log.Fatalf("Error decrypting message: %v", err)
	}

	unarmor, err := privatekeyring.Decrypt(encryptedmsg, publickeyring, crypto.GetUnixTime())
	if err != nil {
		log.Fatalf("Error decrypting message: %v", err)
	}

	if CommandFlags.OutputPath == "" {
		_, err = os.Stdout.Write(unarmor.Data)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err = os.WriteFile(CommandFlags.OutputPath, unarmor.Data, 0o644)
		if err != nil {
			log.Fatalf("Unable to write to %s: %s", CommandFlags.OutputPath, err)
		}
	}
}
