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
