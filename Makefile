mock: ## generate mock
	mockery --output internal/mocks/types --dir internal/types --name IListener
	mockery --output internal/mocks/types --dir internal/types --name ITransaction
	mockery --output internal/mocks/types --dir internal/types --name ILog
	mockery --output internal/mocks/types --dir internal/types --name IReceipt
	mockery --output internal/mocks/types --dir internal/types --name IBlock
	mockery --output internal/mocks/types --dir internal/types --name IJob
	mockery --output internal/mocks/utils --dir internal/utils --name IUtils
	mockery --output internal/mocks/utils --dir internal/utils --name EthClient

bridge:
	go install ./cmd/bridge
	@echo "Done building."
	@echo "Run \"bridge\" to launch bridge."

abigen:
	mkdir -p build/contract/
	cd crossbell-bridge-contracts/ && \
	git pull && \
	yarn && \
	make install && \
	forge inspect MainchainGateway abi > ../build/contract/MainchainGateway.abi && \
	forge inspect CrossbellGateway abi > ../build/contract/CrossbellGateway.abi && \
	forge inspect MainchainGateway b > ../build/contract/MainchainGateway.bin && \
	forge inspect CrossbellGateway b > ../build/contract/CrossbellGateway.bin && \
	cd ../ && \
	mkdir -p bridge-contracts/generated_contracts/mainchainGateway
	mkdir -p bridge-contracts/generated_contracts/crossbellGateway
	abigen --bin=build/contract/MainchainGateway.bin --abi=build/contract/MainchainGateway.abi --pkg=mainchainGateway --out=bridge-contracts/generated_contracts/crossbellGateway/crossbellGateway.go
	abigen --bin=build/contract/CrossbellGateway.bin --abi=build/contract/CrossbellGateway.abi --pkg=crossbellGateway --out=bridge-contracts/generated_contracts/mainchainGateway/mainchainGateway.go
run:
	@cd cmd/bridge && go run main.go
