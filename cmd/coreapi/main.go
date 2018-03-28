package main

import (
	"context"
	"log"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreapi"
	"github.com/ipfs/go-ipfs/repo/fsrepo"

	"github.com/schomatis/datastore_benchmarks/utils"

)

func main() {

	// Create a temporal repository for this test.
	const ipfsRepoFn = "ipfs_test_repo"
	os.Setenv("IPFS_PATH", ipfsRepoFn)
	if _, err := os.Stat(ipfsRepoFn); err == nil {
		os.RemoveAll(ipfsRepoFn)
	}
	utils.RunShellCommand("ipfs", "init", "--profile=badgerds")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo, err := fsrepo.Open(ipfsRepoFn)
	if err != nil {
		// TODO handle case: daemon running
		// TODO handle case: repo doesn't exist or isn't initialized
		panic(err)
	}

	nd, err := core.NewNode(ctx, &core.BuildCfg{Online: false	, Repo: repo})
	if err != nil {
		log.Fatal(err)
	}

	coreAPI := coreapi.NewCoreAPI(nd)

	addedPath, err := coreAPI.Unixfs().Add(context.Background(), bytes.NewReader([]byte("Test content.")))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Added: %s\n", addedPath)

	catReader, err := coreAPI.Unixfs().Cat(context.Background(), addedPath)
	if err != nil {
		log.Fatal(err)
	}

	bytesRead, _ := ioutil.ReadAll(catReader)
	fmt.Printf("Cat: %s\n", string(bytesRead))
}
