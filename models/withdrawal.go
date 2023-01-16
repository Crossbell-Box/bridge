package models

import (
	"gorm.io/gorm"
)

type Withdrawal struct {
	ID                    int    `json:"id" gorm:"primary_key:true;column:id;auto_increment;not null"`
	MainChainId           int64  `json:"mainchainId" gorm:"column:mainchain_id;uniqueIndex:idx_withdrawal;not null"`
	WithdrawalId          int64  `json:"withdrawalId" gorm:"column:withdrawal_id;uniqueIndex:idx_withdrawal;not null"`
	RecipientAddress      string `json:"recipientAddress" gorm:"column:recipient_address;index:idx_withdrawal_recipient_address;not null"`
	MainchainTokenAddress string `json:"mainchainTokenAddress" gorm:"column:mainchain_token_address;index:idx_withdrawal_mainchain_token_address;not null"`
	TokenQuantity         string `json:"tokenQuantity" gorm:"column:token_quantity;not null"`
	Fee                   string `json:"fee" gorm:"column:fee;not null"`
	WithdrawerAddress     string `json:"withdrawerAddress" gorm:"column:withdrawer_address;index:idx_withdrawal_withdrawer_address;not null"`
	Transaction           string `json:"transaction" gorm:"column:transaction;index:idx_withdrawal_transaction;not null"`
}

func (m Withdrawal) BeforeCreate(tx *gorm.DB) (err error) {
	return nil
}

func (m Withdrawal) TableName() string {
	return "withdrawal"
}
