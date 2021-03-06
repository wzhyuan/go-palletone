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
	"fmt"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

func TestGet(t *testing.T) {
	if m := GetPrefix([]byte("unit")); m != nil {
		for k, v := range m {
			log.Println("key1: ", k, "value: ", string(v))
		}
	}

	go func() {
		if m := GetPrefix([]byte("unit")); m != nil {
			for k, v := range m {
				log.Println("key2: ", k, "value: ", string(v))
			}
		}
	}()

	if m := GetPrefix([]byte("unit")); m != nil {
		for k := range m {
			log.Println("key: ", k, "value: ", string(m[k]))
		}
	}
	time.Sleep(2 * time.Second)
}

func TestGetUnit(t *testing.T) {
	if Dbconn == nil {
		log.Println("dbconn is nil , renew db  start ...")
		Dbconn = ReNewDbConn(dagconfig.DbPath)
	}
	GetUnit(Dbconn, common.HexToHash("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"))
}

func TestGetContract(t *testing.T) {
	var keys []string
	var results []interface{}
	var origin modules.Contract
	//copy(origin.Id[32-len(id):], id.Bytes())
	origin.Id = common.HexToHash("123456")

	origin.Name = "test"
	origin.Code = []byte(`log.PrintLn("hello world")`)
	origin.Input = []byte("input")
	if Dbconn == nil {
		Dbconn = ReNewDbConn(dagconfig.DbPath)
	}

	log.Println("store error: ", StoreBytes(Dbconn, append(CONTRACT_PTEFIX, origin.Id[:]...), origin))
	keys = append(keys, "Id", "id", "Name", "Code", "code", "codes", "inputs")
	results = append(results, common.HexToHash("123456"), nil, "test", []byte(`log.PrintLn("hello world")`), nil, nil, nil)
	log.Println("test data: ", keys)

	for i, k := range keys {
		data, err := GetContractKeyValue(Dbconn, origin.Id, k)
		if !reflect.DeepEqual(data, results[i]) {
			t.Error("test error:", err, "the expect key is:", k, " value is :", results[i], ",but the return value is: ", data)
		}
	}
}

func TestUnitNumberIndex(t *testing.T) {
	key1 := fmt.Sprintf("%s_%s_%d", UNIT_NUMBER_PREFIX, modules.BTCCOIN.String(), 10000)
	key2 := fmt.Sprintf("%s_%s_%d", UNIT_NUMBER_PREFIX, modules.PTNCOIN.String(), 678934)

	if key1 != "nh_btcoin_10000" {
		log.Println("not equal.", key1)
	}
	if key2 != "nh_ptncoin_678934" {
		log.Println("not equal.", key2)
	}
}

func TestGetContractState(t *testing.T) {
	version, value := GetContractState("contract0000", "name")
	log.Println(version)
	log.Println(value)
	data := GetContractAllState([]byte("contract0000"))
	for k, v := range data {
		log.Println(k, v)
	}
}
