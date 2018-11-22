package estimator

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

var (
	ErrNoEnoughBlocks = errors.New("no enough blocks to statistic, please wait a moment.")
	ErrInvalidData    = errors.New("invalid statistic data")
	zeroInt           = big.NewInt(0)
	oneInt            = big.NewInt(1)
	tenInt            = big.NewInt(10)
)

func RoundBigInt(x *big.Int, places int64) *big.Int {
	div := new(big.Int).Exp(tenInt, big.NewInt(places), nil)
	rounded, m := new(big.Int).DivMod(x, div, new(big.Int))
	if m.Cmp(zeroInt) != 0 {
		rounded = rounded.Add(rounded, oneInt)
	}
	return rounded.Mul(rounded, div)
}

type MinerData struct {
	Miner         *common.Address
	MinedBlockNum int
	MinPrice      *big.Int
}

type MinerDataList []*MinerData

func (mdl MinerDataList) Len() int           { return len(mdl) }
func (mdl MinerDataList) Swap(i, j int)      { mdl[i], mdl[j] = mdl[j], mdl[i] }
func (mdl MinerDataList) Less(i, j int) bool { return mdl[i].MinPrice.Cmp(mdl[j].MinPrice) > 0 }

type Probability struct {
	GasPrice *big.Int
	Prob     decimal.Decimal
}

type ProbabilityList []*Probability

type BlockPrice struct {
	Miner       *common.Address
	BlockNumber *big.Int
	MinPrice    *big.Int
	Timestamp   *big.Int
}

type BlockPriceList []*BlockPrice

type Estimated struct {
	Latency int64
	Price   uint64
}

type EstimatedList struct {
	ID        string
	PriceList []Estimated
}
