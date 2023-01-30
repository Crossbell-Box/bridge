package stores

import (
	"github.com/axieinfinity/bridge-v2/models"
	"gorm.io/gorm"
)

type requestDepositStore struct {
	*gorm.DB
}

func NewRequestDepositStore(db *gorm.DB) *requestDepositStore {
	return &requestDepositStore{db}
}

func (d *requestDepositStore) Save(requestDeposit *models.RequestDeposit) error {
	return d.Create(requestDeposit).Error
}

func (d *requestDepositStore) Update(requestDeposit *models.RequestDeposit) error {
	columns := map[string]interface{}{
		"status": requestDeposit.Status,
	}
	return d.Model(&models.RequestDeposit{}).Where("mainchain_id = ? AND deposit_id = ?", requestDeposit.MainchainId, requestDeposit.DepositId).Updates(columns).Error
}
