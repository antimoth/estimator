package watcher

import (
	"fmt"
	"math/big"
	"time"

	"github.com/antimoth/ethparser/client"
	"github.com/antimoth/ethparser/log"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("loglevel", "info")
}

var (
	eLogger = log.NewLogger(viper.GetString("loglevel"))
)

type Watcher struct {
	c             *client.Client
	confirmHeight *big.Int
	currentHeight *big.Int
}

func NewWatcher(rawurl string, confirmHeight uint64) (*Watcher, error) {

	eLogger = log.NewLogger(viper.GetString("loglevel"))

	c, err := client.NewClient(rawurl)
	if err != nil {
		return nil, err
	}

	if confirmHeight < uint64(1) {
		confirmHeight = uint64(1)
	}

	return &Watcher{
		c:             c,
		confirmHeight: new(big.Int).SetUint64(confirmHeight),
		currentHeight: new(big.Int),
	}, nil
}

func (ew *Watcher) Close() {
	ew.c.Close()
}

func (ew *Watcher) GetClient() *client.Client {
	return ew.c
}

func (ew *Watcher) ReviewBlock(start *big.Int, ch chan<- *big.Int) {
	cur, err := ew.c.BlockNumber()
	if err != nil {
		panic("get current height error!")
	}
	ew.currentHeight = cur

	curConfirm := new(big.Int).Sub(cur, ew.confirmHeight)
	increaser := big.NewInt(1)
	go func() {
		for i := start; i.Cmp(curConfirm) <= 0; i = new(big.Int).Add(i, increaser) {
			ch <- i
		}
		close(ch)
	}()
}

func (ew *Watcher) WatchBlock(start *big.Int, heightCh chan<- *big.Int) {
	wCh := make(chan *client.RpcHeader, 1000)
	rCh := make(chan *big.Int, 1000)

	sub, err := ew.SubscribeNewHead(wCh)
	if err != nil {
		panic(fmt.Sprintf("create sub new blocks error! e is %v!", err.Error()))
	}

	ew.ReviewBlock(start, rCh)

	bigConfirmH := ew.confirmHeight
	increaser := big.NewInt(1)

	go func() {
		defer sub.Unsubscribe()

		for blockHeight := range rCh {
			select {
			case heightCh <- blockHeight:
			}
		}

		for {
		LoopBlocks:
			select {
			case blockHeader := <-wCh:
				bigIntNumber := (*big.Int)(blockHeader.Number)
				if bigIntNumber.Cmp(ew.currentHeight) > 0 {
					startH := new(big.Int).Add(ew.currentHeight, increaser)

					for i := startH; i.Cmp(bigIntNumber) <= 0; i = new(big.Int).Add(i, increaser) {
						pushH := new(big.Int).Sub(i, bigConfirmH)
						select {
						case heightCh <- pushH:
							eLogger.Debug("watched ethereum block", "height", pushH.Uint64())
						}
					}

					ew.currentHeight = bigIntNumber

				} else {
					eLogger.Warn("receive ethereum block height under current", "height", bigIntNumber.Uint64(), "current", ew.currentHeight.Uint64())
				}

			case err := <-sub.Err():
				eLogger.Error("sub new blocks error", "error", err.Error())

				reConnectTimes := 1
				tiker := time.NewTicker(time.Second * 10)
				for {
					select {
					case <-tiker.C:
						sub, err = ew.SubscribeNewHead(wCh)
						if err == nil {
							eLogger.Info("sub new blocks reconnected!")
							tiker.Stop()
							tiker = nil
							goto LoopBlocks

						} else {
							eLogger.Error("sub new blocks reconnect error", "error", err, "tryTimes", reConnectTimes)
							reConnectTimes += 1
						}
					}
				}
			}
		}
	}()
}

func (ew *Watcher) WatchPendingTx(ch chan<- *common.Hash) {
	txCh := make(chan *common.Hash, 1000)

	sub, err := ew.SubscribePendingTx(txCh)
	if err != nil {
		panic(fmt.Sprintf("create sub pending tranx error! e is %v!", err.Error()))
	}

	go func() {
		defer sub.Unsubscribe()

		for {
		LoopTranx:
			select {
			case txHash := <-txCh:
				ch <- txHash

			case err := <-sub.Err():
				eLogger.Error("sub pending tranx error!", "error", err.Error())

				reConnectTimes := 1
				tiker := time.NewTicker(time.Second * 10)

				for {
					select {
					case <-tiker.C:
						sub, err = ew.SubscribePendingTx(txCh)

						if err == nil {
							eLogger.Info("sub pending tranx reconnected!")
							tiker.Stop()
							tiker = nil
							goto LoopTranx

						} else {
							eLogger.Error("sub pending tranx reconnect error", "error", err, "tryTimes", reConnectTimes)
							reConnectTimes += 1
						}
					}
				}
			}
		}
	}()
}

func (ew *Watcher) SubscribeNewHead(ch chan<- *client.RpcHeader) (ethereum.Subscription, error) {
	return ew.c.EthSubscribe(ch, "newHeads")
}

func (ew *Watcher) SubscribePendingTx(ch chan<- *common.Hash) (ethereum.Subscription, error) {
	return ew.c.EthSubscribe(ch, "newPendingTransactions")
}
