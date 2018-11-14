package selfconf

import (
	"math/big"
)

const (
        // 日志级别 
	LogLevel             = "debug"

	// 平均出块时间，单位秒。公链约在15秒，kovan约在5秒
	EthBlockInterval int = 15

        // 以太坊节点的websocket连接地址
	EstimatorWsUrl                = "ws://1.1.0.0:1"

	// leveldb存储目录
	EstimatorStore                = "/data/estimator/store"

	// mongodb连接地址
	EstimatorMgoUri               = "1.1.0.0:2"

	// mongodb连接用户
	EstimatorMgoUser              = "abc"

	// mongodb连接密码
	EstimatorMgoPwd               = "abc"

	// mongodb使用的库名
	EstimatorMgoDB                = "abc"

	// mongodb使用的表名
	EstimatorMgoTable             = "abc"

	// gasPrice 刷新时间间隔，单位分钟。
	EstimatorTickerInterval int64 = 2

	// 预估gasPrice时所使用的进块概率，为了排除那种只收容高price的miner，可以设在七八十的样子，需要运行一段时间观察情况
	EstimatorBeMinedProb    int64 = 70

	// 块所需的确认数。这个在这儿只是为了程序好处理，可以不用管。
	EstimatorConfirms       int64 = 1
)

var (
        // 监听的起始高度。对于gasPrice需要扫描约5000个块用于统计，因而在程序编译时设置值应低于那时高度约5000块。
	EstimatorStart = new(big.Int).SetInt64(6680000)
)
