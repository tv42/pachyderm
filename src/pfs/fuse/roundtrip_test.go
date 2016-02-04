package fuse_test

import (
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"github.com/pachyderm/pachyderm/src/pfs/server"
	"github.com/pachyderm/pachyderm/src/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/pkg/shard"
	"google.golang.org/grpc"
)

func TestRoundtripWriteAndRead(t *testing.T) {
	t.Parallel()

	// don't leave goroutines running
	var wg sync.WaitGroup
	defer wg.Wait()

	tmp, err := ioutil.TempDir("", "pachyderm-test-")
	if err != nil {
		t.Fatalf("tempdir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmp); err != nil {
			t.Errorf("cannot remove tempdir: %v", err)
		}
	}()

	// closed on successful termination
	quit := make(chan struct{})
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("cannot listen: %v", err)
	}
	defer func() {
		_ = listener.Close()
	}()

	// TODO try to share more of this setup code with various main
	// functions
	localAddress := listener.Addr().String()
	srv := grpc.NewServer()
	const (
		numShards = 1
	)
	sharder := shard.NewLocalSharder(localAddress, numShards)
	hasher := pfs.NewHasher(numShards, 1)
	router := shard.NewRouter(
		sharder,
		grpcutil.NewDialer(
			grpc.WithInsecure(),
		),
		localAddress,
	)

	blockDir := filepath.Join(tmp, "blocks")
	blockServer, err := server.NewLocalBlockAPIServer(blockDir)
	if err != nil {
		t.Fatalf("NewLocalBlockAPIServer: %v", err)
	}
	pfs.RegisterBlockAPIServer(srv, blockServer)

	driver, err := drive.NewDriver(localAddress)
	if err != nil {
		t.Fatalf("NewDriver: %v", err)
	}

	apiServer := server.NewAPIServer(
		hasher,
		router,
	)
	pfs.RegisterAPIServer(srv, apiServer)

	internalAPIServer := server.NewInternalAPIServer(
		hasher,
		router,
		driver,
	)
	pfs.RegisterInternalAPIServer(srv, internalAPIServer)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.Serve(listener); err != nil {
			select {
			case <-quit:
				// orderly shutdown
				return
			default:
				t.Errorf("grpc serve: %v", err)
			}
		}
	}()

	clientConn, err := grpc.Dial(localAddress, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("grpc dial: %v", err)
	}
	apiClient := pfs.NewAPIClient(clientConn)
	mounter := fuse.NewMounter(localAddress, apiClient)

	mountpoint := filepath.Join(tmp, "mnt")
	if err := os.Mkdir(mountpoint, 0700); err != nil {
		t.Fatalf("mkdir mountpoint: %v", err)
	}

	ready := make(chan bool)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := mounter.Mount(mountpoint, nil, nil, ready); err != nil {
			t.Errorf("mount and serve: %v", err)
		}
	}()

	<-ready
	defer func() {
		if err := mounter.Unmount(mountpoint); err != nil {
			t.Errorf("unmount: %v", err)
		}
	}()

	const (
		repoName = "foo"
	)
	if err := pfsutil.CreateRepo(apiClient, repoName); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}
	commit, err := pfsutil.StartCommit(apiClient, repoName, "")
	if err != nil {
		t.Fatalf("StartCommit: %v", err)
	}

	const (
		greeting = "Hello, world\n"
	)
	filePath := filepath.Join(mountpoint, repoName, commit.Id, "greeting")

	if err := ioutil.WriteFile(filePath, []byte(greeting), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	buf, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if g, e := string(buf), greeting; g != e {
		t.Errorf("wrong content: %q != %q", g, e)
	}

	close(quit)
}
