package client

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

type Client struct {
	c *rpc.Client
}

type RpcTx struct {
	BlockNumber *hexutil.Big `json:"blockNumber" gencodec:"required"`
	BlockHash   *common.Hash `json:"blockHash" rlp:"-"`

	TxHash  *common.Hash  `json:"hash"     rlp:"-"`
	TxIndex *hexutil.Uint `json:"transactionIndex"`

	CallFrom *common.Address `json:"from"`
	CallTo   *common.Address `json:"to"       rlp:"nil"`

	GasPrice *hexutil.Big `json:"gasPrice" gencodec:"required"`

	EthAmount *hexutil.Big   `json:"value"    gencodec:"required"`
	Payload   *hexutil.Bytes `json:"input"    gencodec:"required"`
}

type RpcHeader struct {
	Number    *hexutil.Big  `json:"number"`
	BlockHash *common.Hash  `json:"hash"`
	Timestamp *hexutil.Uint `json:"timestamp"`
}

type RpcBlock struct {
	Transactions []*RpcTx        `json:"transactions"`
	BlockNumber  *hexutil.Big    `json:"number"`
	BlockHash    *common.Hash    `json:"hash"`
	Timestamp    *hexutil.Big    `json:"timestamp"`
	Miner        *common.Address `json:"miner"`
}
