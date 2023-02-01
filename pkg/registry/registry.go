package registry

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	FIR_DEPLOYMENT_BLOCK = 7_648_795
	FNR_DEPLOYMENT_BLOCK = 7_648_795
	FIR_CONTRACT_ADDRESS = "0xda107a1caf36d198b12c16c7b6a1d1c795978c42"
	FNR_CONTRACT_ADDRESS = "0xe3be01d99baa8db9905b33a3ca391238234b79d1"
	REGISTER_TOPIC       = "0x3cd6a0ffcc37406d9958e09bba79ff19d8237819eb2e1911f9edbce656499c87"
	TRANSFER_TOPIC       = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
)

type RegistryService struct {
	firAbi   abi.ABI
	firBlock uint64
	fir      map[uint64]string // fid -> address
	fnrAbi   abi.ABI
	fnrBlock uint64
	fnr      map[string]string // username -> address
	client   *ethclient.Client
}

func NewRegistryService(providerWs string) *RegistryService {
	if providerWs == "" {
		log.Println("providerWs is empty, not connecting to the blockchain")
		return nil
	}
	firAbiJson, err := ioutil.ReadFile("pkg/registry/abi/IdRegistryV2.json")
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}
	firAbi, err := abi.JSON(strings.NewReader(string(firAbiJson)))
	if err != nil {
		log.Fatal("Error when parsing abi: ", err)
	}
	fnrAbiJson, err := ioutil.ReadFile("pkg/registry/abi/NameRegistryV2.json")
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}
	fnrAbi, err := abi.JSON(strings.NewReader(string(fnrAbiJson)))
	if err != nil {
		log.Fatal("Error when parsing abi: ", err)
	}
	fmt.Println("Connecting to Ethereum node: ", providerWs)
	client, err := ethclient.Dial(providerWs)
	if err != nil {
		log.Fatal("Error when dialing: ", err)
	}
	registry := &RegistryService{
		// read from IdRegistryV2.json into json.Unmarshal()
		firAbi:   firAbi,
		firBlock: FIR_DEPLOYMENT_BLOCK,
		fir:      make(map[uint64]string),
		fnrAbi:   fnrAbi,
		fnrBlock: FNR_DEPLOYMENT_BLOCK,
		fnr:      make(map[string]string),
		client:   client,
	}
	err = registry.sync()
	if err != nil {
		log.Fatal("Error when syncing registry: ", err)
	}
	return registry
}

func (r *RegistryService) GetFidByAddress(address string) (uint64, error) {
	for fid, addr := range r.fir {
		if addr == address {
			return fid, nil
		}
	}
	return 0, errors.New("address not found")
}

func (r *RegistryService) GetFidByFname(fname string) (uint64, error) {
	if address, ok := r.fnr[fname]; ok {
		for fid, addr := range r.fir {
			if addr == address {
				return fid, nil
			}
		}
	}
	return 0, errors.New("fname not found")
}

func (r *RegistryService) GetAddressByFname(fname string) (string, error) {
	if address, ok := r.fnr[fname]; ok {
		return address, nil
	}
	return "", errors.New("fname not found")
}

func (r *RegistryService) GetAddressByFid(fid uint64) (string, error) {
	if address, ok := r.fir[fid]; ok {
		return address, nil
	}
	return "", errors.New("fid not found")
}

func (r *RegistryService) GetFnameByFid(fid uint64) (string, error) {
	if address, ok := r.fir[fid]; ok {
		if fname, ok := r.fnr[address]; ok {
			return fname, nil
		}
	}
	return "", errors.New("fid not found")
}

func (r *RegistryService) GetFnameByAddress(address string) (string, error) {
	address = strings.ToLower(address)
	for fname, addr := range r.fnr {
		if addr == address {
			return fname, nil
		}
	}
	return "", errors.New("address not found")
}

func (r *RegistryService) sync() error {
	var wg sync.WaitGroup
	wg.Add(2)

	errChan := make(chan error)

	go func() {
		if err := r.syncFirLogs(); err != nil {
			errChan <- err
		}
		wg.Done()
	}()

	go func() {
		if err := r.syncFnrLogs(); err != nil {
			errChan <- err
		}
		wg.Done()
	}()

	wg.Wait()

	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

func (r *RegistryService) syncFirLogs() error {
	contractAddress := common.HexToAddress(FIR_CONTRACT_ADDRESS)
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(r.firBlock),
		Addresses: []common.Address{
			contractAddress,
		},
		Topics: [][]common.Hash{
			{common.HexToHash(REGISTER_TOPIC)},
		},
	}
	logs, err := r.client.FilterLogs(context.Background(), query)
	if err != nil {
		return err
	}
	for _, vLog := range logs {
		blockNumber := vLog.BlockNumber
		if blockNumber > r.firBlock {
			r.firBlock = blockNumber
		}
		address := common.BytesToAddress(vLog.Topics[1].Bytes())
		fid := vLog.Topics[2].Big().Uint64()
		// println("Register event: ", address.Hex(), fid)
		// TODO(ertan): Is this needed?
		// data, err := r.firAbi.Unpack("Register", vLog.Data)
		// if err != nil {
		// 	return err
		// }
		// recovery := data[0].(common.Address).String()
		// url := data[1].(string)
		r.fir[fid] = strings.ToLower(address.Hex())
	}
	return nil
}

func (r *RegistryService) syncFnrLogs() error {
	contractAddress := common.HexToAddress(FNR_CONTRACT_ADDRESS)
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(r.fnrBlock),
		Addresses: []common.Address{
			contractAddress,
		},
		Topics: [][]common.Hash{
			{common.HexToHash(TRANSFER_TOPIC)},
		},
	}
	logs, err := r.client.FilterLogs(context.Background(), query)
	if err != nil {
		return err
	}
	for _, vLog := range logs {
		blockNumber := vLog.BlockNumber
		if blockNumber > r.fnrBlock {
			r.fnrBlock = blockNumber
		}
		address := common.BytesToAddress(vLog.Topics[2].Bytes())
		fnameHexStr := vLog.Topics[3].Hex()
		fname, err := hexutil.Decode(fnameHexStr)
		if err != nil {
			return err
		}
		fname = common.TrimRightZeroes(fname)
		r.fnr[string(fname)] = strings.ToLower(address.Hex())
		// println("Transfer event: ", strings.ToLower(address.Hex()), string(fname))
	}
	return nil
}
