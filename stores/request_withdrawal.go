package stores

import (
	"github.com/axieinfinity/bridge-v2/models"
	"gorm.io/gorm"
)

type requestWithdrawalStore struct {
	*gorm.DB
}

func NewRequestWithdrawalStore(db *gorm.DB) *requestWithdrawalStore {
	return &requestWithdrawalStore{db}
}

func (d *requestWithdrawalStore) Save(requestWithdrawal *models.RequestWithdrawal) error {
	return d.Create(requestWithdrawal).Error
}

func (d *requestWithdrawalStore) Update(requestWithdrawal *models.RequestWithdrawal) error {
	columns := map[string]interface{}{
		"status":                 requestWithdrawal.Status,
		"withdrawal_transaction": requestWithdrawal.WithdrawalTransaction,
	}
	return d.Model(&models.RequestWithdrawal{}).Where("mainchain_id = ? AND withdrawal_id = ?", requestWithdrawal.MainchainId, requestWithdrawal.WithdrawalId).Updates(columns).Error
}
