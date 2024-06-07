package cell

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/xssnick/tonutils-go/address"
)

var data1024, _ = hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000003")

func TestCell(t *testing.T) {
	c := BeginCell()

	bs := []byte{11, 22, 33}

	err := c.StoreUInt(1, 1)
	if err != nil {
		t.Fatal(err)
		return
	}

	err = c.StoreSlice(bs, 24)
	if err != nil {
		t.Fatal(err)
		return
	}

	amount := uint64(777)
	c2 := BeginCell().MustStoreCoins(amount).EndCell()

	err = c.StoreRef(c2)
	if err != nil {
		t.Fatal(err)
		return
	}

	u38val := uint64(0xAABBCCF)

	err = c.StoreUInt(u38val, 40)
	if err != nil {
		t.Fatal(err)
		return
	}

	boc := c.EndCell().ToBOC()

	cl, err := FromBOC(boc)
	if err != nil {
		t.Fatal(err)
		return
	}

	lc := cl.BeginParse()

	i, err := lc.LoadUInt(1)
	if err != nil {
		t.Fatal(err)
		return
	}

	if i != 1 {
		t.Fatal("1 bit not eq 1")
		return
	}

	bl, err := lc.LoadSlice(24)
	if err != nil {
		t.Fatal(err)
		return
	}

	if !bytes.Equal(bs, bl) {
		t.Fatal("slices not eq:\n" + hex.EncodeToString(bs) + "\n" + hex.EncodeToString(bl))
		return
	}

	u38, err := lc.LoadUInt(40)
	if err != nil {
		t.Fatal(err)
		return
	}

	if u38 != u38val {
		t.Fatal("uint38 not eq")
		return
	}

	ref, err := lc.LoadRef()
	if err != nil {
		t.Fatal(err)
		return
	}

	amt := ref.MustLoadBigCoins()
	if amt.Uint64() != amount {
		t.Fatal("coins ref not eq")
		return
	}
}

func TestFairMintEventCell(t *testing.T) {
	hexStr := "b5ee9c7201010101002c000053814d46f7800bc2748303ab5db1d613d360a3138cc322aa67a6484ae9616288acc8baee2c8e2877359401"
	boc := common.Hex2Bytes(hexStr)
	cl, err := FromBOC(boc)
	if err != nil {
		t.Fatal(err)
		return
	}

	lc := cl.BeginParse()
	i, err := lc.LoadUInt(32)
	if err != nil {
		t.Fatal(err)
		return
	}

	if i != 2169325303 {
		t.Fatal("32 bit not eq 2169325303")
		return
	}

	addr2, err := lc.LoadAddr()
	if err != nil {
		t.Fatal(err)
		return
	}

	fmt.Printf("addr: %s\n", addr2.String())
	amt, err := lc.LoadCoins()
	if err != nil {
		t.Fatal(err)
		return
	}

	if amt != 1000000000 {
		t.Fatal("amount not eq")
		return
	}
}

func TestRefererSetEvent(t *testing.T) {
	// Referrer parent contract address: kQD3v3OY94eF5W9is51HSlUh7BDGH8K771G0s7bl5l8jQMnG
	hexStr := "b5ee9c7201010101004900008da351cba2800bc2748303ab5db1d613d360a3138cc322aa67a6484ae9616288acc8baee2c8e3003ed900ba6d49c45c615e8d2000d161596a6807c6ae2914a18b69c15a3017022ae"
	boc := common.Hex2Bytes(hexStr)
	cl, err := FromBOC(boc)
	if err != nil {
		t.Fatal(err)
		return
	}

	lc := cl.BeginParse()
	i, err := lc.LoadUInt(32)
	fmt.Printf("prefix: %d\n", i)
	if err != nil {
		t.Fatal(err)
		return
	}

	if i != 2740046754 {
		t.Fatal("32 bit not eq 2740046754")
		return
	}

	owner, err := lc.LoadAddr()
	if err != nil {
		t.Fatal(err)
		return
	}

	fmt.Printf("addr: %s\n", owner.String())
	// refer to: https://github.com/ton-blockchain/TEPs/blob/master/text/0002-address.md
	// Mainnet bounceable:  EQBeE6QYHVrtjrCemwUYnGYZFVM9MkJXSwsURWZF13FkccVo
	// Mainnet non-bounceable:  UQBeE6QYHVrtjrCemwUYnGYZFVM9MkJXSwsURWZF13FkcZit
	// Testnet bounceable:  kQBeE6QYHVrtjrCemwUYnGYZFVM9MkJXSwsURWZF13FkcX7i
	// Testnet non-bounceable:  0QBeE6QYHVrtjrCemwUYnGYZFVM9MkJXSwsURWZF13FkcSMn
	if owner.String() != "EQBeE6QYHVrtjrCemwUYnGYZFVM9MkJXSwsURWZF13FkccVo" &&
		owner.String() != "UQBeE6QYHVrtjrCemwUYnGYZFVM9MkJXSwsURWZF13FkcZit" &&
		owner.String() != "kQBeE6QYHVrtjrCemwUYnGYZFVM9MkJXSwsURWZF13FkcX7i" &&
		owner.String() != "0QBeE6QYHVrtjrCemwUYnGYZFVM9MkJXSwsURWZF13FkcSMn" {
		t.Fatal("addr not eq")
		return
	}
	referrer, err := lc.LoadAddr()
	fmt.Printf("addr: %s\n", referrer.String())
	// Mainnet bounceable:  EQD7ZALptScRcYV6NIADRYVlqaAfGrikUoYtpwVowFwIq2uo
	// Mainnet non-bounceable:  UQD7ZALptScRcYV6NIADRYVlqaAfGrikUoYtpwVowFwIqzZt
	// Testnet bounceable:  kQD7ZALptScRcYV6NIADRYVlqaAfGrikUoYtpwVowFwIq9Ai
	// Testnet non-bounceable:  0QD7ZALptScRcYV6NIADRYVlqaAfGrikUoYtpwVowFwIq43n
	if referrer.String() != "EQD7ZALptScRcYV6NIADRYVlqaAfGrikUoYtpwVowFwIq2uo" &&
		referrer.String() != "UQD7ZALptScRcYV6NIADRYVlqaAfGrikUoYtpwVowFwIqzZt" &&
		referrer.String() != "kQD7ZALptScRcYV6NIADRYVlqaAfGrikUoYtpwVowFwIq9Ai" &&
		referrer.String() != "0QD7ZALptScRcYV6NIADRYVlqaAfGrikUoYtpwVowFwIq43n" {
		t.Fatal("referrer not eq")
		return
	}
	if err != nil {
		t.Fatal(err)
		return
	}
}

type Coins uint64

type TradeInfo struct {
	Referrer         *address.Address
	ReferrerAmount   Coins
	UpReferrer       *address.Address
	UpReferrerAmount Coins
	FeeValue         Coins
	RemainAmount     Coins
}

func LoadTradeInfo(slice *Slice) TradeInfo {
	sc0 := slice
	referrer := sc0.MustLoadAddr()
	referrerAmount := sc0.MustLoadCoins()
	upReferrer := sc0.MustLoadAddr()
	upReferrerAmount := sc0.MustLoadCoins()
	feeValue := sc0.MustLoadCoins()

	sc1, err := sc0.LoadRef()
	if err != nil {
		fmt.Println(err)
		return TradeInfo{}
	}
	remainAmount := sc1.MustLoadCoins()
	return TradeInfo{
		Referrer:         referrer,
		ReferrerAmount:   Coins(referrerAmount),
		UpReferrer:       upReferrer,
		UpReferrerAmount: Coins(upReferrerAmount),
		FeeValue:         Coins(feeValue),
		RemainAmount:     Coins(remainAmount),
	}
}

func TestBuyEvent(t *testing.T) {
	// jetton master(main) 	contract address: kQADonXe1GRJ7UX3hl3Ql80XARitTxrStrywwZsDGF9v3e1U
	// txid: https://testnet.tonviewer.com/transaction/2a39a484a1d4c8eacccaf33db87f4d16464b5e4efced8dd5d10cae75221e69ff
	hexStr := "b5ee9c7201010301008c000165a2328a25800bc2748303ab5db1d613d360a3138cc322aa67a6484ae9616288acc8baee2c8e2802625a00e72c36e2e359a74203010197800bc2748303ab5db1d613d360a3138cc322aa67a6484ae9616288acc8baee2c8e25d4c1001784e9060756bb63ac27a6c1462719864554cf4c9095d2c2c511599175dc591c4a7100c0c350200200094012c99208"
	boc := common.Hex2Bytes(hexStr)
	cl, err := FromBOC(boc)
	if err != nil {
		t.Fatal(err)
		return
	}
	lc := cl.BeginParse()
	i, err := lc.LoadUInt(32)
	if err != nil {
		t.Fatal(err)
		return
	}

	fmt.Printf("prefix: %d\n", i)
	if i != 2721221157 {
		t.Fatal("32 bit not eq 2721221157")
		return
	}

	addr2, err := lc.LoadAddr()
	if err != nil {
		t.Fatal(err)
		return
	}

	fmt.Printf("addr: %s\n", addr2.String())
	amt, err := lc.LoadCoins()
	if err != nil {
		t.Fatal(err)
		return
	}

	if amt != 2*1e7 {
		t.Fatal("amount not eq")
		return
	}

	tokenAmount, err := lc.LoadCoins()
	fmt.Printf("token amount: %d\n", tokenAmount)
	if err != nil {
		t.Fatal(err)
		return
	}
	lastTokenPrice, err := lc.LoadCoins()
	fmt.Printf("last token price: %d\n", lastTokenPrice)
	if err != nil {
		t.Fatal(err)
		return
	}

	tradeInfoRef, err := lc.LoadRef()
	if err != nil {
		t.Fatal(err)
		return
	}
	tradeSlice := tradeInfoRef.MustToCell().BeginParse()
	tradeInfo := LoadTradeInfo(tradeSlice)
	fmt.Printf("referrer: %s, referrer amount: %d, up referrer: %s, up referrer amount: %d, fee value: %d, remain amount: %d\n", tradeInfo.Referrer.String(), tradeInfo.ReferrerAmount, tradeInfo.UpReferrer.String(), tradeInfo.UpReferrerAmount, tradeInfo.FeeValue, tradeInfo.RemainAmount)
}

func TestSellEvent(t *testing.T) {
	// jetton master(main) 	contract address: kQADonXe1GRJ7UX3hl3Ql80XARitTxrStrywwZsDGF9v3e1U
	// txid: https://testnet.tonviewer.com/transaction/ca39f790cb8bef2670264f14748119f57a4ae62a38e05de8ff4ebf3afbe79fa8
	hexStr := "b5ee9c7201010301008a00016355cbd8fb800bc2748303ab5db1d613d360a3138cc322aa67a6484ae9616288acc8baee2c8e2770e8b0e470de4df820000203010197800bc2748303ab5db1d613d360a3138cc322aa67a6484ae9616288acc8baee2c8e251fa3001784e9060756bb63ac27a6c1462719864554cf4c9095d2c2c511599175dc591c497f80c077d9200200073b874588"
	boc := common.Hex2Bytes(hexStr)
	cl, err := FromBOC(boc)
	if err != nil {
		t.Fatal(err)
		return
	}
	lc := cl.BeginParse()
	i, err := lc.LoadUInt(32)
	if err != nil {
		t.Fatal(err)
		return
	}

	fmt.Printf("prefix: %d\n", i)
	if i != 1439422715 {
		t.Fatal("32 bit not eq 1439422715")
		return
	}

	addr2, err := lc.LoadAddr()
	if err != nil {
		t.Fatal(err)
		return
	}

	fmt.Printf("addr: %s\n", addr2.String())
	amt, err := lc.LoadCoins()
	if err != nil {
		t.Fatal(err)
		return
	}

	fmt.Printf("amount: %d\n", amt)
	if amt != 12088408 {
		t.Fatal("amount not eq")
		return
	}

	tokenAmount, err := lc.LoadCoins()
	fmt.Printf("token amount: %d\n", tokenAmount)
	if err != nil {
		t.Fatal(err)
		return
	}
	lastTokenPrice, err := lc.LoadCoins()
	fmt.Printf("last token price: %d\n", lastTokenPrice)
	if err != nil {
		t.Fatal(err)
		return
	}

	tradeInfoRef, err := lc.LoadRef()
	if err != nil {
		t.Fatal(err)
		return
	}
	tradeSlice := tradeInfoRef.MustToCell().BeginParse()
	tradeInfo := LoadTradeInfo(tradeSlice)
	fmt.Printf("referrer: %s, referrer amount: %d, up referrer: %s, up referrer amount: %d, fee value: %d, remain amount: %d\n", tradeInfo.Referrer.String(), tradeInfo.ReferrerAmount, tradeInfo.UpReferrer.String(), tradeInfo.UpReferrerAmount, tradeInfo.FeeValue, tradeInfo.RemainAmount)
}

func parseTransferNotifiy(hexStr string) {
	fmt.Println("----------------------")

	boc := common.Hex2Bytes(hexStr)
	cl, err := FromBOC(boc)
	if err != nil {
		fmt.Printf("fromBOC error: %s\n", err)
		return
	}
	lc := cl.BeginParse()

	prefix, err := lc.LoadUInt(32)
	if err != nil {
		fmt.Printf("load prefix error: %s\n", err)
		return
	}
	fmt.Print("prefix: ", prefix, "\n")

	i, err := lc.LoadUInt(64)
	fmt.Print("query_id: ", i, "\n")
	if err != nil {
		fmt.Printf("load uint64 error: %s\n", err)
		return
	}

	amount, err := lc.LoadCoins()
	if err != nil {
		fmt.Printf("load coins error: %s\n", err)
		return
	}

	fmt.Printf("amount: %d\n", amount)
	addr, err := lc.LoadAddr()
	if err != nil {
		fmt.Printf("load addr error: %s\n", err)
		return
	}
	fmt.Printf("addr: %s\n", addr)

	forwardPayload, err := lc.LoadRef()
	if err != nil {
		fmt.Printf("load ref cell error: %s\n", err)
		return
	}
	payload := forwardPayload.MustToCell().BeginParse()
	sc0 := payload
	prefix2 := sc0.MustLoadUInt(32)
	fmt.Printf("prefix2: %d\n", prefix2)
	token_walelt := sc0.MustLoadAddr()
	fmt.Printf("token_walelt: %s\n", token_walelt.String())
	minLPOut := sc0.MustLoadCoins()
	fmt.Printf("minLPOut: %d\n", minLPOut)
}

//	message(0x7362d09c) TokenNotification {
//		query_id: Int as uint64;
//		amount: Int as coins;
//		from: Address;
//		forward_payload: Slice as remaining;
//	}
func TestTransferNotification(t *testing.T) {
	successed := "b5ee9c7201010201005e0001647362d09c001c3742bd829d18405f767a0801f6c805d36a4e22e30af46900068b0acb53403e357148a50c5b4e0ad180b8115701004dfcf9e58f80022a16a3164c4d5aa3133f3110ff10496e00ca8ac8abeffc5027e024d33480c3e203"
	parseTransferNotifiy(successed)
	fmt.Println("----------------------")
	fmt.Printf("%s\n", successed)
}

func TestPayoutCompleteEvent(t *testing.T) {
	// Payouts contract address: kQAiL_EralRCIRBPR4o4uF25x87AVcBBqA_oEf9qawDXOTVc
	// payout0 txid: b0b38406e7f16db8d02347c132f7aff82f990c993fa5b07edebc93e7bfd52c06
	// payout1 txid: 0041cb5d41ca0f23c0bedbc384fc4ee7c49a89b085e3cb188f2e37d494d3b826
	// payout2 txid: 1a77ceacc3f2d4e03092e10d020d8e217d9ad168d3530eb569b22d2edcbaed01
	hexStr := "b5ee9c7201010101003b000071d1996541800bc2748303ab5db1d613d360a3138cc322aa67a6484ae9616288acc8baee2c8e27312d00000000000000000000000000000008af"
	// hexStr := "b5ee9c7201010101003b000071d1996541800bc2748303ab5db1d613d360a3138cc322aa67a6484ae9616288acc8baee2c8e27312d000000000000000002000000000000115d"
	// hexStr := "b5ee9c7201010101003b000071d1996541800bc2748303ab5db1d613d360a3138cc322aa67a6484ae9616288acc8baee2c8e27312d0000000000000000040000000000001a0b"
	boc := common.Hex2Bytes(hexStr)
	cl, err := FromBOC(boc)
	if err != nil {
		t.Fatal(err)
		return
	}
	lc := cl.BeginParse()
	i, err := lc.LoadUInt(32)
	if err != nil {
		t.Fatal(err)
		return
	}

	fmt.Printf("prefix: %d\n", i)
	if i != 3516491073 {
		t.Fatal("32 bit not eq 3516491073")
		return
	}

	addr2, err := lc.LoadAddr()
	if err != nil {
		t.Fatal(err)
		return
	}

	fmt.Printf("addr: %s\n", addr2.String())
	value, err := lc.LoadCoins()
	if err != nil {
		t.Fatal(err)
		return
	}

	fmt.Printf("value: %d\n", value)
	nonce, err := lc.LoadUInt(64)
	fmt.Printf("nonce: %d\n", nonce)
	if err != nil {
		t.Fatal(err)
		return
	}
	payout_id, err := lc.LoadUInt(64)
	fmt.Printf("payout_id: %d\n", payout_id)
	if err != nil {
		t.Fatal(err)
		return
	}
}

func TestJettonPayoutCompleteEvent(t *testing.T) {
	// txid: https://testnet.tonviewer.com/transaction/ed86304d49d77f19ce98428034390f963b9d0253603cc4cdddf04803c0f91cde
	hexStr := "b5ee9c720101010100590000ad1df4b988800bc2748303ab5db1d613d360a3138cc322aa67a6484ae9616288acc8baee2c8e2000001d1a94a200000000000000008af002aa6e2788b87da332049a4e146f602fc7405ad238863a5736e637657d4b37ce5e"
	boc := common.Hex2Bytes(hexStr)
	cl, err := FromBOC(boc)
	if err != nil {
		t.Fatal(err)
		return
	}
	lc := cl.BeginParse()
	i, err := lc.LoadUInt(32)
	if err != nil {
		t.Fatal(err)
		return
	}

	fmt.Printf("prefix: %d\n", i)
	if i != 502577544 {
		t.Fatal("32 bit not eq 502577544")
		return
	}

	addr2, err := lc.LoadAddr()
	if err != nil {
		t.Fatal(err)
		return
	}

	fmt.Printf("addr: %s\n", addr2.String())
	value, err := lc.LoadUInt(64)
	if err != nil {
		t.Fatal(err)
		return
	}

	fmt.Printf("value: %d\n", value)
	nonce, err := lc.LoadUInt(32)
	fmt.Printf("nonce: %d\n", nonce)
	if err != nil {
		t.Fatal(err)
		return
	}
	payout_id, err := lc.LoadUInt(32)
	fmt.Printf("payout_id: %d\n", payout_id)
	if err != nil {
		t.Fatal(err)
		return
	}
	user_jetton_wallet, err := lc.LoadAddr()
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Printf("user_jetton_wallet: %s\n", user_jetton_wallet.String())
}

func TestCell24(t *testing.T) {
	c := BeginCell()

	bs := []byte{11, 22, 33}

	err := c.StoreSlice(bs, 24)
	if err != nil {
		t.Fatal(err)
		return
	}

	lc := c.EndCell().BeginParse()

	res, err := lc.LoadSlice(24)
	if err != nil {
		t.Fatal(err)
		return
	}

	if !bytes.Equal(bs, res) {
		t.Fatal("slices not eq:\n" + hex.EncodeToString(bs) + "\n" + hex.EncodeToString(res))
		return
	}
}

func TestCell25(t *testing.T) {
	c := BeginCell()

	bs := []byte{11, 22, 33, 0x80}

	err := c.StoreSlice(bs, 25)
	if err != nil {
		t.Fatal(err)
		return
	}

	lc := c.EndCell().BeginParse()

	res, err := lc.LoadSlice(25)
	if err != nil {
		t.Fatal(err)
		return
	}

	if !bytes.Equal(bs, res) {
		t.Fatal("slices not eq:\n" + hex.EncodeToString(bs) + "\n" + hex.EncodeToString(res))
		return
	}
}

func TestCellReadSmall(t *testing.T) {
	c := BeginCell()

	bs := []byte{0b10101010, 0x00, 0x00}

	err := c.StoreSlice(bs, 24)
	if err != nil {
		t.Fatal(err)
		return
	}

	lc := c.EndCell().BeginParse()

	for i := 0; i < 8; i++ {
		res, err := lc.LoadUInt(1)
		if err != nil {
			t.Fatal(err)
			return
		}

		if (res != 1 && i%2 == 0) || (res != 0 && i%2 == 1) {
			t.Fatal("not eq " + fmt.Sprint(i*2))
			return
		}
	}

	res, err := lc.LoadUInt(1)
	if err != nil {
		t.Fatal(err)
		return
	}

	if res != 0 {
		t.Fatal("not 0")
		return
	}
}

func TestCellReadEmpty(t *testing.T) {
	c := BeginCell().EndCell().BeginParse()
	sz, _, err := c.RestBits()
	if err != nil {
		t.Fatal(err)
		return
	}

	if sz != 0 {
		t.Fatal("not 0")
		return
	}
}

func TestBuilder_MustStoreUInt(t *testing.T) {
	val := BeginCell().MustStoreUInt(516783, 23).EndCell().BeginParse().MustLoadUInt(23)
	if val != 516783 {
		t.Fatal("incorrect", val)
	}

	val = BeginCell().MustStoreUInt(2, 64).EndCell().BeginParse().MustLoadUInt(64)
	if val != 2 {
		t.Fatal("incorrect2", val)
	}

	val = BeginCell().MustStoreUInt(0xFFFFFF, 24).EndCell().BeginParse().MustLoadUInt(24)
	if val != 0xFFFFFF {
		t.Fatal("incorrect3", val)
	}

	val = BeginCell().MustStoreUInt(0xFFFFFF, 24).EndCell().BeginParse().MustLoadUInt(20)
	if val != 0xFFFFF {
		t.Fatal("incorrect4", val)
	}

	val = BeginCell().MustStoreUInt(2, 2).EndCell().BeginParse().MustLoadUInt(2)
	if val != 2 {
		t.Fatal("incorrect5", val)
	}

	val = BeginCell().MustStoreUInt(1, 1).EndCell().BeginParse().MustLoadUInt(1)
	if val != 1 {
		t.Fatal("incorrect6", val)
	}

	val = BeginCell().MustStoreUInt(123456789, 70).EndCell().BeginParse().MustLoadUInt(70)
	if val != 123456789 {
		t.Fatal("incorrect7", val)
	}

	val = BeginCell().MustStoreUInt(0xFFFFFFFFFFFFFFFF, 60).EndCell().BeginParse().MustLoadUInt(60)
	if val != 0xFFFFFFFFFFFFFFF {
		t.Fatal("incorrect8", val)
	}
}

func TestBuilder_StoreBigInt(t *testing.T) {
	c := BeginCell()

	err := c.StoreBigInt(new(big.Int), 300)
	if err != ErrTooBigSize {
		t.Fatal("err incorrect, its:", err)
	}

	err = c.StoreBigInt(new(big.Int).Lsh(big.NewInt(1), 257), 256)
	if err != ErrTooBigValue {
		t.Fatal("err incorrect, its:", err)
	}

	c.MustStoreBigInt(new(big.Int).SetInt64(-3), 256)

	data := hex.EncodeToString(c.EndCell().BeginParse().MustLoadSlice(256))
	if data != "fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffd" {
		t.Fatal("value incorrect, its:", data)
	}
}

func TestBuilder_StoreBigUInt(t *testing.T) {
	c := BeginCell()

	err := c.StoreBigUInt(new(big.Int), 300)
	if err != ErrTooBigSize {
		t.Fatal("err incorrect, its:", err)
	}

	err = c.StoreBigUInt(new(big.Int).Lsh(big.NewInt(1), 257), 256)
	if err != ErrTooBigValue {
		t.Fatal("err incorrect, its:", err)
	}

	err = c.StoreBigUInt(big.NewInt(-1), 256)
	if err != ErrNegative {
		t.Fatal("err incorrect, its:", err)
	}

	c.MustStoreBigUInt(new(big.Int).SetInt64(3), 256)

	data := hex.EncodeToString(c.EndCell().BeginParse().MustLoadSlice(256))
	if data != "0000000000000000000000000000000000000000000000000000000000000003" {
		t.Fatal("value incorrect, its:", data)
	}
}

func TestBuilder_StoreSlice(t *testing.T) {
	c := BeginCell()

	err := c.StoreSlice([]byte{}, 1023)
	if err != ErrSmallSlice {
		t.Fatal("err incorrect, its:", err)
	}

	err = c.StoreSlice(data1024, 1024)
	if err != ErrNotFit1023 {
		t.Fatal("err incorrect, its:", err)
	}

	err = c.StoreSlice(data1024, 1023)
	if err != nil {
		t.Fatal("err incorrect, its:", err)
	}
}

func TestBuilder_StoreRef(t *testing.T) {
	c := BeginCell()

	err := c.StoreRef(nil)
	if err != ErrRefCannotBeNil {
		t.Fatal("err incorrect, its:", err)
	}

	for i := 0; i < 4; i++ {
		err = c.StoreRef(BeginCell().EndCell())
		if err != nil {
			t.Fatal("err incorrect, its:", err)
		}
	}

	err = c.StoreRef(BeginCell().EndCell())
	if err != ErrTooMuchRefs {
		t.Fatal("err incorrect, its:", err)
	}
}

func TestBuilder_VarUint(t *testing.T) {
	for i := uint(3); i <= 18; i++ {
		c := BeginCell().MustStoreVarUInt(777, i).EndCell()
		if c.BeginParse().MustLoadVarUInt(i).Uint64() != 777 {
			t.Fatal("var uint not eq")
		}
	}
}

func TestBuilder_StoreBuilder(t *testing.T) {
	c := BeginCell().MustStoreSlice(data1024, 1015).MustStoreRef(BeginCell().EndCell())
	b1bad := BeginCell().MustStoreSlice([]byte{0xAA, 0xBB}, 16).MustStoreRef(BeginCell().EndCell())
	b2bad := BeginCell().MustStoreSlice([]byte{0xAA}, 8).MustStoreRef(BeginCell().EndCell()).MustStoreRef(BeginCell().EndCell()).MustStoreRef(BeginCell().EndCell()).MustStoreRef(BeginCell().EndCell())
	b3 := BeginCell().MustStoreSlice([]byte{0xAA}, 8).MustStoreRef(BeginCell().EndCell()).MustStoreRef(BeginCell().EndCell()).MustStoreRef(BeginCell().EndCell())

	err := c.StoreBuilder(b1bad)
	if err != ErrNotFit1023 {
		t.Fatal("err incorrect, its:", err)
	}

	err = c.StoreBuilder(b2bad)
	if err != ErrTooMuchRefs {
		t.Fatal("err incorrect, its:", err)
	}

	err = c.StoreBuilder(b3)
	if err != nil {
		t.Fatal("err incorrect, its:", err)
	}

	if val := c.RefsLeft(); val != 0 {
		t.Fatal("refs left incorrect, its:", val)
	}

	if val := c.BitsLeft(); val != 0 {
		t.Fatal("bits left incorrect, its:", val)
	}

	if val := c.BitsUsed(); val != 1023 {
		t.Fatal("bits used incorrect, its:", val)
	}

	if val := c.RefsUsed(); val != 4 {
		t.Fatal("refs used incorrect, its:", val)
	}
}

func TestSliceFuzz(t *testing.T) {
	arr1 := make([]byte, 128)
	arr2 := make([]byte, 128)

	for i := 0; i < 500000; i++ {
		sz1 := uint(int(arr1[0]*arr1[1]) % 512)
		sz2 := uint(int(arr2[0]*arr2[1]) % 512)
		rand.Read(arr1)
		rand.Read(arr2)

		c := BeginCell()

		if err := c.StoreSlice(arr1, sz1); err != nil {
			t.Fatal(err)
		}

		if err := c.StoreSlice(arr2, sz2); err != nil {
			t.Fatal(err)
		}

		s := c.EndCell().BeginParse()
		data1 := s.MustLoadSlice(sz1)
		data2 := s.MustLoadSlice(sz2)

		oneMore := uint(0)
		if sz1%8 != 0 {
			oneMore = 1
		}
		cut1 := arr1[:sz1/8+oneMore]
		if oneMore > 0 {
			cut1[len(cut1)-1] &= 0xFF << (8 - (sz1 % 8))
		}
		if !bytes.Equal(data1, cut1) {
			t.Fatal("data1 not eq after load")
		}

		oneMore = uint(0)
		if sz2%8 != 0 {
			oneMore = 1
		}
		cut2 := arr2[:sz2/8+oneMore]
		if oneMore > 0 {
			cut2[len(cut2)-1] &= 0xFF << (8 - (sz2 % 8))
		}
		if !bytes.Equal(data2, cut2) {
			t.Fatal("data2 not eq after load")
		}
	}
}
