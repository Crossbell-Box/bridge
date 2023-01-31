package models

import (
	"gorm.io/gorm"
)

type RequestWithdrawal struct {
	ID                    int    `json:"id" gorm:"primary_key:true;column:id;auto_increment;not null"`
	MainchainId           int64  `json:"mainchainId" gorm:"column:mainchain_id;uniqueIndex:idx_withdrawal;not null"`
	WithdrawalId          int64  `json:"withdrawalId" gorm:"column:withdrawal_id;uniqueIndex:idx_withdrawal;not null"`
	RecipientAddress      string `json:"recipientAddress" gorm:"column:recipient_address;index:idx_withdrawal_recipient_address;not null"`
	MainchainTokenAddress string `json:"mainchainTokenAddress" gorm:"column:mainchain_token_address;index:idx_withdrawal_mainchain_token_address;not null"`
	TokenQuantity         string `json:"tokenQuantity" gorm:"column:token_quantity;not null"`
	Fee                   string `json:"fee" gorm:"column:fee;not null"`
	Transaction           string `json:"transaction" gorm:"column:transaction;index:idx_withdrawal_transaction;not null"`
	Status                string `json:"status" gorm:"column:status"`
	WithdrawalTransaction string `json:"withdrawalTransaction" gorm:"column:withdrawal_transaction"`
}

func (m RequestWithdrawal) BeforeCreate(tx *gorm.DB) (err error) {
	return nil
}

func (m RequestWithdrawal) TableName() string {
	return "request_withdrawal"
}
