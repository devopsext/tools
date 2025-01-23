package cmd

import (
	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/devopsext/utils"
	"github.com/spf13/cobra"
)

var cryptoRCAKeyOptions = vendors.CryptoRCAKeyOptions{
	Size: envGet("CRYPTO_RCA_KEY_SIZE", 2048).(int),
}

var cryptoRCAEncryptOptions = vendors.CryptoRCAEncryptOptions{
	Text:      envGet("CRYPTO_RCA_ENCRYPT_TEXT", "").(string),
	PublicKey: envGet("CRYPTO_RCA_ENCRYPT_PUBLIC_KEY", "").(string),
}

var cryptoRCADecryptOptions = vendors.CryptoRCADecryptOptions{
	Text:       envGet("CRYPTO_RCA_DECRYPT_TEXT", "").(string),
	PrivateKey: envGet("CRYPTO_RCA_DECRYPT_PRIVATE_KEY", "").(string),
}

var cryptoOutput = common.OutputOptions{
	Output: envGet("CRYPTO_OUTPUT", "").(string),
}

func cryptoNew(stdout *common.Stdout) *vendors.Crypto {

	common.Debug("Crypto", cryptoOutput, stdout)

	return vendors.NewCrypto()
}

func NewCryptoCommand() *cobra.Command {

	cryptoCmd := &cobra.Command{
		Use:   "crypto",
		Short: "Crypto tools",
	}
	flags := cryptoCmd.PersistentFlags()
	flags.StringVar(&cryptoOutput.Output, "crypto-output", cryptoOutput.Output, "Crypto output")

	rcaCmd := &cobra.Command{
		Use:   "rca",
		Short: "Crypto RCA tools",
	}
	cryptoCmd.AddCommand(rcaCmd)

	rcaGenerateKeyCmd := &cobra.Command{
		Use:   "generate-key",
		Short: "Crypto RCA generate keys",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Crypto RCA generate keys...")
			common.Debug("Crypto", cryptoRCAKeyOptions, stdout)

			bytes, err := cryptoNew(stdout).CustomRCAGenerateKey(cryptoRCAKeyOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(cryptoOutput, "Crypto", []interface{}{cryptoRCAKeyOptions}, bytes, stdout)
		},
	}
	flags = rcaGenerateKeyCmd.PersistentFlags()
	flags.IntVar(&cryptoRCAKeyOptions.Size, "crypto-rca-key-size", cryptoRCAKeyOptions.Size, "Crypto RCA key size")
	rcaCmd.AddCommand(rcaGenerateKeyCmd)

	rcaEncryptCmd := &cobra.Command{
		Use:   "encrypt",
		Short: "Crypto RCA encrypt",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Crypto RCA encrypting...")
			common.Debug("Crypto", cryptoRCAEncryptOptions, stdout)

			textBytes, err := utils.Content(cryptoRCAEncryptOptions.Text)
			if err != nil {
				stdout.Panic(err)
			}
			cryptoRCAEncryptOptions.Text = string(textBytes)

			publicKeyBytes, err := utils.Content(cryptoRCAEncryptOptions.PublicKey)
			if err != nil {
				stdout.Panic(err)
			}
			cryptoRCAEncryptOptions.PublicKey = string(publicKeyBytes)

			bytes, err := cryptoNew(stdout).CustomRCAEncrypt(cryptoRCAEncryptOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(cryptoOutput, "Crypto", []interface{}{cryptoRCAEncryptOptions}, bytes, stdout)
		},
	}
	flags = rcaEncryptCmd.PersistentFlags()
	flags.StringVar(&cryptoRCAEncryptOptions.Text, "crypto-rca-encrypt-text", cryptoRCAEncryptOptions.Text, "Crypto RCA encrypt text")
	flags.StringVar(&cryptoRCAEncryptOptions.PublicKey, "crypto-rca-encrypt-public-key", cryptoRCAEncryptOptions.PublicKey, "Crypto RCA encrypt public key")
	rcaCmd.AddCommand(rcaEncryptCmd)

	rcaDecryptCmd := &cobra.Command{
		Use:   "decrypt",
		Short: "Crypto RCA decrypt",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Crypto RCA decrypting...")
			common.Debug("Crypto", cryptoRCADecryptOptions, stdout)

			textBytes, err := utils.Content(cryptoRCADecryptOptions.Text)
			if err != nil {
				stdout.Panic(err)
			}
			cryptoRCADecryptOptions.Text = string(textBytes)

			privateKeyBytes, err := utils.Content(cryptoRCADecryptOptions.PrivateKey)
			if err != nil {
				stdout.Panic(err)
			}
			cryptoRCADecryptOptions.PrivateKey = string(privateKeyBytes)

			bytes, err := cryptoNew(stdout).CustomRCADecrypt(cryptoRCADecryptOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(cryptoOutput, "Crypto", []interface{}{cryptoRCADecryptOptions}, bytes, stdout)
		},
	}
	flags = rcaDecryptCmd.PersistentFlags()
	flags.StringVar(&cryptoRCADecryptOptions.Text, "crypto-rca-decrypt-text", cryptoRCADecryptOptions.Text, "Crypto RCA decrypt text")
	flags.StringVar(&cryptoRCADecryptOptions.PrivateKey, "crypto-rca-decrypt-private-key", cryptoRCADecryptOptions.PrivateKey, "Crypto RCA decrypt private key")
	rcaCmd.AddCommand(rcaDecryptCmd)

	return cryptoCmd
}
