package stores

import (
	"github.com/axieinfinity/bridge-v2/models"
	"gorm.io/gorm"
)

type withdrawalSignaturesStore struct {
	*gorm.DB
}

func NewWithdrawalSignaturesStore(db *gorm.DB) *withdrawalSignaturesStore {
	return &withdrawalSignaturesStore{db}
}

func (w *withdrawalSignaturesStore) Save(withdrawalSignatures *models.WithdrawalSignatures) error {
	return w.Create(withdrawalSignatures).Error
}

// func (w *withdrawalSignaturesStore) Save(withdrawSignatures *models.WithdrawalSignatures) error {
// 	return w.Clauses(clause.OnConflict{DoNothing: true}).Create(withdrawSignatures).Error
// }

// func (w *withdrawalSignaturesStore) Update(withdrawSignatures *models.WithdrawalSignatures) error {
// 	return w.Updates(withdrawalSignaturesStore).Error
// }

// func (w *withdrawalSignaturesStore) GetWithdrawalById(withdrawalId int64) (*models.Withdrawal, error) {
// 	var withdraw *models.Withdrawal
// 	err := w.Model(&models.Withdrawal{}).Where("withdrawal_id = ?", withdrawalId).First(&withdraw).Error
// 	return withdraw, err
// }
