package notification

import (
	"encoding/hex"
	"encoding/json"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/go-errors/errors"
	"github.com/devt3000/blockexplorer/blockdata"
	"github.com/devt3000/blockexplorer/insight"
	"github.com/devt3000/blockexplorer/insightjson"
	"github.com/devt3000/blockexplorer/mongodb"
	"github.com/devt3000/blockexplorer/subsidy"
	"io/ioutil"
	"log"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var isMainChain bool

var dao = mongodb.MongoDAO{
	"127.0.0.1",
	"viacoin",
}

func init() {
	ParseJson()
	IsMainChain()
}

func IsMainChain() {
	blockchainInfo, err := blockdata.GetBlockChainInfo()
	if err != nil {
		log.Fatalf("Error getting Blockchaininfo via RPC: %s", err)
	}

	if blockchainInfo.Chain != "main" {
		isMainChain = false
	}

	isMainChain = true
}

type Pools []struct {
	PoolName      string   `json:"poolName"`
	URL           string   `json:"url"`
	SearchStrings []string `json:"searchStrings"`
}

var pools Pools

// get the current path: notification/
var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

// read and parse the json file and unmarshal
func ParseJson() {
	path := strings.Split(basepath, "notification")
	jsonFile, err := ioutil.ReadFile(path[0] + "pools.json")
	if err != nil {
		panic(err)
	}
	json.Unmarshal([]byte(jsonFile), &pools)
}

func ProcessBlock(block *btcjson.GetBlockVerboseResult) {
	newBlock, _ := insight.ConvertToInsightBlock(block)
	txs := GetTx(block)

	//add pool info to block before adding into mongodb
	coinbaseText := ParseCoinbaseText(txs[0])
	pool, err := getPoolInfo(coinbaseText)
	if err == nil {
		newBlock.PoolInfo = &pool
	}

	//add reward info
	newBlock.Reward = subsidy.CalcViacoinBlockSubsidy(int32(newBlock.Height), isMainChain)
	newBlock.IsMainChain = isMainChain

	go dao.AddBlock(newBlock)

	AddTransactions(txs, newBlock.Height) // this in a go routine def causes a race conditions
}

// get coinbase hex string by getting the first transaction of the block
// in the tx.Vin[0] and decode the hex string into a normal text
// Example: "52062f503253482f04dee0c7530807ffffff010000000d2f6e6f64655374726174756d2f" -> /nodeStratum/
func ParseCoinbaseText(tx *btcjson.TxRawResult) string {
	src := []byte(tx.Vin[0].Coinbase)

	dst := make([]byte, hex.DecodedLen(len(src)))
	n, err := hex.Decode(dst, src)
	if err != nil {
		log.Printf("Error getting coinbase text: %s", err)
	}

	return string(dst[:n])
}

// range over all pools and within that range over all search strings
// check if a poolSearchString matches the coinbase text
func getPoolInfo(coinbaseText string) (insightjson.Pools, error) {
	var blockMinedByPool insightjson.Pools

	for _, pool := range pools {
		for _, PoolSearchString := range pool.SearchStrings {
			if strings.Contains(coinbaseText, PoolSearchString) {
				blockMinedByPool.PoolName = pool.PoolName
				blockMinedByPool.URL = pool.URL
				return blockMinedByPool, nil
			}
		}
	}
	return blockMinedByPool, errors.New("PoolSearchStrings did not match coinbase text. Unknown mining pool or solo miner")
}

func GetTx(block *btcjson.GetBlockVerboseResult) []*btcjson.TxRawResult {
	Transactions := []*btcjson.TxRawResult{}
	for i := 0; i < len(block.Tx); i++ {
		txhash, _ := chainhash.NewHashFromStr(block.Tx[i])
		tx, _ := blockdata.GetRawTransactionVerbose(txhash)
		Transactions = append(Transactions, tx)
	}

	return Transactions
}

func AddTransactions(transactions []*btcjson.TxRawResult, blockheight int64) {
	for _, transaction := range transactions {
		newTx := insight.TxConverter(transaction, blockheight)
		go dao.AddTransaction(&newTx[0])
		AddrIndex(&newTx[0]) //this in a go routine will cause a race condition
	}
}

func AddrIndex(tx *insightjson.Tx) {
	//receive
	for _, txVout := range tx.Vouts {
		for _, voutAdress := range txVout.ScriptPubKey.Addresses {
			dbAddrInfo, err := dao.GetAddressInfo(txVout.ScriptPubKey.Addresses[0])
			if err != nil {
				addressInfo := createAddressInfo(voutAdress, txVout, tx)
				go dao.AddAddressInfo(addressInfo)
			} else {
				value := int64(txVout.Value * 100000000) // satoshi value to coin value
				go dao.UpdateAddressInfoReceived(&dbAddrInfo, value, true, tx.Txid)
			}
		}
	}

	//sent
	for _, txVin := range tx.Vins {

		dbAddrInfo, err := dao.GetAddressInfo(txVin.Addr)
		value := int64(txVin.ValueSat)

		if err == nil {
			go dao.UpdateAddressInfoSent(&dbAddrInfo, value, true, tx.Txid)
		}
	}
}

var addrPool = sync.Pool{
	New: func() interface{} { return new(insightjson.AddressInfo) },
}

// create address info. An address can only "exist" if it ever received a transaction
// the received is the vout values.
func createAddressInfo(address string, txVout *insightjson.Vout, tx *insightjson.Tx) *insightjson.AddressInfo {

	data := addrPool.Get().(*insightjson.AddressInfo)
	defer addrPool.Put(data)

	data.Address = address
	data.Balance = txVout.Value
	data.BalanceSat = int64(txVout.Value * 100000000)
	data.TotalReceived = txVout.Value
	data.TotalReceivedSat = int64(txVout.Value * 100000000)
	data.TxAppearances = 1
	data.TransactionsID = []string{tx.Txid}
	return data
}
