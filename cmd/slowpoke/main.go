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

	"github.com/recoilme/slowpoke"
	"github.com/schomatis/datastore_benchmarks/utils"
)

const valLen = 10
const dataMaxSize = 1 * utils.MiB // rough approximation
var entriesNum = int(math.Floor(dataMaxSize / float64(valLen)))

func createSlowpokeDirs(path string) error {
	if _, err := os.Stat(path); err == nil {
		os.RemoveAll(path)
	}
	// Bypass `BadgerdsDatastoreConfig` to access directly Badger's options
	// which aren't exposed in `Create`.
	// dsc, _ := fsrepo.BadgerdsDatastoreConfig(map[string]interface{} {"path": badgerDBPath, "vlogFileSize": "100MB"})
	// db, err := dsc.Create("")
	return os.MkdirAll(filepath.Dir(path), 0755)

}

func main() {

	testBlockSizes := []int{1, 10, 100, 1000}
	for _, bs := range testBlockSizes {
		fmt.Printf("Testing for Block Size: %d\n", bs)
		createSlowpokeDirs(utils.SlowpokeDBPath)
		//badgerDB, opts, _ := createBadgerDB(utils.BadgerDBPath, bs)
		testDB(utils.SlowpokeDBFile, fmt.Sprintf("%s.%d", utils.BadgerProfPath, bs), "svg")
	}
}

func testDB(storage string, profPath string, profFormat string) {
	profFile, err := utils.CreateFileAndDirs(profPath)
	if err != nil {
		panic(err)
	}

	valBytes := make([]byte, valLen)
	keyList := make([][]byte, entriesNum)

	started := time.Now()
	//txn := db.NewTransaction(true)

	var pairs [][]byte

	for i := 0; i < entriesNum; i++ {
		rand.Read(valBytes)
		hash := sha256.Sum256(valBytes)
		keyList[i] = hash[:]
		//keyList[i] = []byte(fmt.Sprintf("%016d", i))
		// fmt.Println(hex.EncodeToString(keyList[i][:]))
		pairs = append(pairs, keyList[i][:])
		pairs = append(pairs, valBytes)

	}
	slowpoke.Sets(storage, pairs)
	//txn.Commit(nil)
	slowpoke.Close(storage) // Remove keys from memory
	fmt.Printf("Put time: %s\n", time.Since(started))

	_, err = slowpoke.Open(storage)
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
	//txn = db.NewTransaction(true)
	for i := 0; i < entriesNum; i++ {
		//txn.Get(keyList[i][:])
		slowpoke.Get(storage, keyList[i][:])
		//slowpoke has gets command
	}
	//txn.Discard()

	pprof.StopCPUProfile()
	fmt.Printf("Get time: %s\n", time.Since(started))

	// TODO: Include the close operation in the profile?
	slowpoke.Close(storage)

	utils.ExportPprofOutput(profPath, profFormat)
}
