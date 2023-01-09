package models

import (
	"gorm.io/gorm"
)

type Withdrawal struct {
	ID                   int    `json:"id" gorm:"primary_key:true;column:id;auto_increment;not null"`
	WithdrawalId         int64  `json:"withdrawalId" gorm:"column:withdrawal_id;unique;not null"`
	ExternalAddress      string `json:"externalAddress" gorm:"column:external_address;index:idx_withdrawal_external_address;not null"`
	ExternalTokenAddress string `json:"externalTokenAddress" gorm:"column:external_token_address;index:idx_withdrawal_external_token_address;not null"`
	ExternalChainId      int64  `json:"externalChainId" gorm:"column:external_chain_id;not null"`
	RecipientAddress     string `json:"roninAddress" gorm:"column:ronin_address;index:idx_withdrawal_receipt_address;not null"`
	TokenQuantity        string `json:"tokenQuantity" gorm:"column:token_quantity;not null"`
	Transaction          string `json:"transaction" gorm:"column:transaction;index:idx_withdrawal_transaction;not null"`
}

func (m Withdrawal) BeforeCreate(tx *gorm.DB) (err error) {
	return nil
}

func (m Withdrawal) TableName() string {
	return "withdrawal"
}
