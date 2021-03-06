package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"math/rand"
	"math"
	"path/filepath"
	"time"

	"github.com/schomatis/datastore_benchmarks/utils"

	badgerds "gx/ipfs/Qmbjb3c2KRPVNZWSvQED8zAf12Brdbp3ksSnGdsJiytqUs/go-ds-badger"
	// badgerOptions "gx/ipfs/QmdKhi5wUQyV9i3GcTyfUmpfTntWjXu8DcyT9HyNbznYrn/badger/options"
	ds "gx/ipfs/QmPpegoMqhAEqjncrzArm7KVWAkCm78rqL2DPuNjhPrshg/go-datastore"
	// "gx/ipfs/QmZooytqEoUwQjv7KzH4d3xyJnyvD3AWJaCDMYt5pbCtua/chunker"
	repo "github.com/ipfs/go-ipfs/repo"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"
	mh "gx/ipfs/QmZyZDi491cCNTLfAhwcaDii2Kg4pwKRkhqQzURGDvY6ua/go-multihash"
	cid "gx/ipfs/QmcZfnkapfECQGcLZaf9B79NRg7cRa9EnZh4LSbkCzwNvY/go-cid"
)

var doRandomReads = false // false: sequential reads.

// /* const */ var valLen = chunker.DefaultBlockSize
// TODO: Why isn't DefaultBlockSize a constant?
const valLen = 100 * utils.KiB // Average file size in go-ipfs
// TODO: Parametrize this.
const dataMaxSize = 250 * utils.MiB // rough approximation
var entriesNum = int(math.Floor(dataMaxSize / float64(valLen)))


func createBadgerDB(path string) (repo.Datastore, error) {
	if _, err := os.Stat(path); err == nil {
		os.RemoveAll(path)
	}
	// Bypass `BadgerdsDatastoreConfig` to access directly Badger's options
	// which aren't exposed in `Create`.
	// dsc, _ := fsrepo.BadgerdsDatastoreConfig(map[string]interface{} {"path": badgerDBPath, "vlogFileSize": "100MB"})
	// db, err := dsc.Create("")
	os.MkdirAll(filepath.Dir(path), 0755)
	defopts := badgerds.DefaultOptions
	// defopts.ValueLogFileSize = 10 * MiB
	// defopts.MaxTableSize = 1 * MiB
	// defopts.ValueLogLoadingMode = badgerOptions.MemoryMap

	var db repo.Datastore

	db, err := badgerds.NewDatastore(path, &defopts)
	if err != nil {
		panic(err)
	}

	return db, nil
}

func createFlatfsDB(path string) (repo.Datastore, error) {
	if _, err := os.Stat(path); err == nil {
		os.RemoveAll(path)
	}

	dsc, err := fsrepo.FlatfsDatastoreConfig(map[string]interface{}{
		"path": utils.FlatfsDBPath,
		"shardFunc": "/repo/flatfs/shard/v1/next-to-last/2",
		"sync": true})
	if err != nil {
		panic(err)
	}
	flatDB, _ := dsc.Create("")
	return flatDB, nil
}

func main() {
	badgerDB, _ := createBadgerDB(utils.BadgerDBPath)
	// flatfsDB, _ := createFlatfsDB(flatfsDBPath)

	testDB(badgerDB, utils.BadgerProfPath, "svg")
	// testDB(flatfsDB, flatfsProfPath, "svg")
}

func testDB(db repo.Datastore, profPath string, profFormat string) {
	profFile, err := utils.CreateFileAndDirs(profPath)
	if err != nil {
		panic(err)
	}

	valBytes := make([]byte, valLen)
	keyList := make([]ds.Key, entriesNum)

	for i := 0; i < entriesNum; i++ {
		rand.Read(valBytes)

		// Compute the hash of the random data.
		prefixV0 := cid.NewPrefixV0(mh.SHA2_256)
		cid, _ := prefixV0.Sum(valBytes)
		keyList[i] = ds.NewKey(string(cid.String()))

		db.Put(keyList[i], valBytes)
	}

	fmt.Println("Inside profiling..")
	started := time.Now()
	pprof.StartCPUProfile(profFile)

	// TODO: Should close and reopen the DB before the read tests?
	// (To avoid cache performance improvements.)
	// db, _= badgerds.NewDatastore(badgerDBPath, &defopts)

	// Profiling gets.
	for i := 0; i < entriesNum; {
		if doRandomReads {
			_, err = db.Get(keyList[rand.Int31n(int32(entriesNum))]); // Random
		} else {
			_, err = db.Get(keyList[i]) // Sequential
		}
		if err != nil {
			panic(err)
		}
		i++
	}

	pprof.StopCPUProfile()
	fmt.Printf("Get time: %s\n", time.Since(started))

	// TODO: Include the close operation in the profile?
	db.Close()

	utils.ExportPprofOutput(profPath, profFormat)
}
