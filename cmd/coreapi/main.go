package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/ipfs/go-ipfs/core"
	//"github.com/ipfs/go-ipfs/core/coreapi"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	"github.com/ipfs/go-ipfs/unixfs/archive"

	"github.com/ipfs/go-ipfs/path"

	"github.com/schomatis/datastore_benchmarks/utils"
)

func main() {
	testGet(utils.BadgerRepoPath, "badgerds", utils.BadgerProfPath, "svg")
	testGet(utils.FlatfsRepoPath, "default-datastore", utils.FlatfsProfPath, "svg")
}

func testGet(repoPath string, initProfile string, profPath string, profFormat string) {
	profFile, err := utils.CreateFileAndDirs(profPath)
	if err != nil {
		panic(err)
	}

	addedHash := createRepoAddPath(repoPath, fmt.Sprintf("%s/src/github.com/golang/go", os.Getenv("GOPATH")), initProfile)
	// addedHash := "QmUx5yshtZb2J2ww16gRbiM3onLx6BaEoVUk9P3R1ieUt6"
	// TODO: Save repo to avoid adding it every time?
	// TODO: Also, just adding the files being fetched might keep them cached.

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo, err := fsrepo.Open(repoPath)
	if err != nil {
		panic(err)
	}

	nd, err := core.NewNode(ctx, &core.BuildCfg{Online: false, Repo: repo})
	if err != nil {
		panic(err)
	}

	pprof.StartCPUProfile(profFile)
	started := time.Now()

	// Logic taken from `ipfs get`, I didn't see a similar `coreapi` function.
	dn, err := core.Resolve(ctx, nd.Namesys, nd.Resolver, path.Path(fmt.Sprintf("/ipfs/%s", addedHash)))
	if err != nil {
		panic(err)
	}

	getReader, err := archive.DagArchive(ctx, dn, addedHash, nd.DAG, false, 0)
	if err != nil {
		panic(err)
	}

	bytesRead, _ := ioutil.ReadAll(getReader)
	fmt.Printf("Get: %d\n", len(bytesRead))

	pprof.StopCPUProfile()
	fmt.Printf("Get time: %s\n", time.Since(started))

	// TODO: Tear down node?
	// nd.teardown()

	utils.ExportPprofOutput(profPath, profFormat)
}

func createRepoAddPath(repoPath, addPath string, initProfile string) (string) {
	// Create a temporal repository for this test.
	os.Setenv("IPFS_PATH", repoPath)
	// TODO: This should be set in  the commands themselves.

	if _, err := os.Stat(repoPath); err == nil {
		os.RemoveAll(repoPath)
	}
	os.MkdirAll(filepath.Dir(repoPath), os.ModePerm)

	utils.RunShellCommand("ipfs", "init", fmt.Sprintf("--profile=%s", initProfile))

	started := time.Now()
	addedHash := utils.RunShellCommand("ipfs", "add", addPath, "-r", "-Q")
	fmt.Printf("Add time: %s\n", time.Since(started))

	return strings.TrimSuffix(addedHash, "\n")
}
