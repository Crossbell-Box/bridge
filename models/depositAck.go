package models

import (
	"gorm.io/gorm"
)

type DepositAck struct {
	ID               int    `json:"id" gorm:"primary_key:true;column:id;auto_increment;not null"`
	MainchainId      int64  `json:"mainchainId" gorm:"column:mainchain_id;uniqueIndex:idx_depositAck_deposit;priority:1;not null"`
	DepositId        int64  `json:"depositId" gorm:"column:deposit_id;uniqueIndex:idx_depositAck_deposit;priority:2;not null"`
	RecipientAddress string `json:"recipientAddress" gorm:"column:recipient_address;index:idx_depositAck_recipient_address;not null"`
	ValidatorAddress string `json:"validatorAddress" gorm:"column:validator_address;index:idx_depositAck_validator_address;not null"`
	Transaction      string `json:"transaction" gorm:"column:transaction;index:idx_depositAck_transaction;not null"`
}

func (m DepositAck) BeforeCreate(tx *gorm.DB) (err error) {
	return nil
}

func (m DepositAck) TableName() string {
	return "deposit_ack"
}
