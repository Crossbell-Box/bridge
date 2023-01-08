package task

import (
	"crypto/ecdsa"
	"math/big"
	"time"

	roninGovernance "github.com/axieinfinity/bridge-contracts/generated_contracts/ronin/governance"

	crossbellGateway "github.com/Crossbell-Box/bridge-contracts/generated_contracts/crossbell/gateway"
	mainchainGateway "github.com/Crossbell-Box/bridge-contracts/generated_contracts/mainchain/gateway"
	"github.com/axieinfinity/bridge-v2/stores"
	"github.com/ethereum/go-ethereum/signer/core"

	roninGateway "github.com/axieinfinity/bridge-contracts/generated_contracts/ronin/gateway"
	bridgeCore "github.com/axieinfinity/bridge-core"
	"github.com/axieinfinity/bridge-core/metrics"
	"github.com/axieinfinity/bridge-core/utils"
	"github.com/axieinfinity/bridge-v2/models"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

const (
	ErrNotBridgeOperator = "execution reverted: RoninGatewayV2: unauthorized sender"
)

type bulkTask struct {
	util           utils.Utils
	tasks          []*models.Task
	store          stores.BridgeStore
	validator      *ecdsa.PrivateKey
	client         *ethclient.Client
	contracts      map[string]string
	chainId        *big.Int
	maxTry         int
	taskType       string
	listener       bridgeCore.Listener
	releaseTasksCh chan int
}

type withdrawReceipt struct {
	chainId    *big.Int
	withdrawId *big.Int
	recipient  common.Address
	token      common.Address
	amount     *big.Int
	fee        *big.Int
}

type depositReceipt struct {
	chainId   *big.Int
	depositId *big.Int
	recipient common.Address
	token     common.Address
	amount    *big.Int
}

func newBulkTask(listener bridgeCore.Listener, client *ethclient.Client, store stores.BridgeStore, chainId *big.Int, contracts map[string]string, ticker time.Duration, maxTry int, taskType string, releaseTasksCh chan int, util utils.Utils) *bulkTask {
	return &bulkTask{
		util:           util,
		tasks:          make([]*models.Task, 0),
		store:          store,
		client:         client,
		contracts:      contracts,
		chainId:        chainId,
		maxTry:         maxTry,
		taskType:       taskType,
		listener:       listener,
		releaseTasksCh: releaseTasksCh,
	}
}

func (r *bulkTask) collectTask(t *models.Task) {
	if t.Type == r.taskType {
		r.tasks = append(r.tasks, t)
	}
}

func (r *bulkTask) send() {
	log.Info("[bulkTask] sending bulk", "type", r.taskType, "tasks", len(r.tasks))
	if len(r.tasks) == 0 {
		return
	}
	switch r.taskType {
	case DEPOSIT_TASK:
		r.sendBulkTransactions(r.sendDepositTransaction)
	case WITHDRAWAL_TASK:
		r.sendBulkTransactions(r.sendWithdrawalSignaturesTransaction)
	case ACK_WITHDREW_TASK:
		r.sendBulkTransactions(r.sendAckTransactions)
	}
}

func (r *bulkTask) sendBulkTransactions(sendTxs func(tasks []*models.Task) (doneTasks, processingTasks, failedTasks []*models.Task, tx *ethtypes.Transaction)) {
	start, end := 0, len(r.tasks)
	for start < end {
		var (
			txHash string
			next   int
		)
		if start+defaultLimitRecords < end {
			next = start + defaultLimitRecords
		} else {
			next = end
		}
		log.Info("[bulkTask][sendBulkTransactions] start sending txs", "start", start, "end", end, "type", r.taskType)
		doneTasks, processingTasks, failedTasks, transaction := sendTxs(r.tasks[start:next])

		if transaction != nil {
			go updateTasks(r.store, processingTasks, STATUS_PROCESSING, transaction.Hash().Hex(), time.Now().Unix(), r.releaseTasksCh)
			metrics.Pusher.IncrGauge(metrics.ProcessingTaskMetric, len(processingTasks))
		}
		go updateTasks(r.store, doneTasks, STATUS_DONE, txHash, 0, r.releaseTasksCh)
		go updateTasks(r.store, failedTasks, STATUS_FAILED, txHash, 0, r.releaseTasksCh)
		metrics.Pusher.IncrCounter(metrics.SuccessTaskMetric, len(doneTasks))
		metrics.Pusher.IncrCounter(metrics.FailedTaskMetric, len(failedTasks))
		start = next
	}
}

func (r *bulkTask) sendDepositTransaction(tasks []*models.Task) (doneTasks, processingTasks, failedTasks []*models.Task, tx *ethtypes.Transaction) {
	var (
		receipts   []depositReceipt
		chainIds   []*big.Int
		depositIds []*big.Int
		recipients []common.Address
		tokens     []common.Address
		amounts    []*big.Int
	)
	// create caller
	caller, err := crossbellGateway.NewCrossbellGatewayCaller(common.HexToAddress(r.contracts[CROSSBELL_GATEWAY_CONTRACT]), r.client)
	if err != nil {
		for _, t := range tasks {
			t.LastError = err.Error()
			failedTasks = append(failedTasks, t)
		}
		return nil, nil, failedTasks, nil
	}

	// create transactor
	transactor, err := crossbellGateway.NewCrossbellGatewayTransactor(common.HexToAddress(r.contracts[CROSSBELL_GATEWAY_CONTRACT]), r.client)
	if err != nil {
		for _, t := range tasks {
			t.LastError = err.Error()
			failedTasks = append(failedTasks, t)
		}
		return nil, nil, failedTasks, nil
	}

	for _, t := range tasks {
		ok, receipt, err := r.validateDepositTask(caller, t)
		if err != nil {
			t.LastError = err.Error()
			failedTasks = append(failedTasks, t)
			continue
		}

		if receipt.depositId != nil {
			// store receiptId to processed receipt
			if err := r.store.GetProcessedReceiptStore().Save(t.ID, receipt.depositId.Int64()); err != nil {
				log.Error("[bulkTask][sendDepositTransaction] error while saving processed receipt", "err", err)
			}
		}

		// if deposit request is executed or voted (ok) then do nothing and add to doneTasks
		if ok {
			doneTasks = append(doneTasks, t)
			continue
		}

		// otherwise add task to processingTasks to adjust after sending transaction
		processingTasks = append(processingTasks, t)

		// append new receipt into receipts slice
		// TODO change chainId!!!!
		chainIds = append(chainIds, big.NewInt(1))
		depositIds = append(depositIds, receipt.depositId)
		recipients = append(recipients, receipt.recipient)
		tokens = append(tokens, receipt.token)
		amounts = append(amounts, receipt.amount)
		receipts = append(receipts, depositReceipt{
			chainId:   big.NewInt(1),
			depositId: receipt.depositId,
			recipient: receipt.recipient,
			token:     receipt.token,
			amount:    receipt.amount,
		})
	}
	metrics.Pusher.IncrCounter(metrics.DepositTaskMetric, len(tasks))

	if len(receipts) > 0 {
		tx, err = r.util.SendContractTransaction(r.listener.GetValidatorSign(), r.chainId, func(opts *bind.TransactOpts) (*ethtypes.Transaction, error) {
			return transactor.BatchAckDeposit(opts, chainIds, depositIds, recipients, tokens, amounts)
		})
		if err != nil {
			for _, t := range processingTasks {
				t.LastError = err.Error()
				if err.Error() == ErrNotBridgeOperator {
					doneTasks = append(doneTasks, t)
				} else {
					failedTasks = append(failedTasks, t)
				}
			}
			return doneTasks, nil, failedTasks, nil
		}
	}
	return
}

func (r *bulkTask) sendWithdrawalSignaturesTransaction(tasks []*models.Task) (doneTasks, processingTasks, failedTasks []*models.Task, tx *ethtypes.Transaction) {
	var (
		chainIds       []*big.Int
		withdrawIds    []*big.Int
		shouldReplaces []bool

		signatures [][]byte
	)
	//create transactor
	transactor, err := crossbellGateway.NewCrossbellGatewayTransactor(common.HexToAddress(r.contracts[CROSSBELL_GATEWAY_CONTRACT]), r.client)
	if err != nil {
		// append all success tasks into failed tasks
		for _, t := range tasks {
			t.LastError = err.Error()
			failedTasks = append(failedTasks, t)
		}
		return nil, nil, failedTasks, nil
	}
	// create caller
	caller, err := crossbellGateway.NewCrossbellGatewayCaller(common.HexToAddress(r.contracts[CROSSBELL_GATEWAY_CONTRACT]), r.client)
	if err != nil {
		// append all success tasks into failed tasks
		for _, t := range tasks {
			t.LastError = err.Error()
			failedTasks = append(failedTasks, t)
		}
		return nil, nil, failedTasks, nil
	}
	for _, t := range tasks {
		result, receipt, err := r.validateWithdrawalTask(caller, t)
		if err != nil {
			t.LastError = err.Error()
			failedTasks = append(failedTasks, t)
			continue
		}
		if receipt.withdrawId != nil {
			// store receiptId to processed receipt
			if err := r.store.GetProcessedReceiptStore().Save(t.ID, receipt.withdrawId.Int64()); err != nil {
				log.Error("[bulkTask][sendWithdrawalSignaturesTransaction] error while saving processed receipt", "err", err)
			}
		}
		// if validated then do nothing and add to doneTasks
		if result {
			doneTasks = append(doneTasks, t)
			continue
		}
		// otherwise add to processingTasks
		sigs, err := r.signWithdrawalSignatures(receipt)
		if err != nil {
			t.LastError = err.Error()
			failedTasks = append(failedTasks, t)
			continue
		}
		processingTasks = append(processingTasks, t)
		signatures = append(signatures, sigs)
		withdrawIds = append(withdrawIds, receipt.withdrawId)
		chainIds = append(chainIds, receipt.chainId)
		// TODO when to replace?
		shouldReplaces = append(shouldReplaces, true)
	}
	metrics.Pusher.IncrCounter(metrics.WithdrawalTaskMetric, len(tasks))

	if len(withdrawIds) > 0 {
		tx, err = r.util.SendContractTransaction(r.listener.GetValidatorSign(), r.chainId, func(opts *bind.TransactOpts) (*ethtypes.Transaction, error) {
			return transactor.BatchSubmitWithdrawalSignatures(opts, chainIds, withdrawIds, shouldReplaces, signatures)
		})
		if err != nil {
			// append all success tasks into failed tasks
			for _, t := range processingTasks {
				t.LastError = err.Error()
				if err.Error() == ErrNotBridgeOperator {
					doneTasks = append(doneTasks, t)
				} else {
					failedTasks = append(failedTasks, t)
				}
			}
			return doneTasks, nil, failedTasks, nil
		}
	}
	return
}

// func (r *bulkTask) sendAckTransactions(tasks []*models.Task) (doneTasks, processingTasks, failedTasks []*models.Task, tx *ethtypes.Transaction) {
// 	var (
// 		ids []*big.Int
// 	)
// 	// create transactor
// 	transactor, err := crossbellGateway.NewCrossbellGatewayTransactor(common.HexToAddress(r.contracts[CROSSBELL_GATEWAY_CONTRACT]), r.client)
// 	if err != nil {
// 		for _, t := range tasks {
// 			t.LastError = err.Error()
// 		}
// 		return nil, nil, tasks, nil
// 	}

// 	// create caller
// 	caller, err := crossbellGateway.NewCrossbellGatewayCaller(common.HexToAddress(r.contracts[CROSSBELL_GATEWAY_CONTRACT]), r.client)
// 	if err != nil {
// 		for _, t := range tasks {
// 			t.LastError = err.Error()
// 		}
// 		return nil, nil, tasks, nil
// 	}

// 	// loop through tasks, check if they are qualified to send ack transaction or not
// 	for _, t := range tasks {
// 		result, id, err := r.validateAckWithdrawalTask(caller, t)
// 		if err != nil {
// 			t.LastError = err.Error()
// 			failedTasks = append(failedTasks, t)
// 			continue
// 		}

// 		if id != nil {
// 			// store receiptId to processed receipt
// 			if err := r.store.GetProcessedReceiptStore().Save(t.ID, id.Int64()); err != nil {
// 				log.Error("[bulkTask][sendAckTransactions] error while saving processed receipt", "err", err)
// 			}
// 		}

// 		// if validated then do nothing and add to doneTasks
// 		if result {
// 			doneTasks = append(doneTasks, t)
// 			continue
// 		}

// 		// otherwise add id to ids and add task to processingTasks
// 		ids = append(ids, id)
// 		processingTasks = append(processingTasks, t)
// 	}

// 	metrics.Pusher.IncrCounter(metrics.AckWithdrawalTaskMetric, len(tasks))
// 	// TODO: ack withdraw
// 	if len(ids) > 0 {
// 		tx, err = r.util.SendContractTransaction(r.listener.GetValidatorSign(), r.chainId, func(opts *bind.TransactOpts) (*ethtypes.Transaction, error) {
// 			return transactor.BatchAckDeposit(opts, ids, )
// 		})
// 		if err != nil {
// 			// append all success tasks into failed tasks
// 			for _, t := range processingTasks {
// 				t.LastError = err.Error()
// 				if err.Error() == ErrNotBridgeOperator {
// 					doneTasks = append(doneTasks, t)
// 				} else {
// 					failedTasks = append(failedTasks, t)
// 				}
// 			}
// 			return doneTasks, nil, failedTasks, nil
// 		}
// 	}
// 	return
// }

// ValidateDepositTask validates if:
// - current signer has been voted for a deposit request or not
// - deposit request has been executed or not
// also returns transfer receipt
func (r *bulkTask) validateDepositTask(caller *crossbellGateway.CrossbellGatewayCaller, task *models.Task) (bool, depositReceipt, error) {
	mainchainEvent := new(mainchainGateway.MainchainGatewayRequestDeposit)
	mainchainGatewayAbi, err := mainchainGateway.MainchainGatewayMetaData.GetAbi()
	if err != nil {
		return false, depositReceipt{}, err
	}

	data := common.Hex2Bytes(task.Data)
	if err = r.util.UnpackLog(*mainchainGatewayAbi, mainchainEvent, "RequestDeposit", data); err != nil {
		return false, depositReceipt{}, err
	}

	// check if current validator has been voted for this deposit or not
	// TODO change big.NewInt(1) into chainId from event!!!!
	acknowledgementHash, err := caller.GetValidatorAcknowledgementHash(nil, big.NewInt(1), mainchainEvent.DepositId, r.listener.GetValidatorSign().GetAddress())
	voted := int(big.NewInt(0).SetBytes(acknowledgementHash[:]).Uint64()) == 0
	if err != nil {
		return false, depositReceipt{}, err
	}
	return voted, depositReceipt{big.NewInt(1), mainchainEvent.DepositId, mainchainEvent.Recipient, mainchainEvent.Token, mainchainEvent.Amount}, nil
}

// ValidateAckWithdrawalTask validates if:
// - signer has been voted for a withdrawal or not
// - withdrawal request is executed or not
// returns true if withdraw is executed or voted
// also returns receipt id
func (r *bulkTask) validateAckWithdrawalTask(caller *crossbellGateway.CrossbellGatewayCaller, task *models.Task) (bool, *big.Int, error) {
	// Unpack event data
	mainchainEvent := new(mainchainGateway.MainchainGatewayWithdrew)
	mainchainGatewayAbi, err := mainchainGateway.MainchainGatewayMetaData.GetAbi()
	if err != nil {
		return false, nil, err
	}

	if err = r.util.UnpackLog(*mainchainGatewayAbi, mainchainEvent, "Withdrew", common.Hex2Bytes(task.Data)); err != nil {
		return false, nil, err
	}

	// check if withdrew has been voted or not
	// check if withdrew has been voted or not
	// TODO currently there's no chainId in a withdrew event so i just input a random one
	// replace this chainId
	acknowledgementHash, err := caller.GetValidatorAcknowledgementHash(nil, big.NewInt(5), mainchainEvent.WithdrawalId, r.listener.GetValidatorSign().GetAddress())
	if err != nil {
		return false, nil, err
	}
	voted := int(big.NewInt(0).SetBytes(acknowledgementHash[:]).Uint64()) == 0
	return voted, mainchainEvent.WithdrawalId, nil
}

// ValidateWithdrawalTask validates if:
// - Withdrawal request is executed or not
// returns true if it is executed
// also returns transfer receipt
func (r *bulkTask) validateWithdrawalTask(caller *crossbellGateway.CrossbellGatewayCaller, task *models.Task) (bool, *withdrawReceipt, error) {
	// Unpack event from data
	crossbellEvent := new(crossbellGateway.CrossbellGatewayRequestWithdrawal)
	ronGatewayAbi, err := roninGateway.GatewayMetaData.GetAbi()
	if err != nil {
		return false, &withdrawReceipt{}, err
	}
	if err = r.util.UnpackLog(*ronGatewayAbi, crossbellEvent, "WithdrawalRequested", common.Hex2Bytes(task.Data)); err != nil {
		return false, &withdrawReceipt{crossbellEvent.ChainId, crossbellEvent.WithdrawId, crossbellEvent.Recipient, crossbellEvent.Token, crossbellEvent.Amount, crossbellEvent.Fee}, err
	}
	return false, &withdrawReceipt{}, nil
}

func updateTasks(store stores.BridgeStore, tasks []*models.Task, status, txHash string, timestamp int64, releaseTasksCh chan int) {
	// update tasks with given status
	// note: if task.retries < 10 then retries++ and status still be processing
	for _, t := range tasks {
		if timestamp > 0 {
			t.TxCreatedAt = timestamp
		}
		if status == STATUS_FAILED {
			if t.Retries+1 >= 10 {
				t.Status = status
			} else {
				t.Retries += 1
			}
		} else {
			t.Status = status
			t.TransactionHash = txHash
		}
		if err := store.GetTaskStore().Update(t); err != nil {
			log.Error("error while update task", "id", t.ID, "err", err)
		}
		releaseTasksCh <- t.ID
	}
}

func parseSignatureAsRsv(signature []byte) roninGovernance.SignatureConsumerSignature {
	rawR := signature[0:32]
	rawS := signature[32:64]
	v := signature[64]

	if v < 27 {
		v += 27
	}

	var r, s [32]byte
	copy(r[:], rawR)
	copy(s[:], rawS)

	return roninGovernance.SignatureConsumerSignature{
		R: r,
		S: s,
		V: v,
	}
}

func (r *bulkTask) signWithdrawalSignatures(receipt *withdrawReceipt) (hexutil.Bytes, error) {
	typedData := core.TypedData{
		Types: core.Types{
			"EIP712Domain": []core.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"Receipt": []core.Type{
				{Name: "id", Type: "uint256"},
				{Name: "kind", Type: "uint8"},
				{Name: "mainchain", Type: "TokenOwner"},
				{Name: "ronin", Type: "TokenOwner"},
				{Name: "info", Type: "TokenInfo"},
			},
			"TokenOwner": []core.Type{
				{Name: "addr", Type: "address"},
				{Name: "tokenAddr", Type: "address"},
				{Name: "chainId", Type: "uint256"},
			},
			"TokenInfo": []core.Type{
				{Name: "erc", Type: "uint8"},
				{Name: "id", Type: "uint256"},
				{Name: "quantity", Type: "uint256"},
			},
		},
		Domain: core.TypedDataDomain{
			Name:              "MainchainGatewayV2",
			Version:           "2",
			ChainId:           math.NewHexOrDecimal256(receipt.chainId.Int64()),
			VerifyingContract: r.contracts[ETH_GATEWAY_CONTRACT],
		},
		PrimaryType: "Receipt",
		Message: core.TypedDataMessage{
			"id":        receipt.withdrawId.String(),
			"recipient": receipt.recipient,
			"token":     receipt.token,
			"amount":    receipt.amount,
			"fee":       receipt.fee,
			// todo remove unused infos here
			// "kind": fmt.Sprintf("%d", receipt.Kind),
			// "mainchain": core.TypedDataMessage{
			// 	"addr":      receipt.Mainchain.Addr.Hex(),
			// 	"tokenAddr": receipt.Mainchain.TokenAddr.Hex(),
			// 	"chainId":   receipt.Mainchain.ChainId.String(),
			// },
			// "ronin": core.TypedDataMessage{
			// 	"addr":      receipt.Ronin.Addr.Hex(),
			// 	"tokenAddr": receipt.Ronin.TokenAddr.Hex(),
			// 	"chainId":   receipt.Ronin.ChainId.String(),
			// },
			// "info": core.TypedDataMessage{
			// 	"erc":      fmt.Sprintf("%d", receipt.Info.Erc),
			// 	"id":       receipt.Info.Id.String(),
			// 	"quantity": receipt.Info.Quantity.String(),
			// },
		},
	}
	return r.util.SignTypedData(typedData, r.listener.GetValidatorSign())
}
