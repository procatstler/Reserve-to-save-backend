package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// HashString creates SHA256 hash of a string
func HashString(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifySignature verifies an Ethereum signature
func VerifySignature(message, signature string, expectedAddress string) (bool, error) {
	// Remove 0x prefix if present
	sig := strings.TrimPrefix(signature, "0x")
	
	// Decode signature
	sigBytes, err := hexutil.Decode("0x" + sig)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %w", err)
	}

	// Ethereum signatures are 65 bytes (r + s + v)
	if len(sigBytes) != 65 {
		return false, fmt.Errorf("invalid signature length: %d", len(sigBytes))
	}

	// Transform V from Ethereum format (27/28) to standard format (0/1)
	if sigBytes[64] == 27 || sigBytes[64] == 28 {
		sigBytes[64] -= 27
	}

	// Hash the message with Ethereum prefix
	messageHash := accounts.TextHash([]byte(message))

	// Recover public key
	pubKey, err := crypto.SigToPub(messageHash, sigBytes)
	if err != nil {
		return false, fmt.Errorf("failed to recover public key: %w", err)
	}

	// Get address from public key
	recoveredAddress := crypto.PubkeyToAddress(*pubKey)

	// Compare addresses (case-insensitive)
	return strings.EqualFold(recoveredAddress.Hex(), expectedAddress), nil
}

// GenerateNonce generates a random nonce
func GenerateNonce() string {
	bytes := make([]byte, 16)
	crypto.RandRead(bytes)
	return hex.EncodeToString(bytes)
}

// IsValidAddress checks if a string is a valid Ethereum address
func IsValidAddress(address string) bool {
	return common.IsHexAddress(address)
}

// NormalizeAddress converts address to checksummed format
func NormalizeAddress(address string) string {
	if !IsValidAddress(address) {
		return ""
	}
	return common.HexToAddress(address).Hex()
}

// CreateSignMessage creates an EIP-4361 compatible sign-in message
func CreateSignMessage(domain, address, chainID, nonce, issuedAt, expiresAt, requestID string) string {
	return fmt.Sprintf(`%s wants you to sign in with your wallet:
%s

URI: %s
Version: 1
Chain ID: %s
Nonce: %s
Issued At: %s
Expiration Time: %s
Request ID: %s
Statement: Sign to authenticate with R2S platform.`,
		domain, address, domain, chainID, nonce, issuedAt, expiresAt, requestID)
}