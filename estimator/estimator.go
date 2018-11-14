package estimator

import (
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"sync"
	"time"

	"selfconf"

	"github.com/antimoth/ethparser/client"
	"github.com/antimoth/ethparser/log"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	MONGO_DOC_ID = "ethereum_estimated_gas_price"
)

var (
	estLogger           = log.NewLogger(selfconf.LogLevel)
	DELETED_HEIGHT_SLOT = []byte("hadBeenDeletedHeight")
)

type Estimator struct {
	c             *client.Client
	confirmHeight *big.Int
	currentHeight *big.Int
	PrePrice      *big.Int
	deletedHeight *big.Int
	priceList     BlockPriceList
	lock          *sync.Mutex
	deleteLock    *sync.Mutex
	mgo           *mgo.Session
}

func NewEstimator() (*Estimator, error) {
	wsUrl := selfconf.EstimatorWsUrl
	confirms := selfconf.EstimatorConfirms
	ldbPath := selfconf.EstimatorStore

	estLogger.Info("pre connected", "wsUrl", wsUrl, "confirms", confirms, "ldbPath", ldbPath)

	mgoUri := selfconf.EstimatorMgoUri
	mgoUserName := selfconf.EstimatorMgoUser
	mgoPassWord := selfconf.EstimatorMgoPwd

	estLogger.Info("pre connected", "mgoUri", mgoUri, "mgoUserName", mgoUserName)

	return DialEstimator(wsUrl, ldbPath, mgoUri, mgoUserName, mgoPassWord, uint64(confirms))
}

func DialEstimator(wsUrl, ldbPath, mgoUri, mgoUserName, mgoPassWord string, confirms uint64) (*Estimator, error) {

	c, err := client.NewClient(wsUrl)
	if err != nil {
		return nil, err
	}

	if confirms < uint64(1) {
		confirms = uint64(1)
	}

	mgoSession, err := mgo.Dial(mgoUri)
	if err != nil {
		return nil, err
	}
	mgoSession.Login(&mgo.Credential{Username: mgoUserName, Password: mgoPassWord})
	mgoSession.SetMode(mgo.Monotonic, true)

	return &Estimator{
		c:             c,
		confirmHeight: new(big.Int).SetUint64(confirms),
		currentHeight: new(big.Int),
		PrePrice:      new(big.Int),
		deletedHeight: new(big.Int),
		lock:          new(sync.Mutex),
		deleteLock:    new(sync.Mutex),
		mgo:           mgoSession,
	}, nil
}

func (et *Estimator) Close() {
	et.c.Close()
	et.mgo.Close()
}

func (et *Estimator) GetMinPriceFromLatest1Q() *big.Int {
	sumPrice := new(big.Int)
	for ix, blockPrice := range et.priceList {
		if ix >= 1000 {
			break
		} else {
			sumPrice.Add(sumPrice, blockPrice.MinPrice)
		}
	}
	return sumPrice.Div(sumPrice, new(big.Int).SetInt64(1000))
}

func (et *Estimator) GetBlockPriceFromHeight(height *big.Int) *BlockPrice {
	blockInfo, err := et.c.GetBlockInfoFromHeight(height)
	if err != nil {
		estLogger.Error("GetBlockInfoFromHeight error", "height", height.Uint64(), "error", err)
		return &BlockPrice{
			Miner:       new(common.Address),
			BlockNumber: height,
			MinPrice:    et.GetMinPriceFromLatest1Q(),
			Timestamp:   new(big.Int).SetInt64(time.Now().Unix()),
		}
	}

	miner := blockInfo.Miner.Hex()
	var minPrice *big.Int

	for _, tx := range blockInfo.Transactions {
		if tx.CallFrom.Hex() == miner {
			continue
		}

		if minPrice == nil {
			minPrice = tx.GasPrice.ToInt()

		} else if cur := tx.GasPrice.ToInt(); cur.Int64() > 0 && cur.Cmp(minPrice) < 0 {
			minPrice = cur
		}
	}

	if minPrice == nil {
		minPrice = et.GetMinPriceFromLatest1Q()
	}

	return &BlockPrice{Miner: blockInfo.Miner, BlockNumber: height, MinPrice: minPrice, Timestamp: blockInfo.Timestamp.ToInt()}
}

func (et *Estimator) reviewBlock(start *big.Int) {
	cur, err := et.c.BlockNumber()
	if err != nil {
		panic("get current height error!")
	}
	et.currentHeight = cur

	curConfirm := new(big.Int).Sub(cur, et.confirmHeight)
	increaser := big.NewInt(1)
	go func() {
		for i := curConfirm; i.Int64() >= 0; i = new(big.Int).Sub(i, increaser) {
			blockPrice := et.GetBlockPriceFromHeight(i)
			et.lock.Lock()
			et.priceList = append(et.priceList, blockPrice)
			et.lock.Unlock()

			if len(et.priceList) >= 4999 {
				break
			}
		}
	}()
}

func (et *Estimator) StartWatch(start *big.Int) {
	et.deletedHeight = start

	wCh := make(chan *client.RpcHeader, 1000)

	sub, err := et.SubscribeNewHead(wCh)
	if err != nil {
		panic(fmt.Sprintf("create sub new blocks error! e is %v!", err.Error()))
	}

	et.reviewBlock(start)

	bigConfirmH := et.confirmHeight
	increaser := big.NewInt(1)

	go func() {
		defer sub.Unsubscribe()

		for {
		LoopBlocks:
			select {
			case blockHeader := <-wCh:
				bigIntNumber := (*big.Int)(blockHeader.Number)
				if bigIntNumber.Cmp(et.currentHeight) > 0 {
					startH := new(big.Int).Add(et.currentHeight, increaser)

					for i := startH; i.Cmp(bigIntNumber) <= 0; i = new(big.Int).Add(i, increaser) {
						pushH := new(big.Int).Sub(i, bigConfirmH)
						blockPrice := et.GetBlockPriceFromHeight(pushH)
						et.lock.Lock()
						et.priceList = CopyInsert(et.priceList, 0, blockPrice).(BlockPriceList)
						if len(et.priceList) > 5000 {
							et.priceList[5000] = nil
							et.priceList = et.priceList[:5000]
						}
						et.lock.Unlock()
						// estLogger.Debug("watched ethereum block", "height", pushH.Uint64())
					}

					et.currentHeight = bigIntNumber

				} else {
					// estLogger.Warn("receive ethereum block height under current", "height", bigIntNumber.Uint64(), "current", et.currentHeight.Uint64())
				}

			case err := <-sub.Err():
				estLogger.Error("sub new blocks error", "error", err.Error())

				reConnectTimes := 1
				tiker := time.NewTicker(time.Second * 10)
				for {
					select {
					case <-tiker.C:
						sub, err = et.SubscribeNewHead(wCh)
						if err == nil {
							estLogger.Info("sub new blocks reconnected!")
							tiker.Stop()
							tiker = nil
							goto LoopBlocks

						} else {
							estLogger.Error("sub new blocks reconnect error", "error", err, "tryTimes", reConnectTimes)
							reConnectTimes += 1
						}
					}
				}
			}
		}
	}()
}

func (et *Estimator) StatisticMinerData(needBlocks int) (MinerDataList, error) {
	if len(et.priceList) < needBlocks {
		return nil, ErrNoEnoughBlocks
	}

	minerDataMap := make(map[string]*MinerData)
	for ix, blockPrice := range et.priceList {
		if ix >= needBlocks {
			break
		}
		if minerData, ok := minerDataMap[blockPrice.Miner.Hex()]; ok {
			minerData.MinedBlockNum += 1
			if blockPrice.MinPrice.Cmp(minerData.MinPrice) < 0 {
				minerData.MinPrice = blockPrice.MinPrice
			}
		} else {
			minerDataMap[blockPrice.Miner.Hex()] = &MinerData{Miner: blockPrice.Miner, MinedBlockNum: 1, MinPrice: blockPrice.MinPrice}
		}
	}

	var dataList MinerDataList
	for _, data := range minerDataMap {
		dataList = append(dataList, data)
	}
	sort.Sort(dataList)
	return dataList, nil
}

func (et *Estimator) GenerateProbabilities(dataList MinerDataList, waitBlocks int64, needBlocks int) (ProbabilityList, error) {

	var proList ProbabilityList
	for i := 0; i < len(dataList); i++ {
		minPrice := dataList[i].MinPrice
		var mayAcceptBlocks int = 0
		for j := i; j < len(dataList); j++ {
			mayAcceptBlocks += dataList[j].MinedBlockNum
		}

		proPerBlock := decimal.New(int64(needBlocks-mayAcceptBlocks), 0).Div(decimal.New(int64(needBlocks), 0))

		// pro := decimal.New(1, 0).Sub(proPerBlock.Pow(decimal.New(waitBlocks/1000, 0)))
		pro := decimal.New(1, 0).Sub(proPerBlock)
		proList = append(proList, &Probability{GasPrice: minPrice, Prob: pro})
	}
	return proList, nil
}

func (et *Estimator) ComputeGasPrice(probList ProbabilityList, desiredProb decimal.Decimal) (*big.Int, error) {
	first := probList[0]
	last := probList[len(probList)-1]
	if desiredProb.Cmp(first.Prob) >= 0 {
		return first.GasPrice, nil

	} else if desiredProb.Cmp(last.Prob) <= 0 {
		return last.GasPrice, nil
	}

	for i := 0; i < len(probList)-1; i++ {
		left := probList[i]
		right := probList[i+1]
		if desiredProb.Cmp(right.Prob) < 0 {
			continue

		} else if desiredProb.Cmp(left.Prob) > 0 {
			return new(big.Int), ErrInvalidData
		}

		adjProb := desiredProb.Sub(right.Prob)
		windowSize := left.Prob.Sub(right.Prob)
		position := adjProb.Div(windowSize)
		gasWindowSize := new(big.Int).Sub(left.GasPrice, right.GasPrice)
		gasPrice := new(big.Int).Add(right.GasPrice, position.Mul(decimal.NewFromBigInt(gasWindowSize, 0)).Ceil().Coefficient())
		return gasPrice, nil
	}
	return new(big.Int), ErrInvalidData
}

func (et *Estimator) GasPriceEstimator(maxWaitSecs int) (*big.Int, error) {
	needBlocks := maxWaitSecs * 5 / selfconf.EthBlockInterval
	if len(et.priceList) < needBlocks {
		return new(big.Int), ErrNoEnoughBlocks
	}

	avgBlockTime := decimal.NewFromBigInt(new(big.Int).Sub(et.priceList[0].Timestamp, et.priceList[needBlocks-1].Timestamp), 0).Div(decimal.New(int64(needBlocks), 0))
	waitBlocks := decimal.New(int64(maxWaitSecs), 0).Div(avgBlockTime).IntPart()

	minerDataList, err := et.StatisticMinerData(needBlocks)

	if err != nil {
		return new(big.Int), err
	}

	probList, err := et.GenerateProbabilities(minerDataList, waitBlocks, needBlocks)
	if err != nil {
		return new(big.Int), err
	}

	return et.ComputeGasPrice(probList, decimal.New(selfconf.EstimatorBeMinedProb, 0).Div(decimal.New(100, 0)))
}

func (et *Estimator) WatchPendingTx(ch chan<- *common.Hash) {
	txCh := make(chan *common.Hash, 1000)

	sub, err := et.SubscribePendingTx(txCh)
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
				estLogger.Error("sub pending tranx error!", "error", err.Error())

				reConnectTimes := 1
				tiker := time.NewTicker(time.Second * 10)

				for {
					select {
					case <-tiker.C:
						sub, err = et.SubscribePendingTx(txCh)

						if err == nil {
							estLogger.Info("sub pending tranx reconnected!")
							tiker.Stop()
							tiker = nil
							goto LoopTranx

						} else {
							estLogger.Error("sub pending tranx reconnect error", "error", err, "tryTimes", reConnectTimes)
							reConnectTimes += 1
						}
					}
				}
			}
		}
	}()
}

func (et *Estimator) RunEstimator() {
	tiker := time.NewTicker(time.Minute * time.Duration(selfconf.EstimatorTickerInterval))
	go func() {
		for {
			select {
			case <-tiker.C:
				price10min, err := et.GasPriceEstimator(180)
				if err != nil {
					break
				}
				if et.PrePrice.Cmp(price10min) < 0 {
					price10min = decimal.NewFromBigInt(price10min, 0).Mul(decimal.NewFromFloat(1.1)).Ceil().Coefficient()
				}
				et.PrePrice = price10min

				var data []Estimated
				data = append(data, Estimated{Latency: 600, Price: price10min.Uint64()})

				price30min, err := et.GasPriceEstimator(600)
				if err != nil || price30min.Cmp(price10min) >= 0 {
					price30min = decimal.NewFromBigInt(price10min, 0).Mul(decimal.NewFromFloat(0.95)).Ceil().Coefficient()
				}
				data = append(data, Estimated{Latency: 1800, Price: price30min.Uint64()})

				price2h, err := et.GasPriceEstimator(1800)
				if err != nil || price2h.Cmp(price30min) >= 0 {
					price2h = decimal.NewFromBigInt(price30min, 0).Mul(decimal.NewFromFloat(0.95)).Ceil().Coefficient()
				}
				data = append(data, Estimated{Latency: 7200, Price: price2h.Uint64()})

				price12h, err := et.GasPriceEstimator(3600)
				if err != nil || price12h.Cmp(price2h) >= 0 {
					price12h = decimal.NewFromBigInt(price2h, 0).Mul(decimal.NewFromFloat(0.9)).Ceil().Coefficient()
				}
				data = append(data, Estimated{Latency: 43200, Price: price12h.Uint64()})

				price1d, err := et.GasPriceEstimator(7200)
				if err != nil || price1d.Cmp(price12h) >= 0 {
					price1d = decimal.NewFromBigInt(price12h, 0).Mul(decimal.NewFromFloat(0.8)).Ceil().Coefficient()
				}
				data = append(data, Estimated{Latency: 86400, Price: price1d.Uint64()})

				et.mgo.DB(selfconf.EstimatorMgoDB).C(selfconf.EstimatorMgoTable).Upsert(bson.M{"id": MONGO_DOC_ID}, &EstimatedList{ID: MONGO_DOC_ID, PriceList: data})

				estLogger.Info("price", "price10min", price10min.String(), "price30min", price30min.String(), "price2h", price2h, "price12h", price12h, "price1d", price1d)
			}
		}
	}()
}

func (et *Estimator) SubscribeNewHead(ch chan<- *client.RpcHeader) (ethereum.Subscription, error) {
	return et.c.EthSubscribe(ch, "newHeads")
}

func (et *Estimator) SubscribePendingTx(ch chan<- *common.Hash) (ethereum.Subscription, error) {
	return et.c.EthSubscribe(ch, "newPendingTransactions")
}

func CopyInsert(slice interface{}, pos int, value interface{}) interface{} {
	v := reflect.ValueOf(slice)
	v = reflect.Append(v, reflect.ValueOf(value))
	reflect.Copy(v.Slice(pos+1, v.Len()), v.Slice(pos, v.Len()))
	v.Index(pos).Set(reflect.ValueOf(value))
	return v.Interface()
}
