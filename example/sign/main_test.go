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

func createWithdrawTicketContentCell(address *address.Address, amount uint64, nonce, expire, payout_id uint32, user_jetton_wallet *address.Address) (*cell.Cell, error) {
	c := cell.BeginCell()
	c.MustStoreUInt(1162710104, 32) // 1162710104 is prefix for withdraw ticket content, can get from contract abi
	c.MustStoreAddr(address)
	c.MustStoreUInt(amount, 64)
	c.MustStoreUInt(uint64(nonce), 32)
	c.MustStoreUInt(uint64(expire), 32)
	c.MustStoreUInt(uint64(payout_id), 32)
	c.MustStoreAddr(user_jetton_wallet)
	return c.EndCell(), nil
}

// TON payouts
func createTonWithdrawTicketContentCell(address *address.Address, amount uint64, nonce, expire, payout_id uint64) (*cell.Cell, error) {
	c := cell.BeginCell()
	c.MustStoreUInt(1763882716, 32) // 1763882716 is prefix for withdraw ticket content, can get from contract abi
	c.MustStoreAddr(address)
	c.MustStoreUInt(amount, 64)
	c.MustStoreUInt(nonce, 64)
	c.MustStoreUInt(expire, 64)
	c.MustStoreUInt(payout_id, 64)
	return c.EndCell(), nil
}

func createTonWithdrawTicketCell(address *address.Address, amount uint64, nonce, expire, payout_id uint64, signature []byte) (*cell.Cell, error) {
	if len(signature) != 64 {
		return nil, fmt.Errorf("signature length must be 64")
	}
	c := cell.BeginCell()
	c.MustStoreUInt(838693908, 32) // 838693908 is prefix for withdraw ticket, can get from contract abi
	signatureCell := cell.BeginCell().MustStoreSlice(signature, 64*8).EndCell()
	c.MustStoreUInt(1763882716, 32) // 1763882716 is prefix for withdraw ticket content, can get from contract abi
	c.MustStoreAddr(address)
	c.MustStoreUInt(amount, 64)
	c.MustStoreUInt(nonce, 64)
	c.MustStoreUInt(expire, 64)
	c.MustStoreUInt(payout_id, 64)
	return c.MustStoreRef(signatureCell).EndCell(), nil
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

func signTonTicketCell() (string, error) {
	addr := address.MustParseAddr("0QBeE6QYHVrtjrCemwUYnGYZFVM9MkJXSwsURWZF13FkcSMn")
	amount := uint64(1_000_000_000 / 100) // 0.01 TON
	// oneWeekLater := uint64(time.Now().Unix() + 7*24*60*60)
	oneWeekLater := uint64(21234567890)
	nonce := uint64(0)
	payout_id := uint64(1111)
	c, err := createTonWithdrawTicketContentCell(addr, amount, nonce, oneWeekLater, payout_id)
	if err != nil {
		return "", err
	}
	priv, pub, err := signTicketCellByPrivKey()
	if err != nil {
		return "", err
	}

	signature := c.Sign(priv)
	fmt.Printf("signature: %s\n", hex.EncodeToString(signature))
	if !c.Verify(pub, signature) {
		return "", fmt.Errorf("signature verification failed")
	}

	ticketCell, err := createTonWithdrawTicketCell(addr, amount, nonce, oneWeekLater, payout_id, signature)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(ticketCell.ToBOC()), nil
}

func signWithdrawTicketContentCell() ([]byte, error) {
	addr := address.MustParseAddr("EQBeE6QYHVrtjrCemwUYnGYZFVM9MkJXSwsURWZF13FkccVo")
	amount := new(big.Int).SetUint64(1000 * 1_000_000_000).Uint64() // 1000 TON
	oneWeekLater := uint32(2234567890)
	nonce := uint32(0)
	payout_id := uint32(1111)
	user_jetton_wallet := address.MustParseAddr("EQCqm4niLh9ozIEmk4Ub2Avx0Ba0jiGOlc25jdlfUs3zlxof")
	c, err := createWithdrawTicketContentCell(addr, amount, nonce, oneWeekLater, payout_id, user_jetton_wallet)
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

	return signature, nil // ce60f9c3b000d2f30d60eea443cc25862a7d1972c83777741d920fc53c3343ccb83e300f0d8bbe98f9988e624cb0cc89dab9f559eb1e7a1e603f155ef7b9380a
}

func TestSignVerify(t *testing.T) {
	b, err := signTicketCell()
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Println(hex.EncodeToString(b))
}

func TestJettonSignVerify(t *testing.T) {
	b, err := signWithdrawTicketContentCell()
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Println(hex.EncodeToString(b))
}

func TestTonSignVerify(t *testing.T) {
	b, err := signTonTicketCell()
	// ts: te6cckEBAgEAjwABkzH9dBRpIrbcgAvCdIMDq12x1hPTYKMTjMMiqmemSErpYWKIrMi67iyOIAAAAAATEtAAAAAAAAAAAAAAAACeNblaQAAAAAAAAIrwAQCARuk/R+tz09BHWlNPtdj3VJ/B+kHGM4+8tY52/mMNb4xrgeu09k9qBI8KIdBhEDLFDgAr1OCBpscITXp8kfkUAcqJbfc=
	// go: te6cckEBAgEAjwABkzH9dBRpIrbcgAvCdIMDq12x1hPTYKMTjMMiqmemSErpYWKIrMi67iyOIAAAAAATEtAAAAAAAAAAAAAAAACeNblaQAAAAAAAAIrwAQCARuk/R+tz09BHWlNPtdj3VJ/B+kHGM4+8tY52/mMNb4xrgeu09k9qBI8KIdBhEDLFDgAr1OCBpscITXp8kfkUAcqJbfc=
	fmt.Printf("b: %s\n", b)
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Println(b)
}
