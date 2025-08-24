package services

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	
	"r2s/pkg/contracts"
)

type TransactionService struct {
	client         *ethclient.Client
	factoryAddress common.Address
	usdtAddress    common.Address
	chainID        *big.Int
}

type TransactionMessage struct {
	To       string          `json:"to"`
	From     string          `json:"from"`
	Data     string          `json:"data"`
	Value    string          `json:"value"`
	GasLimit uint64          `json:"gasLimit"`
	GasPrice string          `json:"gasPrice"`
	Nonce    uint64          `json:"nonce"`
	ChainID  string          `json:"chainId"`
}

func NewTransactionService(rpcURL, factoryAddress, usdtAddress string) *TransactionService {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to blockchain: %v", err))
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		panic(fmt.Sprintf("Failed to get chain ID: %v", err))
	}

	return &TransactionService{
		client:         client,
		factoryAddress: common.HexToAddress(factoryAddress),
		usdtAddress:    common.HexToAddress(usdtAddress),
		chainID:        chainID,
	}
}

// BuildJoinCampaignTx creates a transaction message for joining a campaign
func (s *TransactionService) BuildJoinCampaignTx(
	userAddress string,
	campaignAddress string,
	amount *big.Int,
) (*TransactionMessage, error) {
	campaign, err := contracts.NewR2scampaign(common.HexToAddress(campaignAddress), s.client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate campaign contract: %w", err)
	}

	// Build transaction data
	auth := &bind.TransactOpts{
		From:  common.HexToAddress(userAddress),
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			// This is just for building the transaction, not signing
			return tx, nil
		},
		NoSend: true,
	}

	// Get ABI
	campaignABI, err := abi.JSON(strings.NewReader(contracts.R2scampaignABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Pack the join function call
	data, err := campaignABI.Pack("join", amount)
	if err != nil {
		return nil, fmt.Errorf("failed to pack join call: %w", err)
	}

	// Estimate gas
	gasLimit, err := s.estimateGas(userAddress, campaignAddress, data)
	if err != nil {
		gasLimit = uint64(300000) // Default gas limit
	}

	// Get gas price
	gasPrice, err := s.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Get nonce
	nonce, err := s.client.PendingNonceAt(context.Background(), common.HexToAddress(userAddress))
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	return &TransactionMessage{
		To:       campaignAddress,
		From:     userAddress,
		Data:     fmt.Sprintf("0x%x", data),
		Value:    "0",
		GasLimit: gasLimit,
		GasPrice: gasPrice.String(),
		Nonce:    nonce,
		ChainID:  s.chainID.String(),
	}, nil
}

// BuildApproveUSDTTx creates a transaction message for approving USDT
func (s *TransactionService) BuildApproveUSDTTx(
	userAddress string,
	spenderAddress string,
	amount *big.Int,
) (*TransactionMessage, error) {
	usdt, err := contracts.NewMockusdt(s.usdtAddress, s.client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate USDT contract: %w", err)
	}

	// Get ABI
	usdtABI, err := abi.JSON(strings.NewReader(contracts.MockusdtABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Pack the approve function call
	data, err := usdtABI.Pack("approve", common.HexToAddress(spenderAddress), amount)
	if err != nil {
		return nil, fmt.Errorf("failed to pack approve call: %w", err)
	}

	// Estimate gas
	gasLimit, err := s.estimateGas(userAddress, s.usdtAddress.Hex(), data)
	if err != nil {
		gasLimit = uint64(100000) // Default gas limit for approve
	}

	// Get gas price
	gasPrice, err := s.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Get nonce
	nonce, err := s.client.PendingNonceAt(context.Background(), common.HexToAddress(userAddress))
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	return &TransactionMessage{
		To:       s.usdtAddress.Hex(),
		From:     userAddress,
		Data:     fmt.Sprintf("0x%x", data),
		Value:    "0",
		GasLimit: gasLimit,
		GasPrice: gasPrice.String(),
		Nonce:    nonce,
		ChainID:  s.chainID.String(),
	}, nil
}

// BuildRequestCancelTx creates a transaction message for requesting cancellation
func (s *TransactionService) BuildRequestCancelTx(
	userAddress string,
	campaignAddress string,
	amount *big.Int,
) (*TransactionMessage, error) {
	// Get ABI
	campaignABI, err := abi.JSON(strings.NewReader(contracts.R2scampaignABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Pack the requestCancel function call
	data, err := campaignABI.Pack("requestCancel", amount)
	if err != nil {
		return nil, fmt.Errorf("failed to pack requestCancel call: %w", err)
	}

	// Estimate gas
	gasLimit, err := s.estimateGas(userAddress, campaignAddress, data)
	if err != nil {
		gasLimit = uint64(200000)
	}

	// Get gas price
	gasPrice, err := s.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Get nonce
	nonce, err := s.client.PendingNonceAt(context.Background(), common.HexToAddress(userAddress))
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	return &TransactionMessage{
		To:       campaignAddress,
		From:     userAddress,
		Data:     fmt.Sprintf("0x%x", data),
		Value:    "0",
		GasLimit: gasLimit,
		GasPrice: gasPrice.String(),
		Nonce:    nonce,
		ChainID:  s.chainID.String(),
	}, nil
}

// GetCampaignInfo retrieves campaign information from blockchain
func (s *TransactionService) GetCampaignInfo(campaignAddress string) (map[string]interface{}, error) {
	campaign, err := contracts.NewR2scampaign(common.HexToAddress(campaignAddress), s.client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate campaign contract: %w", err)
	}

	// Call view functions
	opts := &bind.CallOpts{Context: context.Background()}
	
	// Get campaign parameters
	params, err := campaign.Params(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign params: %w", err)
	}

	// Get current state
	state, err := campaign.GetState(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign state: %w", err)
	}

	// Get current amount
	currentAmount, err := campaign.CurrentAmount(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get current amount: %w", err)
	}

	// Get participant count
	participantCount, err := campaign.GetParticipantCount(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get participant count: %w", err)
	}

	return map[string]interface{}{
		"address":          campaignAddress,
		"merchant":         params.Merchant.Hex(),
		"basePrice":        params.BasePrice.String(),
		"minQuantity":      params.MinQty.String(),
		"targetAmount":     params.TargetAmount.String(),
		"currentAmount":    currentAmount.String(),
		"participantCount": participantCount.String(),
		"lockStart":        params.LockStart.String(),
		"lockEnd":          params.LockEnd.String(),
		"rMaxBps":          params.RMaxBPS,
		"saveFloorBps":     params.SaveFloorBPS,
		"merchantFeeBps":   params.MerchantFeeBPS,
		"opsFeeBps":        params.OpsFeeBPS,
		"state":            state,
	}, nil
}

// EstimateGasPrice returns current gas price
func (s *TransactionService) EstimateGasPrice() (*big.Int, error) {
	return s.client.SuggestGasPrice(context.Background())
}

// estimateGas estimates gas for a transaction
func (s *TransactionService) estimateGas(from, to string, data []byte) (uint64, error) {
	msg := ethereum.CallMsg{
		From: common.HexToAddress(from),
		To:   &common.Address{},
		Data: data,
	}
	
	toAddr := common.HexToAddress(to)
	msg.To = &toAddr
	
	gasLimit, err := s.client.EstimateGas(context.Background(), msg)
	if err != nil {
		return 0, err
	}
	
	// Add 20% buffer
	return gasLimit * 120 / 100, nil
}