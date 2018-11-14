// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package ethclient provides a client for the Ethereum RPC API.
package client

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/antimoth/ethparser/log"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("loglevel", "info")
}

var (
	eLogger = log.NewLogger(viper.GetString("loglevel"))
)

// Dial connects a client to the given URL.
func NewClient(rawurl string) (*Client, error) {
	ctx := context.Background()
	c, err := rpc.DialContext(ctx, rawurl)
	if err != nil {
		return nil, err
	}
	return &Client{c: c}, nil
}

func (ec *Client) Close() {
	ec.c.Close()
}

func (ec *Client) GetClient() *rpc.Client {
	return ec.c
}

func (ec *Client) GetBlockInfoFromHeight(height *big.Int) (*RpcBlock, error) {
	return ec.BlockByNumber(EnsureContext(nil), height)
}

// Blockchain Access

// BlockByNumber returns a block from the current canonical chain. If number is nil, the
// latest known block is returned.
//
// Note that loading full blocks requires two requests. Use HeaderByNumber
// if you don't need all transactions or uncle headers.
func (ec *Client) BlockByNumber(ctx context.Context, number *big.Int) (*RpcBlock, error) {
	return ec.getBlock(ctx, "eth_getBlockByNumber", toBlockNumArg(number), true)
}

func (ec *Client) getBlock(ctx context.Context, method string, args ...interface{}) (*RpcBlock, error) {
	var raw json.RawMessage

	err := ec.c.CallContext(ctx, &raw, method, args...)
	if err != nil {
		return nil, err

	} else if len(raw) == 0 {
		return nil, ethereum.NotFound
	}

	// fmt.Printf("raw prc rsp is %v \n", string(raw))

	// Decode header and transactions.
	var block *RpcBlock
	if err := json.Unmarshal(raw, &block); err != nil {
		return nil, err
	}

	// fmt.Printf("block hash is %v \n", block.BlockHash.Hex())

	return block, nil
}

// HeaderByHash returns the block header with the given hash.
func (ec *Client) HeaderByHash(ctx context.Context, hash common.Hash) (*RpcHeader, error) {
	var raw json.RawMessage

	err := ec.c.CallContext(ctx, &raw, "eth_getBlockByHash", hash, false)
	if err != nil {
		return nil, err

	} else if len(raw) == 0 {
		return nil, ethereum.NotFound
	}

	// fmt.Printf("raw prc rsp is %v \n", string(raw))

	// Decode header
	var head *RpcHeader
	if err := json.Unmarshal(raw, &head); err != nil {
		return nil, err
	}

	// fmt.Printf("block hash is %v \n", head.BlockHash.Hex())

	return head, nil
}

// HeaderByNumber returns a block header from the current canonical chain. If number is
// nil, the latest known header is returned.
func (ec *Client) HeaderByNumber(ctx context.Context, number *big.Int) (*RpcHeader, error) {
	var raw json.RawMessage

	err := ec.c.CallContext(ctx, &raw, "eth_getBlockByNumber", toBlockNumArg(number), false)
	if err != nil {
		return nil, err

	} else if len(raw) == 0 {
		return nil, ethereum.NotFound
	}

	// fmt.Printf("raw prc rsp is %v \n", string(raw))

	// Decode header
	var head *RpcHeader
	if err := json.Unmarshal(raw, &head); err != nil {
		return nil, err
	}

	// fmt.Printf("block hash is %v \n", head.BlockHash.Hex())

	return head, nil
}

func (ec *Client) IsValidTx(ctx context.Context, hash common.Hash) int {
	var raw json.RawMessage
	var tx RpcTx

	err := ec.c.CallContext(ctx, &raw, "eth_getTransactionByHash", hash)
	if err != nil {
		return VALID_TX_RPC_ERROR

	} else if len(raw) == 0 {
		return VALID_TX_NOT_FOUND

	}

	if err = json.Unmarshal(raw, &tx); err != nil {
		return VALID_TX_JSON_ERROR
	}
	return VALID_TX_VALID_TX

}

// TransactionByHash returns the transaction with the given hash.
func (ec *Client) TransactionByHash(ctx context.Context, hash common.Hash) (tx *RpcTx, err error) {
	var raw json.RawMessage

	err = ec.c.CallContext(ctx, &raw, "eth_getTransactionByHash", hash)
	if err != nil {
		return nil, err

	} else if len(raw) == 0 {
		return nil, ethereum.NotFound

	}

	if err := json.Unmarshal(raw, &tx); err != nil {
		return nil, err
	}

	// fmt.Printf("raw tx is %v \n", string(raw))

	return tx, nil
}

// TransactionReceipt returns the receipt of a transaction by transaction hash.
// Note that the receipt is not available for pending transactions.
func (ec *Client) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	var r *types.Receipt
	err := ec.c.CallContext(ctx, &r, "eth_getTransactionReceipt", txHash)
	if err == nil {
		if r == nil {
			return nil, ethereum.NotFound
		}
	}
	return r, err
}

// State Access
// BalanceAt returns the wei balance of the given account.
// The block number can be nil, in which case the balance is taken from the latest known block.
func (ec *Client) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	var result hexutil.Big
	err := ec.c.CallContext(ctx, &result, "eth_getBalance", account, toBlockNumArg(blockNumber))
	return (*big.Int)(&result), err
}

// CodeAt returns the contract code of the given account.
// The block number can be nil, in which case the code is taken from the latest known block.
func (ec *Client) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	var result hexutil.Bytes
	err := ec.c.CallContext(ctx, &result, "eth_getCode", account, toBlockNumArg(blockNumber))
	return result, err
}

// NonceAt returns the account nonce of the given account.
// The block number can be nil, in which case the nonce is taken from the latest known block.
func (ec *Client) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	var result hexutil.Uint64
	err := ec.c.CallContext(ctx, &result, "eth_getTransactionCount", account, toBlockNumArg(blockNumber))
	return uint64(result), err
}

// Pending State

// PendingCodeAt returns the contract code of the given account in the pending state.
func (ec *Client) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	var result hexutil.Bytes
	err := ec.c.CallContext(ctx, &result, "eth_getCode", account, "pending")
	return result, err
}

// PendingNonceAt returns the account nonce of the given account in the pending state.
// This is the nonce that should be used for the next transaction.
func (ec *Client) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {

	var result hexutil.Uint64
	err := ec.c.CallContext(ctx, &result, "eth_getTransactionCount", account, "pending")

	uintR := uint64(result)
	return uintR, err
}

// Contract Calling
// CallContract executes a message call transaction, which is directly executed in the VM
// of the node, but never mined into the blockchain.
// blockNumber selects the block height at which the call runs. It can be nil, in which
// case the code is taken from the latest known block. Note that state from very old
// blocks might not be available.
func (ec *Client) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	var hex hexutil.Bytes
	err := ec.c.CallContext(ctx, &hex, "eth_call", toCallArg(msg), toBlockNumArg(blockNumber))
	if err != nil {
		return nil, err
	}
	return hex, nil
}

func (ec *Client) BlockNumber() (*big.Int, error) {
	var hex hexutil.Big
	if err := ec.c.CallContext(EnsureContext(nil), &hex, "eth_blockNumber"); err != nil {
		eLogger.Error("Get blockNumber error!", "error", err)
		return nil, err
	}
	return (*big.Int)(&hex), nil
}

func (ec *Client) EthSubscribe(channel interface{}, args ...interface{}) (ethereum.Subscription, error) {
	return ec.c.EthSubscribe(EnsureContext(nil), channel, args...)
}

// SuggestGasPrice retrieves the currently suggested gas price to allow a timely
// execution of a transaction.
func (ec *Client) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	var hex hexutil.Big
	if err := ec.c.CallContext(ctx, &hex, "eth_gasPrice"); err != nil {
		return nil, err
	}
	return (*big.Int)(&hex), nil
}

// EstimateGas tries to estimate the gas needed to execute a specific transaction based on
// the current pending state of the backend blockchain. There is no guarantee that this is
// the true gas limit requirement as other transactions may be added or removed by miners,
// but it should provide a basis for setting a reasonable default.
func (ec *Client) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	var hex hexutil.Uint64
	err := ec.c.CallContext(ctx, &hex, "eth_estimateGas", toCallArg(msg))
	if err != nil {
		return 0, err
	}
	return uint64(hex), nil
}

// SendTransaction injects a signed transaction into the pending pool for execution.
// If the transaction was a contract creation use the TransactionReceipt method to get the
// contract address after the transaction has been mined.
func (ec *Client) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return err
	}
	return ec.c.CallContext(ctx, nil, "eth_sendRawTransaction", common.ToHex(data))
}

func toCallArg(msg ethereum.CallMsg) interface{} {
	arg := map[string]interface{}{
		"from": msg.From,
		"to":   msg.To,
	}
	if len(msg.Data) > 0 {
		arg["data"] = hexutil.Bytes(msg.Data)
	}
	if msg.Value != nil {
		arg["value"] = (*hexutil.Big)(msg.Value)
	}
	if msg.Gas != 0 {
		arg["gas"] = hexutil.Uint64(msg.Gas)
	}
	if msg.GasPrice != nil {
		arg["gasPrice"] = (*hexutil.Big)(msg.GasPrice)
	}
	return arg
}

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	return hexutil.EncodeBig(number)
}
