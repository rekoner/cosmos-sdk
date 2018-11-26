package init

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/cosmos/cosmos-sdk/x/stake/client/cli"
	stakeTypes "github.com/cosmos/cosmos-sdk/x/stake/types"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/common"
)

const (
	defaultAmount                  = "100" + stakeTypes.DefaultBondDenom
	defaultCommissionRate          = "0.1"
	defaultCommissionMaxRate       = "0.2"
	defaultCommissionMaxChangeRate = "0.01"
)

// GenTxCmd builds the gaiad gentx command.
// nolint: errcheck
func GenTxCmd(ctx *server.Context, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gentx",
		Short: "Generate a genesis tx carrying a self delegation",
		Long: fmt.Sprintf(`This command is an alias of the 'gaiad tx create-validator' command'.

It creates a genesis piece carrying a self delegation with the
following delegation and commission default parameters:

	delegation amount:           %s
	commission rate:             %s
	commission max rate:         %s
	commission max change rate:  %s
`, defaultAmount, defaultCommissionRate, defaultCommissionMaxRate, defaultCommissionMaxChangeRate),
		RunE: func(cmd *cobra.Command, args []string) error {

			config := ctx.Config
			config.SetRoot(viper.GetString(tmcli.HomeFlag))
			nodeID, valPubKey, err := InitializeNodeValidatorFiles(ctx.Config)
			if err != nil {
				return err
			}
			ip, err := server.ExternalIP()
			if err != nil {
				return err
			}
			genDoc, err := loadGenesisDoc(cdc, config.GenesisFile())
			if err != nil {
				return err
			}

			kb, err := keys.GetKeyBaseFromDir(viper.GetString(flagClientHome))
			if err != nil {
				return err
			}
			if _, err = kb.Get(viper.GetString(client.FlagName)); err != nil {
				return err
			}

			// Read --pubkey, if empty take it from priv_validator.json
			if valPubKeyString := viper.GetString(cli.FlagPubKey); valPubKeyString != "" {
				valPubKey, err = sdk.GetConsPubKeyBech32(valPubKeyString)
				if err != nil {
					return err
				}
			}
			// Run gaiad tx create-validator
			prepareFlagsForTxCreateValidator(config, nodeID, ip, genDoc.ChainID, valPubKey)
			cliCtx, txBldr, msg, err := cli.BuildCreateValidatorMsg(
				context.NewCLIContext().WithCodec(cdc),
				authtxb.NewTxBuilderFromCLI().WithCodec(cdc),
			)
			if err != nil {
				return err
			}

			w, err := ioutil.TempFile("", "gentx")
			if err != nil {
				return err
			}
			unsignedGenTxFilename := w.Name()
			defer os.Remove(unsignedGenTxFilename)

			if err := utils.PrintUnsignedStdTx(w, txBldr, cliCtx, []sdk.Msg{msg}, true); err != nil {
				return err
			}

			prepareFlagsForTxSign()
			signCmd := authcmd.GetSignCommand(cdc)

			outputDocument, err := makeOutputFilepath(config.RootDir, nodeID)
			if err != nil {
				return err
			}
			viper.Set("output-document", outputDocument)
			return signCmd.RunE(nil, []string{unsignedGenTxFilename})
		},
	}

	cmd.Flags().String(tmcli.HomeFlag, app.DefaultNodeHome, "node's home directory")
	cmd.Flags().String(flagClientHome, app.DefaultCLIHome, "client's home directory")
	cmd.Flags().String(client.FlagName, "", "name of private key with which to sign the gentx")
	cmd.Flags().AddFlagSet(cli.FsCommissionCreate)
	cmd.Flags().AddFlagSet(cli.FsAmount)
	cmd.Flags().AddFlagSet(cli.FsPk)
	cmd.MarkFlagRequired(client.FlagName)
	return cmd
}

func prepareFlagsForTxCreateValidator(config *cfg.Config, nodeID, ip, chainID string,
	valPubKey crypto.PubKey) {
	viper.Set(tmcli.HomeFlag, viper.GetString(flagClientHome)) // --home
	viper.Set(client.FlagChainID, chainID)
	viper.Set(client.FlagFrom, viper.GetString(client.FlagName))   // --from
	viper.Set(cli.FlagNodeID, nodeID)                              // --node-id
	viper.Set(cli.FlagIP, ip)                                      // --ip
	viper.Set(cli.FlagPubKey, sdk.MustBech32ifyConsPub(valPubKey)) // --pubkey
	viper.Set(cli.FlagGenesisFormat, true)                         // --genesis-format
	viper.Set(cli.FlagMoniker, config.Moniker)                     // --moniker
	if config.Moniker == "" {
		viper.Set(cli.FlagMoniker, viper.GetString(client.FlagName))
	}
	if viper.GetString(cli.FlagAmount) == "" {
		viper.Set(cli.FlagAmount, defaultAmount)
	}
	if viper.GetString(cli.FlagCommissionRate) == "" {
		viper.Set(cli.FlagCommissionRate, defaultCommissionRate)
	}
	if viper.GetString(cli.FlagCommissionMaxRate) == "" {
		viper.Set(cli.FlagCommissionMaxRate, defaultCommissionMaxRate)
	}
	if viper.GetString(cli.FlagCommissionMaxChangeRate) == "" {
		viper.Set(cli.FlagCommissionMaxChangeRate, defaultCommissionMaxChangeRate)
	}
}

func prepareFlagsForTxSign() {
	viper.Set("offline", true)
}

func makeOutputFilepath(rootDir, nodeID string) (string, error) {
	writePath := filepath.Join(rootDir, "config", "gentx")
	if err := common.EnsureDir(writePath, 0700); err != nil {
		return "", err
	}
	return filepath.Join(writePath, fmt.Sprintf("gentx-%v.json", nodeID)), nil
}