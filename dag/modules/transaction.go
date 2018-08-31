/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/
/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package modules

import (
	"fmt"
	"math/big"
	"strconv"
	"time"
	"bytes"
	//	"fmt"
	"io"
    "encoding/binary"
	"math"
	//
	"github.com/palletone/go-palletone/common"
   // "github.com/palletone/go-palletone/tokenengine/btcd/wire"
	//	"github.com/palletone/go-palletone/common/crypto"
	//	"github.com/palletone/go-palletone/common/crypto/sha3"
	//	"github.com/palletone/go-palletone/common/hexutil"
	//  "github.com/Re-volution/sizestruct"
	"github.com/palletone/go-palletone/common/rlp"
)

var (
	TXFEE      = big.NewInt(5000000) // transaction fee =5ptn
	TX_MAXSIZE = uint32(256 * 1024)
)

type Asset struct {
	AssertId IDType16 `json:"assert_id"` // 资产类别
	UniqueId IDType16 `json:"unique_id"` // every token has its unique id
	ChainId  uint64   `json:"chain_id"`  // main chain id or sub-chain id
}
// TxOut defines a bitcoin transaction output.
type TxOut struct {
	Value    int64
	PkScript []byte
    Asset    Asset
}
// TxIn defines a bitcoin transaction input.
type TxIn struct {
	PreviousOutPoint OutPoint
	SignatureScript  []byte
	Sequence         uint32
}

func NewTransaction(msg []Message, lock uint32) *Transaction {
	return newTransaction(msg, lock)
}

func NewContractCreation(msg []Message, lock uint32) *Transaction {
	return newTransaction(msg, lock)
}

func newTransaction(msg []Message, lock uint32) *Transaction {
	tx := new(Transaction)
	tx.TxMessages = msg[:]
	//tx.Locktime = lock

	return tx
}
// AddTxIn adds a transaction input to the message.
func (pld *PaymentPayload) AddTxIn(ti Input) {
	pld.Input = append(pld.Input, &ti)
}
// AddTxOut adds a transaction output to the message.
func (pld *PaymentPayload) AddTxOut(to Output) {
	pld.Output = append(pld.Output, &to)
}

func (t *Transaction) SetHash(hash common.Hash) {
	if t.TxHash == (common.Hash{}) {
		t.TxHash = hash
	} else {
		t.TxHash.Set(hash)
	}
}

func NewPaymentPayload() *PaymentPayload {
	return &PaymentPayload{
		Input:    make([]*Input, 0, defaultTxInOutAlloc),
		Output:   make([]*Output, 0, defaultTxInOutAlloc),
	}
}
type TxPoolTransaction struct {
	Transaction

	CreationDate string  `json:"creation_date" rlp:"-"`
	Priority_lvl float64 `json:"priority_lvl" rlp:"-"` // 打包的优先级
	Nonce        uint64  // transaction'hash maybe repeat.
	Pending      bool
	Confirmed    bool
	Extra        []byte
}

//// ChainId returns which chain id this transaction was signed for (if at all)
//func (tx Transaction) ChainId() *big.Int {
//	return deriveChainId(tx.data.V)
//}
//
//// Protected returns whether the transaction is protected from replay protection.
//func (tx Transaction) Protected() bool {
//	return isProtectedV(tx.data.V)
//}
//
//func isProtectedV(V *big.Int) bool {
//	if V.BitLen() <= 8 {
//		v := V.Uint64()
//		return v != 27 && v != 28
//	}
//	// anything not 27 or 28 are considered unprotected
//	return true
//}
//
//// EncodeRLP implements rlp.Encoder
//func (tx *Transaction) EncodeRLP(w io.Writer) error {
//	return rlp.Encode(w, &tx.data)
//}
//
//// DecodeRLP implements rlp.Decoder
//func (tx *Transaction) DecodeRLP(s *rlp.Stream) error {
//	_, UnitSize, _ := s.Kind()
//	err := s.Decode(&tx.data)
//	if err == nil {
//		tx.UnitSize.Store(common.StorageSize(rlp.ListSize(UnitSize)))
//	}
//
//	return err
//}
//
//// MarshalJSON encodes the web3 RPC transaction format.
//func (tx *Transaction) MarshalJSON() ([]byte, error) {
//	UnitHash := tx.Hash()
//	data := tx.data
//	data.Hash = &UnitHash
//	return data.MarshalJSON()
//}
//
//// UnmarshalJSON decodes the web3 RPC transaction format.
//func (tx *Transaction) UnmarshalJSON(input []byte) error {
//	var dec txdata
//	if err := dec.UnmarshalJSON(input); err != nil {
//		return err
//	}
//	var V byte
//	if isProtectedV(dec.V) {
//		chainID := deriveChainId(dec.V).Uint64()
//		V = byte(dec.V.Uint64() - 35 - 2*chainID)
//	} else {
//		V = byte(dec.V.Uint64() - 27)
//	}
//	if !crypto.ValidateSignatureValues(V, dec.R, dec.S, false) {
//		return errors.New("invalid transaction v, r, s values")
//	}
//	*tx = Transaction{data: dec}
//	return nil
//}
//
//func (tx Transaction) Data() []byte { return common.CopyBytes(tx.data.Payload) }
//

func (tx *TxPoolTransaction) GetPriorityLvl() float64 {
	// priority_lvl=  fee/size*(1+(time.Now-CreationDate)/24)
	var priority_lvl float64
	if txfee := tx.Fee(); txfee.Int64() > 0 {
		t0, _ := time.Parse(TimeFormatString, tx.CreationDate)
		priority_lvl, _ = strconv.ParseFloat(fmt.Sprintf("%f", float64(txfee.Int64())/tx.Size().Float64()*(1+float64(time.Now().Hour()-t0.Hour())/24)), 64)
	}
	return priority_lvl
}
func (tx *TxPoolTransaction) SetPriorityLvl(priority float64) {
	tx.Priority_lvl = priority
}

// Hash hashes the RLP encoding of tx.
// It uniquely identifies the transaction.
func (tx Transaction) Hash() common.Hash {
	withoutSigTx := Transaction{}
	withoutSigTx.CopyFrTransaction(&tx)
	withoutSigTx.TxHash = common.Hash{}
	v := rlp.RlpHash(withoutSigTx)
	return v
}

// Size returns the true RLP encoded storage UnitSize of the transaction, either by
// encoding and returning it, or returning a previsouly cached value.
func (tx *Transaction) Size() common.StorageSize {
	c := writeCounter(0)
	rlp.Encode(&c, &tx)
	return common.StorageSize(c)
}

func (tx *Transaction) CreateDate() string {
	n := time.Now()
	return n.Format(TimeFormatString)
}

func (tx *Transaction) Fee() *big.Int {
	return TXFEE
}

func (tx *Transaction) Address() common.Address {
	return common.Address{}
}

// Cost returns amount + price
func (tx *Transaction) Cost() *big.Int {
	//if tx.TxFee.Cmp(TXFEE) < 0 {
	//	tx.TxFee = TXFEE
	//}
	//return tx.TxFee
	return TXFEE
}

func (tx *Transaction) CopyFrTransaction(cpy *Transaction) {
	tx.TxHash.Set(cpy.TxHash)
	//tx.Locktime = cpy.Locktime
	tx.TxMessages = make([]Message, len(cpy.TxMessages))
	for _, msg := range cpy.TxMessages {
		newMsg := Message{}
		newMsg = *newMsg.CopyMessages(&msg)
		tx.TxMessages = append(tx.TxMessages, newMsg)
	}
}

//// AsMessage returns the transaction as a core.Message.
////
//// AsMessage requires a signer to derive the sender.
////
//// XXX Rename message to something less arbitrary?
//func (tx *Transaction) AsMessage(s Signer) (Message, error) {
//	msg := Message{
//		from:       *tx.data.From,
//		gasPrice:   new(big.Int).Set(tx.data.Price),
//		to:         tx.data.Recipient,
//		amount:     tx.data.Amount,
//		data:       tx.data.Payload,
//		checkNonce: true,
//	}
//
//	var err error
//	msg.from, err = Sender(s, tx)
//	return msg, err
//}
//

// Len returns the length of s.
func (s Transactions) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s Transactions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// GetRlp implements Rlpable and returns the i'th element of s in rlp.
func (s Transactions) GetRlp(i int) []byte {
	enc, _ := rlp.EncodeToBytes(s[i])
	return enc
}
func (s Transactions) Hash() common.Hash {
	v := rlp.RlpHash(s)
	return v
}

// TxDifference returns a new set t which is the difference between a to b.
func TxDifference(a, b Transactions) (keep Transactions) {
	keep = make(Transactions, 0, len(a))

	remove := make(map[common.Hash]struct{})
	for _, tx := range b {
		remove[tx.Hash()] = struct{}{}
	}

	for _, tx := range a {
		if _, ok := remove[tx.Hash()]; !ok {
			keep = append(keep, tx)
		}
	}

	return keep
}

// single account, otherwise a nonce comparison doesn't make much sense.
type TxByNonce TxPoolTxs

func (s TxByNonce) Len() int           { return len(s) }
func (s TxByNonce) Less(i, j int) bool { return s[i].Nonce < s[j].Nonce }
func (s TxByNonce) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// TxByPrice implements both the sort and the heap interface, making it useful
// for all at once sorting as well as individually adding and removing elements.
type TxByPrice TxPoolTxs

func (s TxByPrice) Len() int      { return len(s) }
func (s TxByPrice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s *TxByPrice) Push(x interface{}) {
	*s = append(*s, x.(*TxPoolTransaction))
}

func (s *TxByPrice) Pop() interface{} {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

type TxByPriority []*TxPoolTransaction

func (s TxByPriority) Len() int           { return len(s) }
func (s TxByPriority) Less(i, j int) bool { return s[i].Priority_lvl > s[j].Priority_lvl }
func (s TxByPriority) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s *TxByPriority) Push(x interface{}) {
	*s = append(*s, x.(*TxPoolTransaction))
}

func (s *TxByPriority) Pop() interface{} {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

//// Message is a fully derived transaction and implements core.Message
////
//// NOTE: In a future PR this will be removed.
//
//func NewMessage(from, to *common.Address, nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, checkNonce bool) Message {
//	return Message{
//		from:       *from,
//		to:         to,
//		nonce:      nonce,
//		amount:     amount,
//		gasLimit:   gasLimit,
//		gasPrice:   gasPrice,
//		data:       data,
//		checkNonce: checkNonce,
//	}
//}
//
//func (m Message) From() *common.Address { return &m.from }
//func (m Message) To() *common.Address   { return m.to }
//func (m Message) GasPrice() *big.Int    { return m.gasPrice }
//func (m Message) Value() *big.Int       { return m.amount }
//func (m Message) Gas() uint64           { return m.gasLimit }
//func (m Message) Nonce() uint64         { return m.nonce }
//func (m Message) Data() []byte          { return m.data }
//func (m Message) CheckNonce() bool      { return m.checkNonce }
//
//// deriveChainId derives the chain id from the given v parameter
//func deriveChainId(v *big.Int) *big.Int {
//	if v.BitLen() <= 64 {
//		v := v.Uint64()
//		if v == 27 || v == 28 {
//			return new(big.Int)
//		}
//		return new(big.Int).SetUint64((v - 35) / 2)
//	}
//	v = new(big.Int).Sub(v, big.NewInt(35))
//	return v.Div(v, big.NewInt(2))
//}
//func rlpHash(x interface{}) (h common.Hash) {
//	hw := sha3.NewKeccak256()
//	rlp.Encode(hw, x)
//	hw.Sum(h[:0])
//	return h
//}
//
//// deriveSigner makes a *best* guess about which signer to use.
//func deriveSigner(V *big.Int) Signer {
//	if V.Sign() != 0 && isProtectedV(V) {
//		return NewEIP155Signer(deriveChainId(V))
//	} else {
//		return HomesteadSigner{}
//	}
//}
//
type writeCounter common.StorageSize

func (c *writeCounter) Write(b []byte) (int, error) {
	*c += writeCounter(len(b))
	return len(b), nil
}

var (
	EmptyRootHash = DeriveSha(Transactions{})
)

type TxLookupEntry struct {
	UnitHash  common.Hash
	UnitIndex uint64
	Index     uint64
}
type Transactions []*Transaction
type Transaction struct {
	TxHash     common.Hash `json:"txhash"`
	TxMessages []Message   `json:"messages"`
	// todo AccountNonce, CreationDate, Priority_lvl 在交易池部分用的比较多，将由杨杰负责删除
	AccountNonce uint64  `json:"account_nonce" rlp:"-"`
	CreationDate string  `json:"creation_date" rlp:"-"`
	Priority_lvl float64 `json:"priority_lvl" rlp:"-"` // 打包的优先级
}
type OutPoint struct {
 TxHash       common.Hash // reference Utxo struct key field
 MessageIndex uint32      // message index in transaction
 OutIndex     uint32
}
func NewOutPoint(hash *common.Hash, messageindex uint32,outindex uint32) *OutPoint {
	return &OutPoint{
		TxHash:  *hash,
		MessageIndex: messageindex,
		OutIndex: outindex,
	}
}
// key: message.UnitHash(message+timestamp)
type Message struct {
	App     string      `json:"app"`     // message type
	Payload interface{} `json:"payload"` // the true transaction data
}
/************************** Payload Details ******************************************/
type PayloadMapStruct struct {
	Key   string
	Value interface{}
}
// Token exchange message and verify message
// App: payment
type PaymentPayload struct {
	Input  []*Input  `json:"inputs"`
	Output []*Output `json:"outputs"`
	LockTime uint32  `json:"lock_time"`
}
// NewTxOut returns a new bitcoin transaction output with the provided
// transaction value and public key script.
func NewTxOut(value uint64, pkScript []byte,asset Asset) *Output {
	return &Output{
		Value:    value,
		PkScript: pkScript,
		Asset : asset,
	}
}
type Output struct {
	Value    uint64
	PkScript []byte
	Asset    Asset
}
type Input struct {
	PreviousOutPoint OutPoint
	SignatureScript  []byte
	Extra            []byte // if user creating a new asset, this field should be it's config data. Otherwise it is null.
}
// NewTxIn returns a new ptn transaction input with the provided
// previous outpoint point and signature script with a default sequence of
// MaxTxInSequenceNum.
func NewTxIn(prevOut *OutPoint, signatureScript []byte) *Input {
	return &Input{
		PreviousOutPoint: *prevOut,
		SignatureScript:  signatureScript,
	}
}
// VarIntSerializeSize returns the number of bytes it would take to serialize
// val as a variable length integer.
func VarIntSerializeSize(val uint64) int {
	// The value is small enough to be represented by itself, so it's
	// just 1 byte.
	if val < 0xfd {
		return 1
	}
	// Discriminant 1 byte plus 2 bytes for the uint16.
	if val <= math.MaxUint16 {
		return 3
	}
	// Discriminant 1 byte plus 4 bytes for the uint32.
	if val <= math.MaxUint32 {
		return 5
	}
	// Discriminant 1 byte plus 8 bytes for the uint64.
	return 9
}
// SerializeSize returns the number of bytes it would take to serialize the
// the transaction output.
func (t *Output) SerializeSize() int {
	// Value 8 bytes + serialized varint size for the length of PkScript +
	// PkScript bytes.
	return 8 + VarIntSerializeSize(uint64(len(t.PkScript))) + len(t.PkScript)
}
func (t *Input) SerializeSize() int {
	// Outpoint Hash 32 bytes + Outpoint Index 4 bytes + Sequence 4 bytes +
	// serialized varint size for the length of SignatureScript +
	// SignatureScript bytes.
	return 40 + VarIntSerializeSize(uint64(len(t.SignatureScript))) +
		len(t.SignatureScript)
}
func (msg *PaymentPayload) SerializeSize() int {
	n := msg.baseSize()
	return n
}
func (msg *Transaction) SerializeSize() int {
	n := msg.baseSize()
	return n
}
// AddTxOut adds a transaction output to the message.
//func (msg *PaymentPayload) AddTxOut(to *Output) {
//	msg.Output = append(msg.Output, to)
//}
// AddTxIn adds a transaction input to the message.
//func (msg *PaymentPayload) AddTxIn(ti *Input) {
//	msg.Input = append(msg.Input, ti)
//}
const HashSize = 32
const defaultTxInOutAlloc = 15
type Hash [HashSize]byte
// DoubleHashH calculates hash(hash(b)) and returns the resulting bytes as a
// Hash.
// TxHash generates the Hash for the transaction.
func (msg *PaymentPayload) TxHash() common.Hash {
	// Encode the transaction and calculate double sha256 on the result.
	// Ignore the error returns since the only way the encode could fail
	// is being out of memory or due to nil pointers, both of which would
	// cause a run-time panic.
	buf := bytes.NewBuffer(make([]byte, 0, msg.SerializeSizeStripped()))
	_ = msg.SerializeNoWitness(buf)
	return common.DoubleHashH(buf.Bytes())
}
// SerializeNoWitness encodes the transaction to w in an identical manner to
// Serialize, however even if the source transaction has inputs with witness
// data, the old serialization format will still be used.
func (msg *PaymentPayload) SerializeNoWitness(w io.Writer) error {
	//return msg.BtcEncode(w, 0, BaseEncoding)
	return nil
}
// baseSize returns the serialized size of the transaction without accounting
// for any witness data.
func (msg *PaymentPayload) baseSize() int {
	// Version 4 bytes + LockTime 4 bytes + Serialized varint size for the
	// number of transaction inputs and outputs.
	n := 8 + VarIntSerializeSize(uint64(len(msg.Input))) +
		VarIntSerializeSize(uint64(len(msg.Output)))
	for _, txIn := range msg.Input {
		n += txIn.SerializeSize()
	}
	for _, txOut := range msg.Output {
		n += txOut.SerializeSize()
	}
	return n
}
func (msg *Transaction) baseSize() int {
	// Version 4 bytes + LockTime 4 bytes + Serialized varint size for the
	// number of transaction inputs and outputs.
	n := 16 + VarIntSerializeSize(uint64(len(msg.TxMessages))) +
		VarIntSerializeSize(uint64(len(msg.TxHash))) +
		VarIntSerializeSize(uint64(len(msg.CreationDate)))
    for _, mtx := range msg.TxMessages {
	    payload := mtx.Payload
		payment, ok := payload.(PaymentPayload)
		if ok == true {
			for _, txIn := range payment.Input {
			    n += txIn.SerializeSize()
		    }
		    for _, txOut := range payment.Output {
			    n += txOut.SerializeSize()
		    }
		}
	}
	return n
}
// SerializeSizeStripped returns the number of bytes it would take to serialize
// the transaction, excluding any included witness data.
func (msg *PaymentPayload) SerializeSizeStripped() int {
	return msg.baseSize()
}
// WriteVarBytes serializes a variable length byte array to w as a varInt
// containing the number of bytes, followed by the bytes themselves.
func WriteVarBytes(w io.Writer, pver uint32, bytes []byte) error {
	slen := uint64(len(bytes))
	err := WriteVarInt(w, pver, slen)
	if err != nil {
		return err
	}
	_, err = w.Write(bytes)
	return err
}
const binaryFreeListMaxItems = 1024
type binaryFreeList chan []byte
var binarySerializer binaryFreeList = make(chan []byte, binaryFreeListMaxItems)
// WriteVarInt serializes val to w using a variable number of bytes depending
// on its value.
func WriteVarInt(w io.Writer, pver uint32, val uint64) error {
	if val < 0xfd {
		return binarySerializer.PutUint8(w, uint8(val))
	}
	if val <= math.MaxUint16 {
		err := binarySerializer.PutUint8(w, 0xfd)
		if err != nil {
			return err
		}
		return binarySerializer.PutUint16(w, littleEndian, uint16(val))
	}
	if val <= math.MaxUint32 {
		err := binarySerializer.PutUint8(w, 0xfe)
		if err != nil {
			return err
		}
		return binarySerializer.PutUint32(w, littleEndian, uint32(val))
	}
	err := binarySerializer.PutUint8(w, 0xff)
	if err != nil {
		return err
	}
	return binarySerializer.PutUint64(w, littleEndian, val)
}
// Borrow returns a byte slice from the free list with a length of 8.  A new
// buffer is allocated if there are not any available on the free list.
func (l binaryFreeList) Borrow() []byte {
	var buf []byte
	select {
	case buf = <-l:
	default:
		buf = make([]byte, 8)
	}
	return buf[:8]
}
// Return puts the provided byte slice back on the free list.  The buffer MUST
// have been obtained via the Borrow function and therefore have a cap of 8.
func (l binaryFreeList) Return(buf []byte) {
	select {
	case l <- buf:
	default:
		// Let it go to the garbage collector.
	}
}
// Uint8 reads a single byte from the provided reader using a buffer from the
// free list and returns it as a uint8.
func (l binaryFreeList) Uint8(r io.Reader) (uint8, error) {
	buf := l.Borrow()[:1]
	if _, err := io.ReadFull(r, buf); err != nil {
		l.Return(buf)
		return 0, err
	}
	rv := buf[0]
	l.Return(buf)
	return rv, nil
}
// Uint16 reads two bytes from the provided reader using a buffer from the
// free list, converts it to a number using the provided byte order, and returns
// the resulting uint16.
func (l binaryFreeList) Uint16(r io.Reader, byteOrder binary.ByteOrder) (uint16, error) {
	buf := l.Borrow()[:2]
	if _, err := io.ReadFull(r, buf); err != nil {
		l.Return(buf)
		return 0, err
	}
	rv := byteOrder.Uint16(buf)
	l.Return(buf)
	return rv, nil
}
// Uint32 reads four bytes from the provided reader using a buffer from the
// free list, converts it to a number using the provided byte order, and returns
// the resulting uint32.
func (l binaryFreeList) Uint32(r io.Reader, byteOrder binary.ByteOrder) (uint32, error) {
	buf := l.Borrow()[:4]
	if _, err := io.ReadFull(r, buf); err != nil {
		l.Return(buf)
		return 0, err
	}
	rv := byteOrder.Uint32(buf)
	l.Return(buf)
	return rv, nil
}
// Uint64 reads eight bytes from the provided reader using a buffer from the
// free list, converts it to a number using the provided byte order, and returns
// the resulting uint64.
func (l binaryFreeList) Uint64(r io.Reader, byteOrder binary.ByteOrder) (uint64, error) {
	buf := l.Borrow()[:8]
	if _, err := io.ReadFull(r, buf); err != nil {
		l.Return(buf)
		return 0, err
	}
	rv := byteOrder.Uint64(buf)
	l.Return(buf)
	return rv, nil
}
// PutUint8 copies the provided uint8 into a buffer from the free list and
// writes the resulting byte to the given writer.
func (l binaryFreeList) PutUint8(w io.Writer, val uint8) error {
	buf := l.Borrow()[:1]
	buf[0] = val
	_, err := w.Write(buf)
	l.Return(buf)
	return err
}
var (
	// littleEndian is a convenience variable since binary.LittleEndian is
	// quite long.
	littleEndian = binary.LittleEndian
	// bigEndian is a convenience variable since binary.BigEndian is quite
	// long.
	bigEndian = binary.BigEndian
)
// PutUint16 serializes the provided uint16 using the given byte order into a
// buffer from the free list and writes the resulting two bytes to the given
// writer.
func (l binaryFreeList) PutUint16(w io.Writer, byteOrder binary.ByteOrder, val uint16) error {
	buf := l.Borrow()[:2]
	byteOrder.PutUint16(buf, val)
	_, err := w.Write(buf)
	l.Return(buf)
	return err
}
// PutUint32 serializes the provided uint32 using the given byte order into a
// buffer from the free list and writes the resulting four bytes to the given
// writer.
func (l binaryFreeList) PutUint32(w io.Writer, byteOrder binary.ByteOrder, val uint32) error {
	buf := l.Borrow()[:4]
	byteOrder.PutUint32(buf, val)
	_, err := w.Write(buf)
	l.Return(buf)
	return err
}
// PutUint64 serializes the provided uint64 using the given byte order into a
// buffer from the free list and writes the resulting eight bytes to the given
// writer.
func (l binaryFreeList) PutUint64(w io.Writer, byteOrder binary.ByteOrder, val uint64) error {
	buf := l.Borrow()[:8]
	byteOrder.PutUint64(buf, val)
	_, err := w.Write(buf)
	l.Return(buf)
	return err
}
func WriteTxOut(w io.Writer, pver uint32, version int32, to *Output) error {
	err := binarySerializer.PutUint64(w, littleEndian, uint64(to.Value))
	if err != nil {
		return err
	}
	return WriteVarBytes(w, pver, to.PkScript)
}
