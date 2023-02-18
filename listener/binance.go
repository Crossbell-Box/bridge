package listener

import (
	"context"
	"math/big"

	bridgeCore "github.com/axieinfinity/bridge-core"
	bridgeCoreModels "github.com/axieinfinity/bridge-core/models"
	bridgeCoreStores "github.com/axieinfinity/bridge-core/stores"
	"github.com/axieinfinity/bridge-core/utils"
	mainchainGateway "github.com/axieinfinity/bridge-v2/bridge-contracts/generated_contracts/mainchainGateway"
	"github.com/axieinfinity/bridge-v2/stores"
	"github.com/axieinfinity/bridge-v2/task"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type BinanceListener struct {
	*EthereumListener
	bridgeStore stores.BridgeStore
}

func NewBinanceListener(ctx context.Context, cfg *bridgeCore.LsConfig, helpers utils.Utils, store bridgeCoreStores.MainStore) (*BinanceListener, error) {
	listener, err := NewEthereumListener(ctx, cfg, helpers, store)
	if err != nil {
		panic(err)
	}
	l := &BinanceListener{EthereumListener: listener}
	l.bridgeStore = stores.NewBridgeStore(store.GetDB())
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (l *BinanceListener) NewJobFromDB(job *bridgeCoreModels.Job) (bridgeCore.JobHandler, error) {
	return newJobFromDB(l, job)
}

// WithdrewDone2SlackCallback send the withdrew event rto slack channel through slack hook
func (l *BinanceListener) WithdrewDone2SlackCallback(fromChainId *big.Int, tx bridgeCore.Transaction, data []byte) error {
	log.Info("[PolygonListener] WithdrewDone2SlackCallback", "tx", tx.GetHash().Hex())
	mainchainEvent := new(mainchainGateway.MainchainGatewayWithdrew)
	mainchainGatewayAbi, err := mainchainGateway.MainchainGatewayMetaData.GetAbi()
	if err != nil {
		return err
	}

	if err = l.utilsWrapper.UnpackLog(*mainchainGatewayAbi, mainchainEvent, "Withdrew", data); err != nil {
		return err
	}

	if l.config.SlackUrl != "" {
		log.Info("[Slack Hook] Sending Withdrew to slack", "tx", tx.GetHash().Hex())

		// create caller
		caller, err := mainchainGateway.NewMainchainGatewayCaller(common.HexToAddress(l.config.Contracts[task.MAINCHAIN_GATEWAY_CONTRACT]), l.client)
		if err != nil {
			return err
		}
		// check remaining quota
		remainingQuota, err := caller.GetDailyWithdrawalRemainingQuota(nil, mainchainEvent.Token)
		if err != nil {
			log.Error("[Slack hook] error while querying remainingQuota ", "error", err)
			return err
		}

		RequestWithdrawInfo := RequestWithdrawInfo{
			MainchainId:      mainchainEvent.ChainId.Int64(),
			WithdrawalId:     mainchainEvent.WithdrawalId.Int64(),
			FromAddress:      tx.GetFromAddress(),
			RecipientAddress: mainchainEvent.Recipient.Hex(),
			TokenQuantity:    mainchainEvent.Amount.String(),
			Fee:              mainchainEvent.Fee.String(),
			Transaction:      tx.GetHash().Hex(),
			RemainingQuota:   remainingQuota.String(),
		}
		if err = ReqPostRequestWithdraw(l.config.SlackUrl, RequestWithdrawInfo, tx, ":mega::mega::mega:New withdrawal completed!!!", ":o:"); err != nil {
			log.Error("[Slack hook] error while sending post req to slack", "error", err)
			return err
		}
	}
	return nil
}

type BinanceCallBackJob struct {
	*EthCallbackJob
}
