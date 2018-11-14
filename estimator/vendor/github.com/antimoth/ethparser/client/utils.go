package client

import (
	"context"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func EnsureContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.TODO()
	}
	return ctx
}

func HexToAddress(sHexAddr string) common.Address {
	return common.HexToAddress(sHexAddr)
}

func GetAddressFromPub(sPub string) (common.Address, []byte, error) {
	bPub, err := hex.DecodeString(sPub)
	if err != nil {
		return common.Address{}, bPub, err
	}

	pubKey, err := crypto.UnmarshalPubkey(bPub)
	if err != nil {
		return common.Address{}, bPub, err
	}

	return crypto.PubkeyToAddress(*pubKey), bPub, nil
}
