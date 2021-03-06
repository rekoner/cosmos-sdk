package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgCreateValidator{}, "cosmos-sdk/MsgCreateValidator", nil)
	cdc.RegisterConcrete(MsgEditValidator{}, "cosmos-sdk/MsgEditValidator", nil)
	cdc.RegisterConcrete(MsgDelegate{}, "cosmos-sdk/MsgDelegate", nil)
	cdc.RegisterConcrete(MsgBeginUnbonding{}, "cosmos-sdk/BeginUnbonding", nil)
	cdc.RegisterConcrete(MsgCompleteUnbonding{}, "cosmos-sdk/CompleteUnbonding", nil)
	cdc.RegisterConcrete(MsgBeginRedelegate{}, "cosmos-sdk/BeginRedelegate", nil)
	cdc.RegisterConcrete(MsgCompleteRedelegate{}, "cosmos-sdk/CompleteRedelegate", nil)
}

// generic sealed codec to be used throughout sdk
var MsgCdc *codec.Codec

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	MsgCdc = cdc.Seal()
}
