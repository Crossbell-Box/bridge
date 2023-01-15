package internal

import (
	"context"
	"fmt"

	bridgeCore "github.com/axieinfinity/bridge-core"
	"github.com/axieinfinity/bridge-core/metrics"
	bridgeCoreStores "github.com/axieinfinity/bridge-core/stores"
	"github.com/axieinfinity/bridge-core/utils"
	"github.com/axieinfinity/bridge-v2/listener"
	roninTask "github.com/axieinfinity/bridge-v2/task"
	"github.com/ethereum/go-ethereum/log"
	"gorm.io/gorm"
)

type BridgeController struct {
	*bridgeCore.Controller
}

func NewBridgeController(cfg *bridgeCore.Config, db *gorm.DB, helpers utils.Utils) (*BridgeController, error) {
	bridgeCore.AddListener("Ethereum", InitEthereum)
	bridgeCore.AddListener("Crossbell", InitCrossbell)
	bridgeCore.AddListener("Binance", InitBinance)
	bridgeCore.AddListener("Polygon", InitPolygon)
	controller, err := bridgeCore.New(cfg, db, helpers)
	if err != nil {
		return nil, err
	}
	return &BridgeController{controller}, nil
}

func InitEthereum(ctx context.Context, lsConfig *bridgeCore.LsConfig, store bridgeCoreStores.MainStore, helpers utils.Utils) bridgeCore.Listener {
	ethListener, err := listener.NewEthereumListener(ctx, lsConfig, helpers, store)
	if err != nil {
		log.Error("[EthereumListener]Error while init new ethereum listener", "err", err)
		return nil
	}
	metrics.Pusher.AddCounter(fmt.Sprintf(metrics.ListenerProcessedBlockMetric, ethListener.GetName()), "count number of processed block in ethereum listener")
	return ethListener
}

func InitCrossbell(ctx context.Context, lsConfig *bridgeCore.LsConfig, store bridgeCoreStores.MainStore, helpers utils.Utils) bridgeCore.Listener {
	crossbellLinListener, err := listener.NewCrossbellListener(ctx, lsConfig, helpers, store)
	if err != nil {
		log.Error("[CrossbellListener]Error while init new crossbell listener", "err", err)
		return nil
	}
	metrics.Pusher.AddCounter(fmt.Sprintf(metrics.ListenerProcessedBlockMetric, crossbellLinListener.GetName()), "count number of processed block in ethereum listener")

	task, err := roninTask.NewRoninTask(crossbellLinListener, store.GetDB(), helpers)
	if err != nil {
		log.Error("[CrossbellListener][InitCrossbell] Error while adding new task", "err", err)
		return nil
	}
	crossbellLinListener.AddTask(task)
	return crossbellLinListener
}

func InitBinance(ctx context.Context, lsConfig *bridgeCore.LsConfig, store bridgeCoreStores.MainStore, helpers utils.Utils) bridgeCore.Listener {
	binanceListener, err := listener.NewEthereumListener(ctx, lsConfig, helpers, store)
	if err != nil {
		log.Error("[BinanceListener]Error while init new binance listener", "err", err)
		return nil
	}
	metrics.Pusher.AddCounter(fmt.Sprintf(metrics.ListenerProcessedBlockMetric, binanceListener.GetName()), "count number of processed block in ethereum listener")
	return binanceListener
}

func InitPolygon(ctx context.Context, lsConfig *bridgeCore.LsConfig, store bridgeCoreStores.MainStore, helpers utils.Utils) bridgeCore.Listener {
	polygonListener, err := listener.NewEthereumListener(ctx, lsConfig, helpers, store)
	if err != nil {
		log.Error("[PolygonListener]Error while init new Polygon listener", "err", err)
		return nil
	}
	metrics.Pusher.AddCounter(fmt.Sprintf(metrics.ListenerProcessedBlockMetric, polygonListener.GetName()), "count number of processed block in ethereum listener")
	return polygonListener
}
