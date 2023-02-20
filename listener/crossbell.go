package listener

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ashwanthkumar/slack-go-webhook"
	bridgeCore "github.com/axieinfinity/bridge-core"
	bridgeCoreModels "github.com/axieinfinity/bridge-core/models"
	bridgeCoreStores "github.com/axieinfinity/bridge-core/stores"
	"github.com/axieinfinity/bridge-core/utils"
	crossbellGateway "github.com/axieinfinity/bridge-v2/bridge-contracts/generated_contracts/crossbellGateway"
	mainchainGateway "github.com/axieinfinity/bridge-v2/bridge-contracts/generated_contracts/mainchainGateway"
	"github.com/axieinfinity/bridge-v2/models"
	"github.com/axieinfinity/bridge-v2/stores"
	"github.com/axieinfinity/bridge-v2/task"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
)

const oneHour = 3600

type CrossbellListener struct {
	*EthereumListener
	bridgeStore stores.BridgeStore
}

func NewCrossbellListener(ctx context.Context, cfg *bridgeCore.LsConfig, helpers utils.Utils, store bridgeCoreStores.MainStore) (*CrossbellListener, error) {
	listener, err := NewEthereumListener(ctx, cfg, helpers, store)
	if err != nil {
		panic(err)
	}
	l := &CrossbellListener{EthereumListener: listener}
	l.bridgeStore = stores.NewBridgeStore(store.GetDB())
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (l *CrossbellListener) NewJobFromDB(job *bridgeCoreModels.Job) (bridgeCore.JobHandler, error) {
	return newJobFromDB(l, job)
}

// StoreMainchainWithdrawCallback stores the receipt to own database for future check from ProvideReceiptSignatureCallback
func (l *CrossbellListener) RequestWithdrewDoneCallback(fromChainId *big.Int, tx bridgeCore.Transaction, data []byte) error {
	log.Info("[CrossbellListener] RequestWithdrewDoneCallback", "tx", tx.GetHash().Hex())
	mainchainEvent := new(mainchainGateway.MainchainGatewayWithdrew)
	mainchainGatewayAbi, err := mainchainGateway.MainchainGatewayMetaData.GetAbi()
	if err != nil {
		return err
	}

	if err = l.utilsWrapper.UnpackLog(*mainchainGatewayAbi, mainchainEvent, "Withdrew", data); err != nil {
		return err
	}

	// store ronEvent to database at withdrawal
	return l.bridgeStore.GetRequestWithdrawalStore().Update(&models.RequestWithdrawal{
		WithdrawalId:          mainchainEvent.WithdrawalId.Int64(),
		MainchainId:           mainchainEvent.ChainId.Int64(),
		Status:                "done",
		WithdrawalTransaction: tx.GetHash().Hex(),
	})
}

// StoreMainchainWithdrawCallback stores the receipt to own database for future check from ProvideReceiptSignatureCallback
func (l *CrossbellListener) StoreCrossbellDepositedCallback(fromChainId *big.Int, tx bridgeCore.Transaction, data []byte) error {
	log.Info("[CrossbellListener] StoreCrossbellDepositedCallback", "tx", tx.GetHash().Hex())
	crossbellEvent := new(crossbellGateway.CrossbellGatewayDeposited)
	crossbellGatewayAbi, err := crossbellGateway.CrossbellGatewayMetaData.GetAbi()
	if err != nil {
		return err
	}

	if err = l.utilsWrapper.UnpackLog(*crossbellGatewayAbi, crossbellEvent, "Deposited", data); err != nil {
		return err
	}
	// store ronEvent to database at withdrawal
	return l.bridgeStore.GetDepositStore().Save(&models.Deposit{
		DepositId:             crossbellEvent.DepositId.Int64(),
		MainchainId:           crossbellEvent.ChainId.Int64(),
		RecipientAddress:      crossbellEvent.Recipient.Hex(), // from address (the address who submits the signatures into mainchain and gets the fee)
		CrossbellTokenAddress: crossbellEvent.Token.Hex(),     // token address on mainchain
		TokenQuantity:         crossbellEvent.Amount.String(),
		Transaction:           tx.GetHash().Hex(),
	})
}

func (l *CrossbellListener) RequestDepositedDoneCallback(fromChainId *big.Int, tx bridgeCore.Transaction, data []byte) error {
	log.Info("[CrossbellListener] StoreCrossbellDepositedCallback", "tx", tx.GetHash().Hex())
	crossbellEvent := new(crossbellGateway.CrossbellGatewayDeposited)
	crossbellGatewayAbi, err := crossbellGateway.CrossbellGatewayMetaData.GetAbi()
	if err != nil {
		return err
	}

	if err = l.utilsWrapper.UnpackLog(*crossbellGatewayAbi, crossbellEvent, "Deposited", data); err != nil {
		return err
	}

	// store ronEvent to database at withdrawal
	return l.bridgeStore.GetRequestDepositStore().Update(&models.RequestDeposit{
		DepositId:   crossbellEvent.DepositId.Int64(),
		MainchainId: crossbellEvent.ChainId.Int64(),
		Status:      "done",
	})
}

// StoreCrossbellDepositedCallback stores the signatures to own database for future check from ProvideReceiptSignatureCallback
func (l *CrossbellListener) StoreBatchSubmitWithdrawalSignatures(fromChainId *big.Int, tx bridgeCore.Transaction, data []byte) error {
	log.Info("[CrossbellListener] StoreBatchSubmitWithdrawalSignatures", "tx", tx.GetHash().Hex())
	crossbellEvent := new(crossbellGateway.CrossbellGatewaySubmitWithdrawalSignature)
	crossbellGatewayAbi, err := crossbellGateway.CrossbellGatewayMetaData.GetAbi()
	if err != nil {
		return err
	}

	if err = l.utilsWrapper.UnpackLog(*crossbellGatewayAbi, crossbellEvent, "SubmitWithdrawalSignature", data); err != nil {
		return err
	}
	// store ronEvent to database at withdrawal
	return l.bridgeStore.GetWithdrawalSignaturesStore().Save(&models.WithdrawalSignatures{
		MainchainId:      crossbellEvent.ChainId.Int64(),
		WithdrawalId:     crossbellEvent.WithdrawalId.Int64(),
		ValidatorAddress: crossbellEvent.Validator.Hex(), // from address (the address who submits the signatures into mainchain and gets the fee)
		Signature:        string(crossbellEvent.Signature),
		Transaction:      tx.GetHash().Hex(),
	})
}

// StoreCrossbellDepositedCallback stores the signatures to own database for future check from ProvideReceiptSignatureCallback
func (l *CrossbellListener) StoreDepositAck(fromChainId *big.Int, tx bridgeCore.Transaction, data []byte) error {
	log.Info("[CrossbellListener] StoreDepositAck", "tx", tx.GetHash().Hex())
	crossbellEvent := new(crossbellGateway.CrossbellGatewayAckDeposit)
	crossbellGatewayAbi, err := crossbellGateway.CrossbellGatewayMetaData.GetAbi()
	if err != nil {
		return err
	}

	if err = l.utilsWrapper.UnpackLog(*crossbellGatewayAbi, crossbellEvent, "AckDeposit", data); err != nil {
		return err
	}
	// store ronEvent to database at withdrawal
	return l.bridgeStore.GetDepositAckStore().Save(&models.DepositAck{
		MainchainId:      crossbellEvent.ChainId.Int64(),
		DepositId:        crossbellEvent.DepositId.Int64(),
		RecipientAddress: crossbellEvent.Recipient.Hex(),
		ValidatorAddress: crossbellEvent.Validator.Hex(), // from address (the address who submits the signatures into mainchain and gets the fee)
		Transaction:      tx.GetHash().Hex(),
	})
}

// StoreCrossbellDepositedCallback stores the signatures to own database for future check from ProvideReceiptSignatureCallback
func (l *CrossbellListener) StoreRequestWithdrawal(fromChainId *big.Int, tx bridgeCore.Transaction, data []byte) error {
	log.Info("[CrossbellListener] StoreRequestWithdrawal", "tx", tx.GetHash().Hex())
	crossbellEvent := new(crossbellGateway.CrossbellGatewayRequestWithdrawal)
	crossbellGatewayAbi, err := crossbellGateway.CrossbellGatewayMetaData.GetAbi()
	if err != nil {
		return err
	}

	if err = l.utilsWrapper.UnpackLog(*crossbellGatewayAbi, crossbellEvent, "RequestWithdrawal", data); err != nil {
		return err
	}

	if l.config.SlackUrl != "" {
		webhookUrl := l.config.SlackUrl
		log.Info("[CrossbellListener] Sending to slack", "slack hook url", webhookUrl)
		attachment1 := slack.Attachment{}
		attachment1.AddField(slack.Field{Title: "Event", Value: ":mega:RequestWithdraw"})
		attachment1.AddField(slack.Field{Title: "Mainchain ID", Value: crossbellEvent.ChainId.String()})
		attachment1.AddField(slack.Field{Title: "Withdraw ID", Value: crossbellEvent.WithdrawalId.String()})
		attachment1.AddField(slack.Field{Title: "Amount", Value: crossbellEvent.Amount.String()})
		attachment1.AddField(slack.Field{Title: "Fee", Value: crossbellEvent.Fee.String()})
		attachment1.AddAction(slack.Action{Type: "button", Text: "View Details", Url: fmt.Sprintf("https://sepolia.etherscan.io/tx/%s", tx.GetHash().Hex()), Style: "primary"})

		payload := slack.Payload{
			Text:        fmt.Sprintf(":mega:*New <https://sepolia.etherscan.io/tx/%s|*withdraw request*> submitted!*:mega:\n", tx.GetHash().Hex()),
			IconEmoji:   ":monkey_face:",
			Attachments: []slack.Attachment{attachment1},
		}

		err := slack.Send(webhookUrl, "", payload)
		if len(err) > 0 {
			fmt.Printf("error: %s\n", err)
		}
	}

	// store ronEvent to database at withdrawal
	return l.bridgeStore.GetRequestWithdrawalStore().Save(&models.RequestWithdrawal{
		MainchainId:           crossbellEvent.ChainId.Int64(),
		WithdrawalId:          crossbellEvent.WithdrawalId.Int64(),
		RecipientAddress:      crossbellEvent.Recipient.Hex(),
		MainchainTokenAddress: crossbellEvent.Token.Hex(), // token address on mainchain
		TokenQuantity:         crossbellEvent.Amount.String(),
		Fee:                   crossbellEvent.Fee.String(),
		Transaction:           tx.GetHash().Hex(),
		Status:                "pending",
		WithdrawalTransaction: "",
	})
}

// StoreCrossbellDepositedCallback stores the signatures to own database for future check from ProvideReceiptSignatureCallback
func (l *CrossbellListener) StoreRequestDeposit(fromChainId *big.Int, tx bridgeCore.Transaction, data []byte) error {
	log.Info("[CrossbellListener] StoreRequestDeposit", "tx", tx.GetHash().Hex())
	mainchainEvent := new(mainchainGateway.MainchainGatewayRequestDeposit)
	crossbellGatewayAbi, err := crossbellGateway.CrossbellGatewayMetaData.GetAbi()
	if err != nil {
		return err
	}

	if err = l.utilsWrapper.UnpackLog(*crossbellGatewayAbi, mainchainEvent, "RequestDeposit", data); err != nil {
		return err
	}
	// store ronEvent to database at withdrawal
	return l.bridgeStore.GetRequestDepositStore().Save(&models.RequestDeposit{
		MainchainId:           mainchainEvent.ChainId.Int64(),
		DepositId:             mainchainEvent.DepositId.Int64(),
		RecipientAddress:      mainchainEvent.Recipient.Hex(),
		CrossbellTokenAddress: mainchainEvent.Token.Hex(), // token address on mainchain
		TokenQuantity:         mainchainEvent.Amount.String(),
		Transaction:           tx.GetHash().Hex(),
		Status:                "pending",
	})
}

func (l *CrossbellListener) IsUpTodate() bool {
	latestBlock, err := l.GetLatestBlock()
	if err != nil {
		log.Error("[CrossbellListener][IsUpTodate] error while get latest block", "err", err, "listener", l.GetName())
		return false
	}
	// true if timestamp is within 1 hour
	distance := uint64(time.Now().Unix()) - latestBlock.GetTimestamp()
	if distance > uint64(oneHour) {
		log.Info("Node is not up-to-date, keep waiting...", "distance (s)", distance, "listener", l.GetName())
		return false
	}
	return true
}

func (l *CrossbellListener) provideReceiptSignature(fromChainId *big.Int, tx bridgeCore.Transaction, data []byte) error {
	var eventName = "RequestWithdrawal"
	crossbellEvent := new(crossbellGateway.CrossbellGatewayRequestWithdrawal)
	crossbellGatewayAbi, err := crossbellGateway.CrossbellGatewayMetaData.GetAbi()
	if err != nil {
		return err
	}
	if err = l.utilsWrapper.UnpackLog(*crossbellGatewayAbi, crossbellEvent, eventName, data); err != nil {
		return err
	}
	log.Info("[CrossbellListener][ProvideReceiptSignatureCallback] result of calling MainchainWithdrew function", "receiptId", crossbellEvent.WithdrawalId.Int64(), "tx", tx.GetHash().Hex())

	// otherwise, create a task for submitting signature
	// get chainID
	chainId, err := l.GetChainID()
	if err != nil {
		return err
	}
	// create task and store to database
	withdrawalTask := &models.Task{
		ChainId:         hexutil.EncodeBig(chainId),
		FromChainId:     hexutil.EncodeBig(fromChainId),
		FromTransaction: tx.GetHash().Hex(),
		Type:            task.WITHDRAWAL_TASK,
		Data:            common.Bytes2Hex(data),
		Retries:         0,
		Status:          task.STATUS_PENDING,
		LastError:       "",
		CreatedAt:       time.Now().Unix(),
	}
	return l.bridgeStore.GetTaskStore().Save(withdrawalTask)
}

func (l *CrossbellListener) ProvideReceiptSignatureCallback(fromChainId *big.Int, tx bridgeCore.Transaction, data []byte) error {
	return l.provideReceiptSignature(fromChainId, tx, data)
}

func (l *CrossbellListener) DepositRequestedCallback(fromChainId *big.Int, tx bridgeCore.Transaction, data []byte) error {
	log.Info("[CrossbellListener] DepositRequestedCallback", "tx", tx.GetHash().Hex())

	// otherwise, create a task for submitting signature
	// get chainID
	chainId, err := l.GetChainID()
	if err != nil {
		return err
	}

	// check whether deposit is done or not
	// Unpack event data
	ethEvent := new(mainchainGateway.MainchainGatewayRequestDeposit)
	ethGatewayAbi, err := mainchainGateway.MainchainGatewayMetaData.GetAbi()
	if err != nil {
		return err
	}

	if err = l.utilsWrapper.UnpackLog(*ethGatewayAbi, ethEvent, "RequestDeposit", data); err != nil {
		return err
	}
	// create caller
	caller, err := crossbellGateway.NewCrossbellGatewayCaller(common.HexToAddress(l.config.Contracts[task.CROSSBELL_GATEWAY_CONTRACT]), l.client)
	if err != nil {
		return err
	}

	log.Info("[CrossbellListener][DepositRequestedCallback] result of calling DepositVoted function", "receiptId", ethEvent.DepositId, "tx", tx.GetHash().Hex())
	if err != nil {
		return err
	}

	// check if current validator has been voted for this deposit or not
	acknowledgementHash, err := caller.GetValidatorAcknowledgementHash(nil, ethEvent.ChainId, ethEvent.DepositId, l.GetValidatorSign().GetAddress())
	if err != nil {
		return err
	}
	voted := int(big.NewInt(0).SetBytes(acknowledgementHash[:]).Uint64()) != 0
	if voted {
		return nil
	}

	// create task and store to database
	depositTask := &models.Task{
		ChainId:         hexutil.EncodeBig(chainId),
		FromChainId:     hexutil.EncodeBig(fromChainId),
		FromTransaction: tx.GetHash().Hex(),
		Type:            task.DEPOSIT_TASK,
		Data:            common.Bytes2Hex(data),
		Retries:         0,
		Status:          task.STATUS_PENDING,
		LastError:       "",
		CreatedAt:       time.Now().Unix(),
	}
	return l.bridgeStore.GetTaskStore().Save(depositTask)
}

type CrossbellCallBackJob struct {
	*EthCallbackJob
}
