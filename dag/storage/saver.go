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

package storage

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/palletone/go-palletone/common"
	palletdb "github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/internal/ptnapi"
)

var (
	Dbconn             *palletdb.LDBDatabase = nil
	AssocUnstableUnits map[string]modules.Joint
	//DBPath             string = "/Users/jay/code/gocode/src/palletone/bin/leveldb"
	DBPath string = dagconfig.DbPath
)

func SaveJoint(objJoint *modules.Joint, onDone func()) (err error) {
	if objJoint.Unsigned != "" {
		return errors.New(objJoint.Unsigned)
	}
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DbPath)
	}
	obj_unit := objJoint.Unit
	obj_unit_byte, _ := json.Marshal(obj_unit)

	if err = Dbconn.Put(append(UNIT_PREFIX, obj_unit.Hash().Bytes()...), obj_unit_byte); err != nil {
		return
	}
	// add key in  unit_keys
	log.Println("add unit key:", string(UNIT_PREFIX)+obj_unit.Hash().String(), AddUnitKeys(string(UNIT_PREFIX)+obj_unit.Hash().String()))

	if dagconfig.SConfig.Blight {
		// save  update utxo , message , transaction

	}

	if onDone != nil {
		onDone()
	}
	return
}

/**
key: [HEADER_PREFIX][chain index number]_[chain index]_[unit hash]
value: unit header rlp encoding bytes
*/
// save header
func SaveHeader(uHash common.Hash, h *modules.Header) error {
	encNum := encodeBlockNumber(h.Number.Index)
	key := append(HEADER_PREFIX, encNum...)
	key = append(key, h.Number.Bytes()...)
	return StoreBytes(Dbconn, append(key, uHash.Bytes()...), h)
	//key := fmt.Sprintf("%s%v_%s_%s", HEADER_PREFIX, h.Number.Index, h.Number.String(), uHash.Bytes())
	//return StoreBytes(Dbconn, []byte(key), h)
}

func SaveHashNumber(uHash common.Hash, height modules.ChainIndex) error {
	return StoreBytes(Dbconn, append(UNIT_HASH_NUMBER_Prefix, uHash.Bytes()...), height)
}

// height and assetid can get a unit key.
func SaveUHashIndex(cIndex modules.ChainIndex, uHash common.Hash) error {
	key := fmt.Sprintf("%s_%s_%d", UNIT_NUMBER_PREFIX, cIndex.AssetID.String(), cIndex.Index)
	return Store(Dbconn, key, uHash)
}

/**
key: [BODY_PREFIX][unit hash]
value: all transactions hash set's rlp encoding bytes
*/
func SaveBody(unitHash common.Hash, txsHash []common.Hash) error {
	// Dbconn.Put(append())
	return StoreBytes(Dbconn, append(BODY_PREFIX, unitHash.Bytes()...), txsHash)
}

func GetBody(unitHash common.Hash) ([]common.Hash, error) {
	data, err := Get(append(BODY_PREFIX, unitHash.Bytes()...))
	if err != nil {
		return nil, err
	}
	var txHashs []common.Hash
	if err := rlp.DecodeBytes(data, &txHashs); err != nil {
		return nil, err
	}
	return txHashs, nil
}

func SaveTransactions(txs *modules.Transactions) error {
	key := fmt.Sprintf("%s%s", TRANSACTIONS_PREFIX, txs.Hash())
	return Store(Dbconn, key, *txs)
}

/**
key: [TRANSACTION_PREFIX][tx hash]
value: transaction struct rlp encoding bytes
*/
func SaveTransaction(tx *modules.Transaction) error {
	// save transaction
	if err := StoreBytes(Dbconn, append(TRANSACTION_PREFIX, tx.TxHash.Bytes()...), tx); err != nil {
		return err
	}

	if err := StoreBytes(Dbconn, append(Transaction_Index, tx.TxHash.Bytes()...), tx); err != nil {
		return err
	}
	updateAddrTransactions(tx.Address().String(), tx.TxHash)
	// store output by addr
	for i, msg := range tx.TxMessages {
		payload, ok := msg.Payload.(modules.PaymentPayload)
		if ok {
			for _, output := range payload.Output {
				//  pkscript to addr
				addr, _ := ptnapi.GetAddressFromScript(output.PkScript[:])
				saveOutputByAddr(addr, tx.TxHash, i, *output)
			}
		}
	}

	return nil
}
func updateAddrTransactions(addr string, hash common.Hash) error {
	if hash == (common.Hash{}) {
		return errors.New("empty tx hash.")
	}
	hashs := make([]common.Hash, 0)
	data, err := Get(append(AddrTransactionsHash_Prefix, []byte(addr)...))
	if err != nil {
		if err.Error() != "leveldb: not found" {
			return err
		} else { // first store the addr
			hashs = append(hashs, hash)
			if err := StoreBytes(Dbconn, append(AddrTransactionsHash_Prefix, []byte(addr)...), hashs); err != nil {
				return err
			}
			return nil
		}
	}
	if err := rlp.DecodeBytes(data, hashs); err != nil {
		return err
	}
	hashs = append(hashs, hash)
	if err := StoreBytes(Dbconn, append(AddrTransactionsHash_Prefix, []byte(addr)...), hashs); err != nil {
		return err
	}
	return nil
}
func saveOutputByAddr(addr string, hash common.Hash, msgindex int, output modules.Output) error {
	if hash == (common.Hash{}) {
		return errors.New("empty tx hash.")
	}
	key := append(AddrOutput_Prefix, []byte(addr)...)
	key = append(key, hash.Bytes()...)
	if err := StoreBytes(Dbconn, append(key, new(big.Int).SetInt64(int64(msgindex)).Bytes()...), output); err != nil {
		return err
	}
	return nil
}
func SaveTxLookupEntry(unit *modules.Unit) error {
	for i, tx := range unit.Transactions() {
		in := modules.TxLookupEntry{
			UnitHash:  unit.Hash(),
			UnitIndex: unit.NumberU64(),
			Index:     uint64(i),
		}
		data, err := rlp.EncodeToBytes(in)
		if err != nil {
			return err
		}
		if err := StoreBytes(Dbconn, append(LookupPrefix, tx.TxHash.Bytes()...), data); err != nil {
			return err
		}
	}
	return nil
}

// encodeBlockNumber encodes a block number as big endian uint64
func encodeBlockNumber(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}

func GetUnitKeys() []string {
	var keys []string
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DbPath)
	}
	if keys_byte, err := Dbconn.Get([]byte("array_units")); err != nil {
		log.Println("get units error:", err)
	} else {
		if err := rlp.DecodeBytes(keys_byte[:], &keys); err != nil {
			log.Println("error:", err)
		}
	}
	return keys
}
func AddUnitKeys(key string) error {
	keys := GetUnitKeys()
	if len(keys) <= 0 {
		return errors.New("null keys.")
	}
	for _, v := range keys {

		if v == key {
			return errors.New("key is already exist.")
		}
	}
	keys = append(keys, key)
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DbPath)
	}

	return Store(Dbconn, "array_units", keys)
}
func ConvertBytes(val interface{}) (re []byte) {
	var err error
	if re, err = json.Marshal(val); err != nil {
		log.Println("json.marshal error:", err)
	}
	return re[:]
}

func IsGenesisUnit(unit string) bool {
	return unit == constants.GENESIS_UNIT
}

func GetKeysWithTag(tag string) []string {
	var keys []string
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DbPath)
	}
	if keys_byte, err := Dbconn.Get([]byte(tag)); err != nil {
		log.Println("get keys error:", err)
	} else {
		if err := json.Unmarshal(keys_byte, &keys); err != nil {
			log.Println("error:", err)
		}
	}
	return keys
}
func AddKeysWithTag(key, tag string) error {
	keys := GetKeysWithTag(tag)
	if len(keys) <= 0 {
		return errors.New("null keys.")
	}
	log.Println("keys:=", keys)
	for _, v := range keys {
		if v == key {
			return errors.New("key is already exist.")
		}
	}
	keys = append(keys, key)
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DbPath)
	}

	if err := Dbconn.Put([]byte(tag), ConvertBytes(keys)); err != nil {
		return err
	}
	return nil

}

func SaveContract(contract *modules.Contract) (common.Hash, error) {
	if common.EmptyHash(contract.CodeHash) {
		contract.CodeHash = rlp.RlpHash(contract.Code)
	}
	// key = cs+ rlphash(contract)
	if common.EmptyHash(contract.Id) {
		ids := rlp.RlpHash(contract)
		if len(ids) > len(contract.Id) {
			id := ids[len(ids)-common.HashLength:]
			copy(contract.Id[common.HashLength-len(id):], id)
		} else {
			//*contract.Id = new(common.Hash)
			copy(contract.Id[common.HashLength-len(ids):], ids[:])
		}

	}
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DbPath)
	}
	return contract.Id, StoreBytes(Dbconn, append(CONTRACT_PTEFIX, contract.Id[:]...), contract)
}

//  get  unit chain version
// GetUnitChainVersion reads the version number from db.
func GetUnitChainVersion() int {
	var vsn uint
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DbPath)
	}
	enc, _ := Dbconn.Get([]byte("UnitchainVersion"))
	rlp.DecodeBytes(enc, &vsn)
	return int(vsn)
}

// SaveUnitChainVersion writes vsn as the version number to db.
func SaveUnitChainVersion(vsn int) error {
	enc, _ := rlp.EncodeToBytes(uint(vsn))
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DbPath)
	}
	return Dbconn.Put([]byte("UnitchainVersion"), enc)
}

/**
保存合约属性信息
To save contract
*/
func SaveContractState(prefix []byte, id []byte, name string, value interface{}, version modules.StateVersion) bool {
	key := fmt.Sprintf("%s%s^*^%s^*^%s",
		prefix,
		id,
		name,
		version.String())
	if err := Store(Dbconn, key, value); err != nil {
		log.Println("Save contract template", "error", err.Error())
		return false
	}
	return true
}
