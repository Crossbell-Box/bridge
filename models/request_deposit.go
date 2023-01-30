package models

import (
	"gorm.io/gorm"
)

type RequestDeposit struct {
	ID                    int    `json:"id" gorm:"primary_key:true;column:id;auto_increment;not null"`
	MainchainId           int64  `json:"mainchainId" gorm:"column:mainchain_id;uniqueIndex:idx_deposit;not null"`
	DepositId             int64  `json:"depositId" gorm:"column:deposit_id;uniqueIndex:idx_deposit;not null"`
	RecipientAddress      string `json:"recipientAddress" gorm:"column:recipient_address;index:idx_deposit_recipient_address;not null"`
	CrossbellTokenAddress string `json:"crossbellTokenAddress" gorm:"column:crossbell_token_address;index:idx_deposit_crossbell_token_address;not null"`
	TokenQuantity         string `json:"tokenQuantity" gorm:"column:token_quantity;not null"`
	Transaction           string `json:"transaction" gorm:"column:transaction;index:idx_deposit_transaction;not null"`
	Status                string `json:"status" gorm:"column:status"`
}

func (m RequestDeposit) BeforeCreate(tx *gorm.DB) (err error) {
	return nil
}

func (m RequestDeposit) TableName() string {
	return "request_deposit"
}
