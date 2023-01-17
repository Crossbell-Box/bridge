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
