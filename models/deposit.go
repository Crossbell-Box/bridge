package models

import (
	"gorm.io/gorm"
)

type Deposit struct {
	ID                    int    `json:"id" gorm:"primary_key:true;column:id;auto_increment;not null"`
	DepositId             int64  `json:"depositId" gorm:"column:deposit_id;unique;not null"`
	MainChainId           int64  `json:"chainId" gorm:"column:chain_id;index:idx_deposit_chain_id;not null"`
	RecipientAddress      string `json:"recipientAddress" gorm:"column:recipient_address;index:idx_deposit_recipient_address;not null"`
	CrossbellTokenAddress string `json:"crossbellTokenAddress" gorm:"column:crossbell_token_address;index:idx_deposit_crossbell_token_address;not null"`
	TokenQuantity         string `json:"tokenQuantity" gorm:"column:token_quantity;not null"`
	Fee                   string `json:"fee" gorm:"column:fee;not null"`
	FromAddress           string `json:"withdrawerAddress" gorm:"column:withdrawer_address;index:idx_deposit_withdrawer_address;not null"`
	Transaction           string `json:"transaction" gorm:"column:transaction;index:idx_deposit_transaction;not null"`
}

func (m Deposit) BeforeCreate(tx *gorm.DB) (err error) {
	return nil
}

func (m Deposit) TableName() string {
	return "deposit"
}
