package listener

import (
	"context"
	"math/big"
	"time"

	crossbellGateway "github.com/Crossbell-Box/bridge-contracts/generated_contracts/crossbell/gateway"
	crossbellValidator "github.com/Crossbell-Box/bridge-contracts/generated_contracts/crossbell/governance"
	mainchainGateway "github.com/Crossbell-Box/bridge-contracts/generated_contracts/mainchain/gateway"
	bridgeCore "github.com/axieinfinity/bridge-core"
	bridgeCoreModels "github.com/axieinfinity/bridge-core/models"
	bridgeCoreStores "github.com/axieinfinity/bridge-core/stores"
	"github.com/axieinfinity/bridge-core/utils"
	"github.com/axieinfinity/bridge-v2/models"
	"github.com/axieinfinity/bridge-v2/stores"
	"github.com/axieinfinity/bridge-v2/task"
	bridgeUtils "github.com/axieinfinity/bridge-v2/utils"
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
func (l *CrossbellListener) StoreMainchainWithdrewCallback(fromChainId *big.Int, tx bridgeCore.Transaction, data []byte) error {
	log.Info("[CrossbellListener] StoreMainchainWithdrawCallback", "tx", tx.GetHash().Hex())
	mainchainEvent := new(mainchainGateway.MainchainGatewayWithdrew)
	mainchainGatewayAbi, err := mainchainGateway.MainchainGatewayMetaData.GetAbi()
	if err != nil {
		return err
	}

	if err = l.utilsWrapper.UnpackLog(*mainchainGatewayAbi, mainchainEvent, "Withdrew", data); err != nil {
		return err
	}
	// store ronEvent to database at withdrawal
	return l.bridgeStore.GetWithdrawalStore().Save(&models.Withdrawal{
		WithdrawalId:         mainchainEvent.WithdrawalId.Int64(),
		ExternalAddress:      tx.GetFromAddress(),        // from address (the address who submits the signatures into mainchain and gets the fee)
		ExternalTokenAddress: mainchainEvent.Token.Hex(), // token address on mainchain
		ExternalChainId:      mainchainEvent.ChainId.Int64(),
		RecipientAddress:     mainchainEvent.Recipient.Hex(),
		// RoninTokenAddress:    mainchainEvent.Token.Hex(), // how to get token address on crossbell
		TokenQuantity: mainchainEvent.Amount.String(),
		Transaction:   tx.GetHash().Hex(),
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

func (l *CrossbellListener) provideReceiptSignatureAgain(fromChainId *big.Int, tx bridgeCore.Transaction, data []byte) error {
	var eventName string
	eventName = "RequestWithdrawalSignatures"
	crossbellEvent := new(crossbellGateway.CrossbellGatewayRequestWithdrawalSignatures)
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
		Type:            task.WITHDRAWAL_AGAIN_TASK,
		Data:            common.Bytes2Hex(data),
		Retries:         0,
		Status:          task.STATUS_PENDING,
		LastError:       "",
		CreatedAt:       time.Now().Unix(),
	}
	return l.bridgeStore.GetTaskStore().Save(withdrawalTask)
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

func (l *CrossbellListener) ProvideReceiptSignatureAgainCallback(fromChainId *big.Int, tx bridgeCore.Transaction, data []byte) error {
	return l.provideReceiptSignatureAgain(fromChainId, tx, data)
}

func (l *CrossbellListener) DepositRequestedCallback(fromChainId *big.Int, tx bridgeCore.Transaction, data []byte) error {
	log.Info("[CrossbellListener] DepositRequestedCallback", "tx", tx.GetHash().Hex())
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
	// get chainID
	chainId, err := l.GetChainID()
	if err != nil {
		return err
	}
	log.Info("[CrossbellListener][DepositRequestedCallback] result of calling DepositVoted function", "receiptId", ethEvent.DepositId, "tx", tx.GetHash().Hex())
	if err != nil {
		return err
	}

	// check if current validator has been voted for this deposit or not
	acknowledgementHash, err := caller.GetValidatorAcknowledgementHash(nil, chainId, ethEvent.DepositId, l.GetValidatorSign().GetAddress())
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

func (l *CrossbellListener) isValidatorNode() (bool, error) {
	validatorSign := l.GetValidatorSign()
	if validatorSign == nil {
		log.Warn("The current node is not set validator key")
		return false, nil
	}

	ethClient := l.GetListener(bridgeUtils.Ethereum).GetEthClient()
	crossbellGovernanceCaller, err := crossbellValidator.NewValidatorCaller(common.HexToAddress(l.config.Contracts[task.CROSSBELL_VALIDATOR]), ethClient)
	if err != nil {
		return false, err
	}

	addr := validatorSign.GetAddress()
	isRelayer, err := crossbellGovernanceCaller.IsValidator(nil, addr)
	if err != nil {
		return false, err
	}

	return isRelayer, nil
}

type CrossbellCallBackJob struct {
	*EthCallbackJob
}
