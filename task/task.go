package task

import (
	"bytes"
	"crypto/ecdsa"
	"math/big"
	"time"

	bridgeCore "github.com/axieinfinity/bridge-core"
	"github.com/axieinfinity/bridge-core/metrics"
	"github.com/axieinfinity/bridge-core/utils"
	"github.com/axieinfinity/bridge-v2/models"
	"github.com/axieinfinity/bridge-v2/stores"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

const (
	Salt                   = "0xe3922a0bff7e80c6f7465bc1b150f6c95d9b9203f1731a09f86e759ea1eaa306"
	ErrSigAlreadySubmitted = "execution reverted: BOsGovernanceRelay: query for outdated period"
	ErrOutdatedPeriod      = "execution reverted: BOsGovernanceProposal: query for outdated period"
)

type task struct {
	util           utils.Utils
	task           *models.Task
	store          stores.BridgeStore
	validator      *ecdsa.PrivateKey
	client         bind.ContractBackend
	ethClient      bind.ContractBackend
	contracts      map[string]string
	chainId        *big.Int
	maxTry         int
	taskType       string
	listener       bridgeCore.Listener
	releaseTasksCh chan int
}

func newTask(listener bridgeCore.Listener, client, ethClient bind.ContractBackend, store stores.BridgeStore, chainId *big.Int, contracts map[string]string, maxTry int, taskType string, releaseTasksCh chan int, util utils.Utils) *task {
	return &task{
		util:           util,
		task:           nil,
		store:          store,
		client:         client,
		ethClient:      ethClient,
		contracts:      contracts,
		chainId:        chainId,
		maxTry:         maxTry,
		taskType:       taskType,
		listener:       listener,
		releaseTasksCh: releaseTasksCh,
	}
}

func (r *task) collectTask(t *models.Task) {
	log.Debug("Received new task", "id", t.ID, "status", t.Status, "type", t.Type)
	if t.Type == r.taskType {
		r.task = t
	}
}

func (r *task) sendTransaction(sendTx func(task *models.Task) (doneTasks, processingTasks, failedTasks []*models.Task, tx *ethtypes.Transaction)) {
	if r.task == nil {
		return
	}

	var txHash string

	doneTasks, processingTasks, failedTasks, transaction := sendTx(r.task)

	if transaction != nil {
		log.Debug("[task] Transaction", "type", r.taskType, "hash", transaction.Hash().Hex())
		go updateTasks(r.store, processingTasks, STATUS_PROCESSING, transaction.Hash().Hex(), time.Now().Unix(), r.releaseTasksCh)
		_ = metrics.Pusher.IncrGauge(metrics.ProcessingTaskMetric, len(processingTasks))
	}
	go updateTasks(r.store, doneTasks, STATUS_DONE, txHash, 0, r.releaseTasksCh)
	go updateTasks(r.store, failedTasks, STATUS_FAILED, txHash, 0, r.releaseTasksCh)
	_ = metrics.Pusher.IncrCounter(metrics.SuccessTaskMetric, len(doneTasks))
	_ = metrics.Pusher.IncrCounter(metrics.FailedTaskMetric, len(failedTasks))
}

type signDataOpts struct {
	SignTypedDataCallback func(typedData apitypes.TypedData) (hexutil.Bytes, error)
}

// validatorsAscending implements the sort interface to allow sorting a list of addresses
type validatorsAscending []common.Address

func (s validatorsAscending) Len() int           { return len(s) }
func (s validatorsAscending) Less(i, j int) bool { return bytes.Compare(s[i][:], s[j][:]) < 0 }
func (s validatorsAscending) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
