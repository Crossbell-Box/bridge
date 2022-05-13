// Code generated by mockery v2.12.1. DO NOT EDIT.

package mocks

import (
	abi "github.com/ethereum/go-ethereum/accounts/abi"
	apitypes "github.com/ethereum/go-ethereum/signer/core/apitypes"

	big "math/big"

	bind "github.com/ethereum/go-ethereum/accounts/abi/bind"

	common "github.com/ethereum/go-ethereum/common"

	ecdsa "crypto/ecdsa"

	ethclient "github.com/ethereum/go-ethereum/ethclient"

	hexutil "github.com/ethereum/go-ethereum/common/hexutil"

	mock "github.com/stretchr/testify/mock"

	reflect "reflect"

	testing "testing"

	time "time"

	types "github.com/ethereum/go-ethereum/core/types"

	utils "github.com/axieinfinity/bridge-v2/internal/utils"
)

// IUtils is an autogenerated mock type for the IUtils type
type IUtils struct {
	mock.Mock
}

// FilterLogs provides a mock function with given fields: client, opts, contractAddresses, filteredMethods
func (_m *IUtils) FilterLogs(client utils.EthClient, opts *bind.FilterOpts, contractAddresses []common.Address, filteredMethods map[*abi.ABI]map[string]struct{}) ([]types.Log, error) {
	ret := _m.Called(client, opts, contractAddresses, filteredMethods)

	var r0 []types.Log
	if rf, ok := ret.Get(0).(func(utils.EthClient, *bind.FilterOpts, []common.Address, map[*abi.ABI]map[string]struct{}) []types.Log); ok {
		r0 = rf(client, opts, contractAddresses, filteredMethods)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.Log)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(utils.EthClient, *bind.FilterOpts, []common.Address, map[*abi.ABI]map[string]struct{}) error); ok {
		r1 = rf(client, opts, contractAddresses, filteredMethods)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetArguments provides a mock function with given fields: a, name, data, isInput
func (_m *IUtils) GetArguments(a abi.ABI, name string, data []byte, isInput bool) (abi.Arguments, error) {
	ret := _m.Called(a, name, data, isInput)

	var r0 abi.Arguments
	if rf, ok := ret.Get(0).(func(abi.ABI, string, []byte, bool) abi.Arguments); ok {
		r0 = rf(a, name, data, isInput)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(abi.Arguments)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(abi.ABI, string, []byte, bool) error); ok {
		r1 = rf(a, name, data, isInput)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Invoke provides a mock function with given fields: any, name, args
func (_m *IUtils) Invoke(any interface{}, name string, args ...interface{}) (reflect.Value, error) {
	var _ca []interface{}
	_ca = append(_ca, any, name)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	var r0 reflect.Value
	if rf, ok := ret.Get(0).(func(interface{}, string, ...interface{}) reflect.Value); ok {
		r0 = rf(any, name, args...)
	} else {
		r0 = ret.Get(0).(reflect.Value)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(interface{}, string, ...interface{}) error); ok {
		r1 = rf(any, name, args...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LoadAbi provides a mock function with given fields: path
func (_m *IUtils) LoadAbi(path string) (*abi.ABI, error) {
	ret := _m.Called(path)

	var r0 *abi.ABI
	if rf, ok := ret.Get(0).(func(string) *abi.ABI); ok {
		r0 = rf(path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*abi.ABI)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewEthClient provides a mock function with given fields: url
func (_m *IUtils) NewEthClient(url string) (utils.EthClient, error) {
	ret := _m.Called(url)

	var r0 utils.EthClient
	if rf, ok := ret.Get(0).(func(string) utils.EthClient); ok {
		r0 = rf(url)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(utils.EthClient)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(url)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SendContractTransaction provides a mock function with given fields: key, chainId, fn
func (_m *IUtils) SendContractTransaction(key *ecdsa.PrivateKey, chainId *big.Int, fn func(*bind.TransactOpts) (*types.Transaction, error)) (*types.Transaction, error) {
	ret := _m.Called(key, chainId, fn)

	var r0 *types.Transaction
	if rf, ok := ret.Get(0).(func(*ecdsa.PrivateKey, *big.Int, func(*bind.TransactOpts) (*types.Transaction, error)) *types.Transaction); ok {
		r0 = rf(key, chainId, fn)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Transaction)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*ecdsa.PrivateKey, *big.Int, func(*bind.TransactOpts) (*types.Transaction, error)) error); ok {
		r1 = rf(key, chainId, fn)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SignTypedData provides a mock function with given fields: typedData, privateKey
func (_m *IUtils) SignTypedData(typedData apitypes.TypedData, privateKey *ecdsa.PrivateKey) (hexutil.Bytes, error) {
	ret := _m.Called(typedData, privateKey)

	var r0 hexutil.Bytes
	if rf, ok := ret.Get(0).(func(apitypes.TypedData, *ecdsa.PrivateKey) hexutil.Bytes); ok {
		r0 = rf(typedData, privateKey)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(hexutil.Bytes)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(apitypes.TypedData, *ecdsa.PrivateKey) error); ok {
		r1 = rf(typedData, privateKey)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SubscribeTransactionReceipt provides a mock function with given fields: client, tx, ticker, maxTry
func (_m *IUtils) SubscribeTransactionReceipt(client *ethclient.Client, tx *types.Transaction, ticker time.Duration, maxTry int) error {
	ret := _m.Called(client, tx, ticker, maxTry)

	var r0 error
	if rf, ok := ret.Get(0).(func(*ethclient.Client, *types.Transaction, time.Duration, int) error); ok {
		r0 = rf(client, tx, ticker, maxTry)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Title provides a mock function with given fields: text
func (_m *IUtils) Title(text string) string {
	ret := _m.Called(text)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(text)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// UnpackToInterface provides a mock function with given fields: a, name, data, isInput, v
func (_m *IUtils) UnpackToInterface(a abi.ABI, name string, data []byte, isInput bool, v interface{}) error {
	ret := _m.Called(a, name, data, isInput, v)

	var r0 error
	if rf, ok := ret.Get(0).(func(abi.ABI, string, []byte, bool, interface{}) error); ok {
		r0 = rf(a, name, data, isInput, v)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewIUtils creates a new instance of IUtils. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewIUtils(t testing.TB) *IUtils {
	mock := &IUtils{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
