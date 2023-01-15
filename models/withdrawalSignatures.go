package models

import (
	"gorm.io/gorm"
)

type WithdrawalSignatures struct {
	ID               int    `json:"id" gorm:"primary_key:true;column:id;auto_increment;not null"`
	MainchainId      int64  `json:"mainchainId" gorm:"column:mainchain_id;uniqueIndex:idx_withdrawalSignatures_withdrawal;priority:1;not null"`
	WithdrawalId     int64  `json:"withdrawalId" gorm:"column:withdrawal_id;uniqueIndex:idx_withdrawalSignatures_withdrawal;priority:2;not null"`
	ValidatorAddress string `json:"validatorAddress" gorm:"column:validator_address;index:idx_withdrawalSignatures_validator_address;not null"`
	Signature        string `json:"signature" gorm:"column:signature;not null"`
	Transaction      string `json:"transaction" gorm:"column:transaction;index:idx_withdrawalSignatures_transaction;not null"`
}

func (m WithdrawalSignatures) BeforeCreate(tx *gorm.DB) (err error) {
	return nil
}

func (m WithdrawalSignatures) TableName() string {
	return "withdrawal_signatures"
}
