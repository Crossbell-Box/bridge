package task

import "github.com/axieinfinity/bridge-v2/models"

const (
	ACK_WITHDREW_TASK     = "acknowledgeWithdrew"
	DEPOSIT_TASK          = "deposit"
	WITHDRAWAL_TASK       = "withdrawal"
	WITHDRAWAL_AGAIN_TASK = "withdrawAgain"

	STATUS_PENDING    = "pending"
	STATUS_FAILED     = "failed"
	STATUS_PROCESSING = "processing"
	STATUS_DONE       = "done"

	GATEWAY_CONTRACT              = "Gateway"
	GOVERNANCE_CONTRACT           = "Governance"
	TRUSTED_ORGANIZATION_CONTRACT = "TrustedOrganization"
	ETH_GOVERNANCE_CONTRACT       = "EthGovernance"
	ETH_GATEWAY_CONTRACT          = "EthGateway"
	BRIDGEADMIN_CONTRACT          = "BridgeAdmin"

	CROSSBELL_GATEWAY_CONTRACT = "CrossbellGateway"
	CROSSBELL_VALIDATOR        = "CrossbellValidator"
	MAINCHAIN_GATEWAY_CONTRACT = "MainchainGateway"
	MAINCHAIN_VALIDATOR        = "MainchainValidator"
)

type Tasker interface {
	collectTask(t *models.Task)
	send()
}
