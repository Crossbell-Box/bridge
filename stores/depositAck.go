package stores

import (
	"github.com/axieinfinity/bridge-v2/models"
	"gorm.io/gorm"
)

type depositAckStore struct {
	*gorm.DB
}

func NewDepositAckStore(db *gorm.DB) *depositAckStore {
	return &depositAckStore{db}
}

func (d *depositAckStore) Save(depositAck *models.DepositAck) error {
	return d.Create(depositAck).Error
}
