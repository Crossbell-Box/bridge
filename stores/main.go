package stores

import (
	"github.com/axieinfinity/bridge-v2/models"
	"gorm.io/gorm"
)

type TaskStore interface {
	Save(task *models.Task) error
	Update(task *models.Task) error
	GetTasks(chain, status string, limit, retrySeconds int, before int64, excludeIds []int) ([]*models.Task, error)
	UpdateTasksWithTransactionHash(txs []string, transactionStatus int, status string) error
	DeleteTasks([]string, uint64) error
	Count() int64
	ResetTo(ids []string, status string) error
}

type DepositStore interface {
	Save(deposit *models.Deposit) error
}
type RequestDepositStore interface {
	Save(requestDeposit *models.RequestDeposit) error
}

type RequestWithdrawalStore interface {
	Save(RequestWithdrawal *models.RequestWithdrawal) error
}

type ProcessedReceiptStore interface {
	Save(taskId int, receiptId int64) error
}

type WithdrawalStore interface {
	Save(withdraw *models.Withdrawal) error
	Update(withdraw *models.Withdrawal) error
	GetWithdrawalById(withdrawalId int64) (*models.Withdrawal, error)
}

type WithdrawalSignaturesStore interface {
	Save(WithdrawalSignatures *models.WithdrawalSignatures) error
}

type DepositAckStore interface {
	Save(DepositAck *models.DepositAck) error
}

type BridgeStore interface {
	GetDepositStore() DepositStore
	GetWithdrawalStore() WithdrawalStore
	GetTaskStore() TaskStore
	GetProcessedReceiptStore() ProcessedReceiptStore
	GetWithdrawalSignaturesStore() WithdrawalSignaturesStore
	GetDepositAckStore() DepositAckStore
	GetRequestDepositStore() RequestDepositStore
	GetRequestWithdrawalStore() RequestWithdrawalStore
}

type bridgeStore struct {
	*gorm.DB

	DepositStore              DepositStore
	WithdrawalStore           WithdrawalStore
	TaskStore                 TaskStore
	ProcessedReceiptStore     ProcessedReceiptStore
	WithdrawalSignaturesStore WithdrawalSignaturesStore
	DepositAckStore           DepositAckStore
	RequestDepositStore       RequestDepositStore
	RequestWithdrawalStore    RequestWithdrawalStore
}

func NewBridgeStore(db *gorm.DB) BridgeStore {
	store := &bridgeStore{
		DB: db,

		TaskStore:                 NewTaskStore(db),
		DepositStore:              NewDepositStore(db),
		WithdrawalStore:           NewWithdrawalStore(db),
		ProcessedReceiptStore:     NewProcessedReceiptStore(db),
		WithdrawalSignaturesStore: NewWithdrawalSignaturesStore(db),
		DepositAckStore:           NewDepositAckStore(db),
		RequestDepositStore:       NewRequestDepositStore(db),
		RequestWithdrawalStore:    NewRequestWithdrawalStore(db),
	}
	return store
}

func (m *bridgeStore) RelationalDatabaseCheck() error {
	return m.Raw("SELECT 1").Error
}

func (m *bridgeStore) GetDB() *gorm.DB {
	return m.DB
}

func (m *bridgeStore) GetDepositStore() DepositStore {
	return m.DepositStore
}

func (m *bridgeStore) GetWithdrawalStore() WithdrawalStore {
	return m.WithdrawalStore
}

func (m *bridgeStore) GetRequestDepositStore() RequestDepositStore {
	return m.RequestDepositStore
}

func (m *bridgeStore) GetRequestWithdrawalStore() RequestWithdrawalStore {
	return m.RequestWithdrawalStore
}

func (m *bridgeStore) GetTaskStore() TaskStore {
	return m.TaskStore
}

func (m *bridgeStore) GetProcessedReceiptStore() ProcessedReceiptStore {
	return m.ProcessedReceiptStore
}

func (m *bridgeStore) GetWithdrawalSignaturesStore() WithdrawalSignaturesStore {
	return m.WithdrawalSignaturesStore
}

func (m *bridgeStore) GetDepositAckStore() DepositAckStore {
	return m.DepositAckStore
}
