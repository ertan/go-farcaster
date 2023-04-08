package registry

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	BLOCK_RANGE          = 2000
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
	logger   *log.Logger
}

func NewRegistryService(providerWs string) *RegistryService {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	if providerWs == "" {
		logger.Println("providerWs is empty, not connecting to the blockchain")
		return nil
	}
	firAbiJson, err := ioutil.ReadFile("pkg/registry/abi/IdRegistryV2.json")
	if err != nil {
		logger.Fatal("Error when opening file: ", err)
	}
	firAbi, err := abi.JSON(strings.NewReader(string(firAbiJson)))
	if err != nil {
		logger.Fatal("Error when parsing abi: ", err)
	}
	fnrAbiJson, err := ioutil.ReadFile("pkg/registry/abi/NameRegistryV2.json")
	if err != nil {
		logger.Fatal("Error when opening file: ", err)
	}
	fnrAbi, err := abi.JSON(strings.NewReader(string(fnrAbiJson)))
	if err != nil {
		logger.Fatal("Error when parsing abi: ", err)
	}
	logger.Println("Connecting to Ethereum node: ", providerWs)
	client, err := ethclient.Dial(providerWs)
	if err != nil {
		logger.Fatal("Error when dialing: ", err)
	}
	logger.Println("Connected to Ethereum node: ", providerWs)
	registry := &RegistryService{
		// read from IdRegistryV2.json into json.Unmarshal()
		firAbi:   firAbi,
		firBlock: FIR_DEPLOYMENT_BLOCK,
		fir:      make(map[uint64]string),
		fnrAbi:   fnrAbi,
		fnrBlock: FNR_DEPLOYMENT_BLOCK,
		fnr:      make(map[string]string),
		client:   client,
		logger:   logger,
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
	header, err := r.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return err
	}
	currentBlock := header.Number.Uint64()
	r.logger.Println("Syncing registry until block ", currentBlock)
	errChan := make(chan error)

	go func() {
		if err := r.syncFirLogs(currentBlock); err != nil {
			errChan <- err
		}
		wg.Done()
	}()

	go func() {
		if err := r.syncFnrLogs(currentBlock); err != nil {
			errChan <- err
		}
		wg.Done()
	}()

	wg.Wait()
	r.logger.Println("Synced registry until block ", currentBlock)

	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

func (r *RegistryService) syncFirLogs(blockNo uint64) error {
	contractAddress := common.HexToAddress(FIR_CONTRACT_ADDRESS)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{
			contractAddress,
		},
		Topics: [][]common.Hash{
			{common.HexToHash(REGISTER_TOPIC)},
		},
	}
	for i := r.firBlock; i < blockNo; i += BLOCK_RANGE {
		query.FromBlock = new(big.Int).SetUint64(i)
		query.ToBlock = new(big.Int).SetUint64(i + BLOCK_RANGE)
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
			r.logger.Println("Register event: ", address.Hex(), fid)
			// TODO(ertan): Is this needed?
			// data, err := r.firAbi.Unpack("Register", vLog.Data)
			// if err != nil {
			// 	return err
			// }
			// recovery := data[0].(common.Address).String()
			// url := data[1].(string)
			r.fir[fid] = strings.ToLower(address.Hex())
		}
		time.Sleep(time.Millisecond * 100)
	}
	r.firBlock = blockNo
	return nil
}

func (r *RegistryService) syncFnrLogs(blockNo uint64) error {
	contractAddress := common.HexToAddress(FNR_CONTRACT_ADDRESS)

	query := ethereum.FilterQuery{
		Addresses: []common.Address{
			contractAddress,
		},
		Topics: [][]common.Hash{
			{common.HexToHash(TRANSFER_TOPIC)},
		},
	}
	for i := r.fnrBlock; i < blockNo; i += BLOCK_RANGE {
		query.FromBlock = new(big.Int).SetUint64(i)
		query.ToBlock = new(big.Int).SetUint64(i + BLOCK_RANGE)
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
			r.logger.Println("Transfer event: ", strings.ToLower(address.Hex()), string(fname))
		}
		time.Sleep(time.Millisecond * 100)
	}
	r.fnrBlock = blockNo
	return nil
}
