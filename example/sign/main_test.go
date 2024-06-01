package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func createTicketCell(address *address.Address, amount *big.Int, nonce, expire, payout_id uint64) (*cell.Cell, error) {
	c := cell.BeginCell()
	c.MustStoreAddr(address)
	c.MustStoreCoins(amount.Uint64())
	c.MustStoreUInt(nonce, 64)
	c.MustStoreUInt(expire, 64)
	c.MustStoreUInt(payout_id, 64)
	return c.EndCell(), nil
}

func signTicketCellByPrivKey() ([]byte, ed25519.PublicKey, error) {
	w, err := wallet.FromSeed(nil, strings.Split(os.Getenv("WALLET_SEED"), " "), wallet.V4R2) // 24 mnemonic words
	if err != nil {
		return nil, nil, err
	}
	fmt.Printf("wallet address: %s\n", w.Address().String())
	return []byte(w.PrivateKey()), w.PrivateKey().Public().(ed25519.PublicKey), nil
}

func signTicketCell() ([]byte, error) {
	addr := address.MustParseAddr("0QBeE6QYHVrtjrCemwUYnGYZFVM9MkJXSwsURWZF13FkcSMn")
	amount := new(big.Int).SetUint64(1_000_000_000 / 100) // 0.01 TON
	// oneWeekLater := uint64(time.Now().Unix() + 7*24*60*60)
	oneWeekLater := uint64(21234567890)
	nonce := uint64(0)
	payout_id := uint64(1111)
	c, err := createTicketCell(addr, amount, nonce, oneWeekLater, payout_id)
	if err != nil {
		return nil, err
	}
	priv, pub, err := signTicketCellByPrivKey()
	if err != nil {
		return nil, err
	}

	signature := c.Sign(priv)
	fmt.Printf("signature: %s\n", hex.EncodeToString(signature))
	if !c.Verify(pub, signature) {
		return nil, fmt.Errorf("signature verification failed")
	}

	builder := cell.BeginCell().
		MustStoreSlice(signature, 64*8).
		MustStoreCoins(amount.Uint64()).
		MustStoreUInt(nonce, 64).
		MustStoreUInt(oneWeekLater, 64).
		MustStoreUInt(payout_id, 64)
	for ((1023 - builder.BitsLeft()) % 8) != 0 {
		if err := builder.StoreBoolBit(false); err != nil {
			fmt.Printf("store padding error: %s\n", err)
		}
	}
	builderCell := builder.EndCell()
	fmt.Printf("BitsSize: %d\n", builderCell.BitsSize())
	signatureWithPadding := builderCell.BeginParse().MustLoadSlice(builderCell.BitsSize())
	fmt.Printf("signatureWithPadding: %s\n", hex.EncodeToString(signatureWithPadding))
	base64Encoded := base64.RawURLEncoding.EncodeToString(signatureWithPadding)
	// GRXkWDvEnKBOvBa7UM8h97sDGZEtB3p-5F9NgdMUEMmsWM4-MSTaWEbvf17RF5vFDeVeiL8khQa55ATPb1NABDmJaAAAAAAAAAAAAAAAAE8a3K0gAAAAAAAARXA
	// GRXkWDvEnKBOvBa7UM8h97sDGZEtB3p-5F9NgdMUEMmsWM4-MSTaWEbvf17RF5vFDeVeiL8khQa55ATPb1NABDmJaAAAAAAAAAAAAAAAAE8a3K0gAAAAAAAARXA
	fmt.Printf("base64Encoded: %s\n", base64Encoded)

	return signatureWithPadding, nil
}

func TestSignVerify(t *testing.T) {
	b, err := signTicketCell()
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Println(hex.EncodeToString(b))
}
