package listener

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ashwanthkumar/slack-go-webhook"
	bridgeCore "github.com/axieinfinity/bridge-core"
	"github.com/axieinfinity/bridge-core/metrics"
	bridgeCoreModels "github.com/axieinfinity/bridge-core/models"
	"github.com/axieinfinity/bridge-core/stores"
	"github.com/axieinfinity/bridge-core/utils"
	bridgeCoreUtils "github.com/axieinfinity/bridge-core/utils"
	mainchainGateway "github.com/axieinfinity/bridge-v2/bridge-contracts/generated_contracts/mainchainGateway"
	"github.com/axieinfinity/bridge-v2/task"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

type EthereumListener struct {
	ctx       context.Context
	cancelCtx context.CancelFunc

	config *bridgeCore.LsConfig

	chainId *big.Int
	jobId   int32

	rpcUrl               string
	slackUrl             string
	scanUrl              string
	name                 string
	period               time.Duration
	currentBlock         atomic.Value
	safeBlockRange       uint64
	preventOmissionRange uint64
	fromHeight           uint64
	domainSeparators     map[uint64]string
	decimals             map[uint64]uint64
	batches              sync.Map
	utilsWrapper         utils.Utils
	client               utils.EthClient
	validatorSign        bridgeCoreUtils.ISign
	store                stores.MainStore
	listeners            map[string]bridgeCore.Listener

	prepareJobChan chan bridgeCore.JobHandler
	tasks          []bridgeCore.TaskHandler
}

func (e *EthereumListener) AddListeners(m map[string]bridgeCore.Listener) {
	e.listeners = m
}

func (e *EthereumListener) GetListener(s string) bridgeCore.Listener {
	return e.listeners[s]
}

func NewEthereumListener(ctx context.Context, cfg *bridgeCore.LsConfig, helpers utils.Utils, store stores.MainStore) (*EthereumListener, error) {
	newCtx, cancelFunc := context.WithCancel(ctx)
	ethListener := &EthereumListener{
		name:                 cfg.Name,
		period:               cfg.LoadInterval,
		currentBlock:         atomic.Value{},
		ctx:                  newCtx,
		cancelCtx:            cancelFunc,
		fromHeight:           cfg.FromHeight,
		domainSeparators:     cfg.DomainSeparators,
		decimals:             cfg.Decimals,
		utilsWrapper:         utils.NewUtils(),
		store:                store,
		config:               cfg,
		listeners:            make(map[string]bridgeCore.Listener),
		chainId:              hexutil.MustDecodeBig(cfg.ChainId),
		safeBlockRange:       cfg.SafeBlockRange,
		preventOmissionRange: cfg.PreventOmissionRange,
		tasks:                make([]bridgeCore.TaskHandler, 0),
	}
	if helpers != nil {
		ethListener.utilsWrapper = helpers
	}
	client, err := ethListener.utilsWrapper.NewEthClient(cfg.RpcUrl)
	if err != nil {
		log.Error(fmt.Sprintf("[New%sListener] error while dialing rpc client", cfg.Name), "err", err, "url", cfg.RpcUrl, "slackUrl", cfg.SlackUrl)
		return nil, err
	}
	ethListener.client = client

	if cfg.Secret.Validator != nil {
		ethListener.validatorSign, err = bridgeCoreUtils.NewSignMethod(cfg.Secret.Validator)
		if err != nil {
			log.Error(fmt.Sprintf("[New%sListener] error while getting validator key", cfg.Name), "err", err)
			return nil, err
		}
	}
	return ethListener, nil
}

func (e *EthereumListener) GetStore() stores.MainStore {
	return e.store
}

func (e *EthereumListener) Config() *bridgeCore.LsConfig {
	return e.config
}

func (e *EthereumListener) Start() {
	for _, task := range e.tasks {
		go task.Start()
	}
}

func (e *EthereumListener) GetName() string {
	return e.name
}

func (e *EthereumListener) Period() time.Duration {
	return e.period
}

func (e *EthereumListener) GetSafeBlockRange() uint64 {
	return e.safeBlockRange
}

func (e *EthereumListener) GetPreventOmissionRange() uint64 {
	return e.preventOmissionRange
}

func (e *EthereumListener) IsDisabled() bool {
	return e.config.Disabled
}

func (e *EthereumListener) SetInitHeight(height uint64) {
	e.fromHeight = height
}

func (e *EthereumListener) GetInitHeight() uint64 {
	return e.fromHeight
}

func (e *EthereumListener) GetDomainSeparators() map[uint64]string {
	return e.domainSeparators
}

func (e *EthereumListener) GetDecimals() map[uint64]uint64 {
	return e.decimals
}

func (e *EthereumListener) GetTask(index int) bridgeCore.TaskHandler {
	return e.tasks[index]
}

func (e *EthereumListener) GetTasks() []bridgeCore.TaskHandler {
	return e.tasks
}

func (e *EthereumListener) AddTask(task bridgeCore.TaskHandler) {
	e.tasks = append(e.tasks, task)
}

func (e *EthereumListener) GetCurrentBlock() bridgeCore.Block {
	if _, ok := e.currentBlock.Load().(bridgeCore.Block); !ok {
		var (
			block bridgeCore.Block
			err   error
		)
		block, err = e.GetProcessedBlock()
		if err != nil {
			log.Error(fmt.Sprintf("[%sListener] error on getting processed block from database", e.GetName()), "err", err)
			if e.fromHeight > 0 {
				block, err = e.GetBlock(e.fromHeight)
				if err != nil {
					log.Error(fmt.Sprintf("[%sListener] error on getting block from rpc", e.GetName()), "err", err, "fromHeight", e.fromHeight)
				}
			}
		}
		// if block is still nil, get latest block from rpc
		if block == nil {
			block, err = e.GetLatestBlock()
			if err != nil {
				log.Error(fmt.Sprintf("[%sListener] error on getting latest block from rpc", e.GetName()), "err", err)
				return nil
			}
		}
		e.currentBlock.Store(block)
		return block
	}
	return e.currentBlock.Load().(bridgeCore.Block)
}

func (e *EthereumListener) IsUpTodate() bool {
	return true
}

func (e *EthereumListener) GetProcessedBlock() (bridgeCore.Block, error) {
	chainId, err := e.GetChainID()
	if err != nil {
		log.Error(fmt.Sprintf("[%sListener][GetLatestBlock] error while getting chainId", e.GetName()), "err", err)
		return nil, err
	}
	encodedChainId := hexutil.EncodeBig(chainId)
	height, err := e.store.GetProcessedBlockStore().GetLatestBlock(encodedChainId)
	if err != nil {
		log.Error(fmt.Sprintf("[%sListener][GetLatestBlock] error while getting latest height from database", e.GetName()), "err", err, "chainId", encodedChainId)
		return nil, err
	}
	block, err := e.client.BlockByNumber(e.ctx, big.NewInt(height))
	if err != nil {
		return nil, err
	}
	return NewEthBlock(e.client, chainId, block, false)
}

func (e *EthereumListener) GetLatestBlock() (bridgeCore.Block, error) {
	block, err := e.client.BlockByNumber(e.ctx, nil)
	if err != nil {
		return nil, err
	}
	return NewEthBlock(e.client, e.chainId, block, false)
}

func (e *EthereumListener) GetLatestBlockHeight() (uint64, error) {
	return e.client.BlockNumber(e.ctx)
}

func (e *EthereumListener) GetChainID() (*big.Int, error) {
	if e.chainId != nil {
		return e.chainId, nil
	}
	return e.client.ChainID(e.ctx)
}

func (e *EthereumListener) Context() context.Context {
	return e.ctx
}

func (e *EthereumListener) GetSubscriptions() map[string]*bridgeCore.Subscribe {
	return e.config.Subscriptions
}

func (e *EthereumListener) UpdateCurrentBlock(block bridgeCore.Block) error {
	if block != nil && e.GetCurrentBlock().GetHeight() < block.GetHeight() {
		log.Info(fmt.Sprintf("[%sListener] UpdateCurrentBlock", e.name), "new block", block.GetHeight(), "current block", e.GetCurrentBlock().GetHeight())
		e.currentBlock.Store(block)
		return e.SaveCurrentBlockToDB()
	} else if e.GetCurrentBlock().GetHeight() >= block.GetHeight() {
		log.Info(fmt.Sprintf("[%sListener] UpdateCurrentBlock: block <= current block. Check the RPC if needed", e.name), "new block", block.GetHeight(), "current block", e.GetCurrentBlock().GetHeight())
	}
	return nil
}

func (e *EthereumListener) SaveCurrentBlockToDB() error {
	chainId, err := e.GetChainID()
	if err != nil {
		return err
	}

	if err := e.store.GetProcessedBlockStore().Save(hexutil.EncodeBig(chainId), int64(e.GetCurrentBlock().GetHeight())); err != nil {
		return err
	}

	metrics.Pusher.IncrCounter(fmt.Sprintf(metrics.ListenerProcessedBlockMetric, e.GetName()), 1)
	return nil
}

func (e *EthereumListener) SaveTransactionsToDB(txs []bridgeCore.Transaction) error {
	return nil
}

func (e *EthereumListener) SetPrepareJobChan(jobChan chan bridgeCore.JobHandler) {
	e.prepareJobChan = jobChan
}

func (e *EthereumListener) GetEthClient() utils.EthClient {
	return e.client
}

func (e *EthereumListener) GetListenHandleJob(subscriptionName string, tx bridgeCore.Transaction, eventId string, data []byte) bridgeCore.JobHandler {
	// validate if data contains subscribed name
	subscription, ok := e.GetSubscriptions()[subscriptionName]
	if !ok {
		return nil
	}
	handlerName := subscription.Handler.Name
	if subscription.Type == bridgeCore.TxEvent {
		method, ok := subscription.Handler.ABI.Methods[handlerName]
		if !ok {
			return nil
		}
		if method.RawName != common.Bytes2Hex(data[0:4]) {
			return nil
		}
	} else if subscription.Type == bridgeCore.LogEvent {
		event, ok := subscription.Handler.ABI.Events[handlerName]
		if !ok {
			return nil
		}
		if hexutil.Encode(event.ID.Bytes()) != eventId {
			return nil
		}
	} else {
		return nil
	}
	return NewEthListenJob(bridgeCore.ListenHandler, e, subscriptionName, tx, data)
}

func (e *EthereumListener) SendCallbackJobs(listeners map[string]bridgeCore.Listener, subscriptionName string, tx bridgeCore.Transaction, inputData []byte) {
	log.Info(fmt.Sprintf("[%sListener][SendCallbackJobs] Start", e.GetName()), "subscriptionName", subscriptionName, "listeners", len(listeners), "fromTx", tx.GetHash().Hex())
	chainId, err := e.GetChainID()
	if err != nil {
		return
	}
	subscription, ok := e.GetSubscriptions()[subscriptionName]
	if !ok {
		log.Warn(fmt.Sprintf("[%sListener][SendCallbackJobs] cannot find subscription", e.GetName()), "subscriptionName", subscriptionName)
		return
	}
	log.Info(fmt.Sprintf("[%sListener][SendCallbackJobs] subscription found", e.GetName()), "subscriptionName", subscriptionName, "numberOfCallbacks", len(subscription.CallBacks))
	for listenerName, methodName := range subscription.CallBacks {
		log.Info(fmt.Sprintf("[%sListener][SendCallbackJobs] Loop through callbacks", e.GetName()), "subscriptionName", subscriptionName, "listenerName", listenerName, "methodName", methodName)
		l := listeners[listenerName]
		job := NewEthCallbackJob(l, methodName, tx, inputData, chainId, e.utilsWrapper)
		if job != nil {
			e.prepareJobChan <- job
		}
	}
}

func (e *EthereumListener) GetBlock(height uint64) (bridgeCore.Block, error) {
	block, err := e.client.BlockByNumber(e.ctx, big.NewInt(int64(height)))
	if err != nil {
		return nil, err
	}
	return NewEthBlock(e.client, e.chainId, block, false)
}

func (e *EthereumListener) GetBlockWithLogs(height uint64) (bridgeCore.Block, error) {
	block, err := e.client.BlockByNumber(e.ctx, big.NewInt(int64(height)))
	if err != nil {
		return nil, err
	}
	return NewEthBlock(e.client, e.chainId, block, true)
}

func (e *EthereumListener) GetReceipt(txHash common.Hash) (*ethtypes.Receipt, error) {
	return e.client.TransactionReceipt(e.ctx, txHash)
}

func (e *EthereumListener) NewJobFromDB(job *bridgeCoreModels.Job) (bridgeCore.JobHandler, error) {
	return newJobFromDB(e, job)
}

func newJobFromDB(listener bridgeCore.Listener, job *bridgeCoreModels.Job) (bridgeCore.JobHandler, error) {
	chainId, err := hexutil.DecodeBig(job.FromChainId)
	if err != nil {
		return nil, err
	}
	// get transaction from hash
	tx, _, err := listener.GetEthClient().TransactionByHash(context.Background(), common.HexToHash(job.Transaction))
	if err != nil {
		return nil, err
	}
	transaction, err := NewEthTransaction(chainId, tx)
	if err != nil {
		return nil, err
	}
	baseJob, err := bridgeCore.NewBaseJob(listener, job, transaction)
	if err != nil {
		return nil, err
	}
	switch job.Type {
	case bridgeCore.ListenHandler:
		return &EthListenJob{
			BaseJob: baseJob,
		}, nil
	case bridgeCore.CallbackHandler:
		if job.Method == "" {
			return nil, nil
		}
		return &EthCallbackJob{
			BaseJob: baseJob,
			method:  job.Method,
		}, nil
	}
	return nil, errors.New("jobType does not match")
}

func (e *EthereumListener) Close() {
	e.client.Close()
	e.cancelCtx()
}

func (e *EthereumListener) GetValidatorSign() bridgeCoreUtils.ISign {
	return e.validatorSign
}

// WithdrewDone2SlackCallback send the withdrew event rto slack channel through slack hook
func (l *EthereumListener) WithdrewDone2SlackCallback(fromChainId *big.Int, tx bridgeCore.Transaction, data []byte) error {
	log.Info("[EthereumListener] WithdrewDone2SlackCallback", "tx", tx.GetHash().Hex())
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
		log.Info("[EthereumListener] Sending to slack", "slack hook url", webhookUrl)
		// create caller
		caller, err := mainchainGateway.NewMainchainGatewayCaller(common.HexToAddress(l.config.Contracts[task.MAINCHAIN_GATEWAY_CONTRACT]), l.client)
		if err != nil {
			return err
		}
		decimal := l.config.Decimals[mainchainEvent.ChainId.Uint64()]

		// check remaining quota
		remainingQuota, err := caller.GetDailyWithdrawalRemainingQuota(nil, mainchainEvent.Token)
		remainingQuotaDecimal := l.utilsWrapper.ToDecimal(remainingQuota, decimal)
		if err != nil {
			log.Error("[Slack hook] error while querying remainingQuota ", "error", err)
			return err
		}

		attachment1 := slack.Attachment{}
		attachment1.AddAction(slack.Action{Type: "divider"})
		attachment1.AddField(slack.Field{Title: "Event", Value: ":golf:Withdrew"})
		// query for character
		response, err := fetchCharacters(mainchainEvent.Recipient.String())
		if err != nil {
			log.Error("[Query Primary Character] error while querying primary character ", "error", err)
		} else {
			attachment1.AddField(slack.Field{Title: "Event", Value: response})
		}
		attachment1.AddField(slack.Field{Title: "Mainchain ID", Value: mainchainEvent.ChainId.String()})
		attachment1.AddField(slack.Field{Title: "Withdraw ID", Value: mainchainEvent.WithdrawalId.String()})
		attachment1.AddField(slack.Field{Title: "Amount", Value: fmt.Sprintf("%s $MIRA", l.utilsWrapper.ToDecimal(mainchainEvent.Amount, decimal))})
		attachment1.AddField(slack.Field{Title: "Fee", Value: fmt.Sprintf("%s $MIRA", l.utilsWrapper.ToDecimal(mainchainEvent.Fee, decimal))})
		attachment1.AddField(slack.Field{Title: "Remainning Quota", Value: fmt.Sprintf("%s $MIRA", remainingQuotaDecimal)})
		attachment1.AddAction(slack.Action{Type: "button", Text: "View Details", Url: fmt.Sprintf("%s%s", l.config.ScanUrl, tx.GetHash().Hex()), Style: "primary"})

		payload := slack.Payload{
			Text:        fmt.Sprintf(":golf:*Successfully <%s%s|*Withdrew*> on %s!*:golf:\n", l.config.ScanUrl, tx.GetHash().Hex(), l.config.Name),
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

func fetchCharacters(address string) (string, error) {
	const (
		baseURL      = "https://indexer.crossbell.io/v1/addresses/"
		queryParams  = "/characters?limit=20&primary=true"
		acceptHeader = "application/json"
	)
	client := &http.Client{}
	url := baseURL + address + queryParams

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", acceptHeader)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", err
	}

	handle := ""
	list, ok := data["list"].([]interface{})
	if ok && len(list) > 0 {
		firstItem, ok := list[0].(map[string]interface{})
		if ok {
			handle, _ = firstItem["handle"].(string)
		}
	}
	return handle, nil
}
