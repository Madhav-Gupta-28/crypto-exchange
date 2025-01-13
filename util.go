package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Convert ETH to Wei
func EthToWei(eth float64) *big.Int {
	wei := new(big.Float).Mul(big.NewFloat(eth), big.NewFloat(1e18))
	weiInt, _ := wei.Int(nil) // Convert to big.Int
	return weiInt
}

func TransferETH(client *ethclient.Client, from *ecdsa.PrivateKey, to string, amount float64) error {
	// Convert ETH to Wei
	value := EthToWei(amount)

	// Since we already have the private key, we can use it directly
	publicKey := from.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}

	gasLimit := uint64(21000) // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	toAddress := common.HexToAddress(to)
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	chainID := big.NewInt(1337)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), from)
	if err != nil {
		return err
	}

	return client.SendTransaction(context.Background(), signedTx)

}
