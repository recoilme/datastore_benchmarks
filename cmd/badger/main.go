package main

import (
	"crypto/sha256"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"
	//"encoding/hex"

	"github.com/schomatis/datastore_benchmarks/utils"

	"github.com/dgraph-io/badger"
)

const valLen = 10
const dataMaxSize = 1 * utils.MiB // rough approximation
var entriesNum = int(math.Floor(dataMaxSize / float64(valLen)))

func createBadgerDB(path string, blockSize int) (*badger.DB, badger.Options, error) {
	if _, err := os.Stat(path); err == nil {
		os.RemoveAll(path)
	}
	// Bypass `BadgerdsDatastoreConfig` to access directly Badger's options
	// which aren't exposed in `Create`.
	// dsc, _ := fsrepo.BadgerdsDatastoreConfig(map[string]interface{} {"path": badgerDBPath, "vlogFileSize": "100MB"})
	// db, err := dsc.Create("")
	os.MkdirAll(filepath.Dir(path), 0755)
	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path
	// defopts.ValueLogFileSize = 10 * MiB
	// defopts.MaxTableSize = 1 * MiB
	// defopts.ValueLogLoadingMode = badgerOptions.MemoryMap

	//opts.BlockSize = blockSize

	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}

	return db, opts, nil
}

func main() {

	testBlockSizes := []int{1, 10, 100, 1000}
	for _, bs := range testBlockSizes {
		fmt.Printf("Testing for Block Size: %d\n", bs)
		badgerDB, opts, _ := createBadgerDB(utils.BadgerDBPath, bs)
		testDB(badgerDB, opts, fmt.Sprintf("%s.%d", utils.BadgerProfPath, bs), "svg")
	}
}

func testDB(db *badger.DB, opts badger.Options, profPath string, profFormat string) {
	profFile, err := utils.CreateFileAndDirs(profPath)
	if err != nil {
		panic(err)
	}

	valBytes := make([]byte, valLen)
	keyList := make([][]byte, entriesNum)

	started := time.Now()
	txn := db.NewTransaction(true)
	for i := 0; i < entriesNum; i++ {
		rand.Read(valBytes)
		hash := sha256.Sum256(valBytes)
		keyList[i] = hash[:]
		//keyList[i] = []byte(fmt.Sprintf("%016d", i))
		// fmt.Println(hex.EncodeToString(keyList[i][:]))
		txn.Set(keyList[i][:], valBytes)
	}
	txn.Commit(nil)
	db.Close() // Flush SST to disk
	fmt.Printf("Put time: %s\n", time.Since(started))

	db, err = badger.Open(opts)
	if err != nil {
		panic(err)
	}

	fmt.Println("Inside profiling..")
	pprof.StartCPUProfile(profFile)
	started = time.Now()

	// TODO: Should close and reopen the DB before the read tests?
	// (To avoid cache performance improvements.)
	// db, _= badgerds.NewDatastore(badgerDBPath, &defopts)

	// Profiling gets.
	txn = db.NewTransaction(true)
	for i := 0; i < entriesNum; i++ {
		txn.Get(keyList[i][:])
	}
	txn.Discard()

	pprof.StopCPUProfile()
	fmt.Printf("Get time: %s\n", time.Since(started))

	// TODO: Include the close operation in the profile?
	db.Close()

	utils.ExportPprofOutput(profPath, profFormat)
}
