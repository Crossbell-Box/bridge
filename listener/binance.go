package listener

import (
	"context"
	"fmt"
	"math"
	"math/big"

	"github.com/ashwanthkumar/slack-go-webhook"
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
	log.Info("[BinanceListener] WithdrewDone2SlackCallback", "tx", tx.GetHash().Hex())
	mainchainEvent := new(mainchainGateway.MainchainGatewayWithdrew)
	mainchainGatewayAbi, err := mainchainGateway.MainchainGatewayMetaData.GetAbi()
	if err != nil {
		return err
	}

	if err = l.utilsWrapper.UnpackLog(*mainchainGatewayAbi, mainchainEvent, "Withdrew", data); err != nil {
		return err
	}

	if l.config.SlackUrl != "" {
		webhookUrl := l.config.SlackUrl
		log.Info("[BinanceListener] Sending to slack", "slack hook url", webhookUrl)
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
		attachment1 := slack.Attachment{}
		attachment1.AddAction(slack.Action{Type: "divider"})
		attachment1.AddField(slack.Field{Title: "Event", Value: ":golf:Withdrew"})
		attachment1.AddField(slack.Field{Title: "Mainchain ID", Value: mainchainEvent.ChainId.String()})
		attachment1.AddField(slack.Field{Title: "Withdraw ID", Value: mainchainEvent.WithdrawalId.String()})
		attachment1.AddField(slack.Field{Title: "Amount", Value: fmt.Sprintf("%.18v", (float64(mainchainEvent.Amount.Uint64()) / float64(math.Pow(10, 18))))})
		attachment1.AddField(slack.Field{Title: "Fee", Value: fmt.Sprintf("%.18v", (float64(mainchainEvent.Fee.Uint64()) / float64(math.Pow(10, 18))))})
		attachment1.AddField(slack.Field{Title: "Remainning Quota", Value: remainingQuota.String()})
		attachment1.AddAction(slack.Action{Type: "button", Text: "View Details", Url: fmt.Sprintf("https://testnet.bscscan.com/tx/%s", tx.GetHash().Hex()), Style: "primary"})

		payload := slack.Payload{
			Text:        fmt.Sprintf(":golf:*Successfully <https://testnet.bscscan.com/tx/%s|*Withdrew*> on BSC!*:golf:\n", tx.GetHash().Hex()),
			IconEmoji:   ":monkey_face:",
			Attachments: []slack.Attachment{attachment1},
		}

		error := slack.Send(webhookUrl, "", payload)
		if len(error) > 0 {
			fmt.Printf("error: %s\n", err)
		}
	}
	return nil
}

type BinanceCallBackJob struct {
	*EthCallbackJob
}
