package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"math/big"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/tyler-smith/go-bip39"
)

// AccountKey For Storing the PrivateKey with a given username
type AccountKey struct {
	Name string
	Key  *ecdsa.PrivateKey
}

func main() {
	// mnemonic := "despair lamp valley faint pull taxi proud panic vibrant truck paper genuine swap permit assist scare found verify moon still volume monkey pony journey"
	mnemonic, err := generateMnemonic()
	if err != nil {
		fmt.Printf("Error: %+v", err)
	}

	privateKey, publicKey, path, err := hdWallet(*mnemonic)
	if err != nil {
		fmt.Printf("Error: %+v", err)
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := hexutil.Encode(privateKeyBytes) // hexutil.Encode(privateKeyBytes)[2:] for without 0x
	publicKeyBytes := crypto.FromECDSAPub(publicKey)
	publicKeyHex := hexutil.Encode(publicKeyBytes[1:])
	walletAddress := crypto.PubkeyToAddress(*publicKey).Hex()

	client, err := ethclient.Dial("https://rinkeby.infura.io/v3/my-api-key")
	if err != nil {
		log.Fatal(err)
	}

	// Display mnemonic and keys
	fmt.Println("Mnemonic: ", *mnemonic)
	fmt.Println("ETH Private Key: ", privateKeyHex)
	fmt.Println("ETH Public Key: ", publicKeyHex)
	fmt.Println("ETH Wallet Address: ", walletAddress)
	fmt.Println("Path: ", *path)

	nonce, err := client.PendingNonceAt(context.Background(), crypto.PubkeyToAddress(*publicKey))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Nonce: ", nonce)

	value := big.NewInt(100000000000000000) // in wei (.1 eth)
	gasLimit := uint64(21000)               // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Gas Price: ", gasPrice)

	toAddress := common.HexToAddress("0x12B974EFB65C98c830107fCbbE27c5b85E7359f8")
	var data []byte
	var tx *types.Transaction
	tx = types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("ChainID: ", chainID)

	balance, err := client.BalanceAt(context.Background(), common.HexToAddress(walletAddress), nil)
	if err != nil {
		log.Fatal(err)
	}
	ethbalance := new(big.Float)
	ethbalance.SetString(balance.String())
	ethValue := new(big.Float).Quo(ethbalance, big.NewFloat(math.Pow10(18)))
	fmt.Println("ETH Balance: ", ethValue)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	ts := types.Transactions{signedTx}
	rawTx := hex.EncodeToString(ts.GetRlp(0))
	fmt.Println("Raw TX: ", rawTx)

	rawTxBytes, err := hex.DecodeString(rawTx)
	rlp.DecodeBytes(rawTxBytes, &tx)

	err = client.SendTransaction(context.Background(), tx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("TX sent: %s", tx.Hash().Hex())
}

func generateMnemonic() (*string, error) {
	// Generate a mnemonic for memorization or user-friendly seeds
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return nil, err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, err
	}
	return &mnemonic, nil
}

func hdWallet(mnemonic string) (*ecdsa.PrivateKey, *ecdsa.PublicKey, *string, error) {
	// Generate a Bip32 HD wallet for the mnemonic and a user supplied password
	seed := bip39.NewSeed(mnemonic, "")
	// Generate a new master node using the seed.
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, nil, nil, err
	}
	//fmt.Println("Master Public Key: ", masterKey.PublicKey())
	// This gives the path: m/44H
	acc44H, err := masterKey.Child(hdkeychain.HardenedKeyStart + 44)
	if err != nil {
		return nil, nil, nil, err
	}
	// This gives the path: m/44H/60H
	acc44H60H, err := acc44H.Child(hdkeychain.HardenedKeyStart + 60)
	if err != nil {
		return nil, nil, nil, err
	}
	// This gives the path: m/44H/60H/0H
	acc44H60H0H, err := acc44H60H.Child(hdkeychain.HardenedKeyStart + 0)
	if err != nil {
		return nil, nil, nil, err
	}
	// This gives the path: m/44H/60H/0H/0
	acc44H60H0H0, err := acc44H60H0H.Child(0)
	if err != nil {
		return nil, nil, nil, err
	}
	// This gives the path: m/44H/60H/0H/0/0
	acc44H60H0H00, err := acc44H60H0H0.Child(0)
	if err != nil {
		return nil, nil, nil, err
	}
	btcecPrivKey, err := acc44H60H0H00.ECPrivKey()
	if err != nil {
		return nil, nil, nil, err
	}
	privateKey := btcecPrivKey.ToECDSA()
	publicKey := &privateKey.PublicKey
	path := "m/44H/60H/0H/0/0"
	return privateKey, publicKey, &path, nil
}

// SaveKey - save key as file in current working directory
func (ac *AccountKey) SaveKey() error {
	err := crypto.SaveECDSA(ac.Name, ac.Key)
	return err
}

// LoadKey - Load key (if it exists) from current directory
func (ac *AccountKey) LoadKey() error {
	key, err := crypto.LoadECDSA(ac.Name)
	if err == nil {
		ac.Key = key
	}
	return err
}
