package fuse_test

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"bazil.org/fuse/fs/fstestutil"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"github.com/pachyderm/pachyderm/src/pfs/server"
	"github.com/pachyderm/pachyderm/src/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/pkg/shard"
	"google.golang.org/grpc"
)

func TestRootReadDir(t *testing.T) {
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
	defer close(quit)
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

	if err := pfsutil.CreateRepo(apiClient, "one"); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}
	if err := pfsutil.CreateRepo(apiClient, "two"); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	if err := fstestutil.CheckDir(mountpoint, map[string]fstestutil.FileInfoCheck{
		"one": func(fi os.FileInfo) error {
			if g, e := fi.Mode(), os.ModeDir|0555; g != e {
				return fmt.Errorf("wrong mode: %v != %v", g, e)
			}
			// TODO show repoSize in repo stat?
			if g, e := fi.Size(), int64(0); g != e {
				t.Errorf("wrong size: %v != %v", g, e)
			}
			// TODO show RepoInfo.Created as time
			// if g, e := fi.ModTime().UTC(), repoModTime; g != e {
			// 	t.Errorf("wrong mtime: %v != %v", g, e)
			// }
			return nil
		},
		"two": func(fi os.FileInfo) error {
			if g, e := fi.Mode(), os.ModeDir|0555; g != e {
				return fmt.Errorf("wrong mode: %v != %v", g, e)
			}
			// TODO show repoSize in repo stat?
			if g, e := fi.Size(), int64(0); g != e {
				t.Errorf("wrong size: %v != %v", g, e)
			}
			// TODO show RepoInfo.Created as time
			// if g, e := fi.ModTime().UTC(), repoModTime; g != e {
			// 	t.Errorf("wrong mtime: %v != %v", g, e)
			// }
			return nil
		},
	}); err != nil {
		t.Errorf("wrong directory content: %v", err)
	}
}

func TestRepoReadDir(t *testing.T) {
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
	defer close(quit)
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
	commitA, err := pfsutil.StartCommit(apiClient, repoName, "")
	if err != nil {
		t.Fatalf("StartCommit: %v", err)
	}
	if err := pfsutil.FinishCommit(apiClient, repoName, commitA.Id); err != nil {
		t.Fatalf("FinishCommit: %v", err)
	}
	t.Logf("finished commit %v", commitA.Id)

	commitB, err := pfsutil.StartCommit(apiClient, repoName, "")
	if err != nil {
		t.Fatalf("StartCommit: %v", err)
	}
	t.Logf("open commit %v", commitB.Id)

	if err := fstestutil.CheckDir(filepath.Join(mountpoint, repoName), map[string]fstestutil.FileInfoCheck{
		commitA.Id: func(fi os.FileInfo) error {
			if g, e := fi.Mode(), os.ModeDir|0555; g != e {
				return fmt.Errorf("wrong mode: %v != %v", g, e)
			}
			// TODO show commitSize in commit stat?
			if g, e := fi.Size(), int64(0); g != e {
				t.Errorf("wrong size: %v != %v", g, e)
			}
			// TODO show CommitInfo.StartTime as ctime, CommitInfo.Finished as mtime
			// TODO test ctime via .Sys
			// if g, e := fi.ModTime().UTC(), commitFinishTime; g != e {
			// 	t.Errorf("wrong mtime: %v != %v", g, e)
			// }
			return nil
		},
		commitB.Id: func(fi os.FileInfo) error {
			if g, e := fi.Mode(), os.ModeDir|0775; g != e {
				return fmt.Errorf("wrong mode: %v != %v", g, e)
			}
			// TODO show commitSize in commit stat?
			if g, e := fi.Size(), int64(0); g != e {
				t.Errorf("wrong size: %v != %v", g, e)
			}
			// TODO show CommitInfo.StartTime as ctime, ??? as mtime
			// TODO test ctime via .Sys
			// if g, e := fi.ModTime().UTC(), commitFinishTime; g != e {
			// 	t.Errorf("wrong mtime: %v != %v", g, e)
			// }
			return nil
		},
	}); err != nil {
		t.Errorf("wrong directory content: %v", err)
	}
}

func TestCommitOpenReadDir(t *testing.T) {
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
	defer close(quit)
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
	t.Logf("open commit %v", commit.Id)

	const (
		greetingName = "greeting"
		greeting     = "Hello, world\n"
		greetingPerm = 0644
	)
	if err := ioutil.WriteFile(filepath.Join(mountpoint, repoName, commit.Id, greetingName), []byte(greeting), greetingPerm); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	const (
		scriptName = "script"
		script     = "#!/bin/sh\necho foo\n"
		scriptPerm = 0750
	)
	if err := ioutil.WriteFile(filepath.Join(mountpoint, repoName, commit.Id, scriptName), []byte(script), scriptPerm); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := fstestutil.CheckDir(filepath.Join(mountpoint, repoName, commit.Id), map[string]fstestutil.FileInfoCheck{
		greetingName: func(fi os.FileInfo) error {
			// TODO respect greetingPerm
			if g, e := fi.Mode(), os.FileMode(0666); g != e {
				return fmt.Errorf("wrong mode: %v != %v", g, e)
			}
			if g, e := fi.Size(), int64(len(greeting)); g != e {
				t.Errorf("wrong size: %v != %v", g, e)
			}
			// TODO show fileModTime as mtime
			// if g, e := fi.ModTime().UTC(), fileModTime; g != e {
			// 	t.Errorf("wrong mtime: %v != %v", g, e)
			// }
			return nil
		},
		scriptName: func(fi os.FileInfo) error {
			// TODO respect scriptPerm
			if g, e := fi.Mode(), os.FileMode(0666); g != e {
				return fmt.Errorf("wrong mode: %v != %v", g, e)
			}
			if g, e := fi.Size(), int64(len(script)); g != e {
				t.Errorf("wrong size: %v != %v", g, e)
			}
			// TODO show fileModTime as mtime
			// if g, e := fi.ModTime().UTC(), fileModTime; g != e {
			// 	t.Errorf("wrong mtime: %v != %v", g, e)
			// }
			return nil
		},
	}); err != nil {
		t.Errorf("wrong directory content: %v", err)
	}
}

func TestCommitFinishedReadDir(t *testing.T) {
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
	defer close(quit)
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
	t.Logf("open commit %v", commit.Id)

	const (
		greetingName = "greeting"
		greeting     = "Hello, world\n"
		greetingPerm = 0644
	)
	if err := ioutil.WriteFile(filepath.Join(mountpoint, repoName, commit.Id, greetingName), []byte(greeting), greetingPerm); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	const (
		scriptName = "script"
		script     = "#!/bin/sh\necho foo\n"
		scriptPerm = 0750
	)
	if err := ioutil.WriteFile(filepath.Join(mountpoint, repoName, commit.Id, scriptName), []byte(script), scriptPerm); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := pfsutil.FinishCommit(apiClient, repoName, commit.Id); err != nil {
		t.Fatalf("FinishCommit: %v", err)
	}

	if err := fstestutil.CheckDir(filepath.Join(mountpoint, repoName, commit.Id), map[string]fstestutil.FileInfoCheck{
		greetingName: func(fi os.FileInfo) error {
			// TODO respect greetingPerm
			if g, e := fi.Mode(), os.FileMode(0666); g != e {
				return fmt.Errorf("wrong mode: %v != %v", g, e)
			}
			if g, e := fi.Size(), int64(len(greeting)); g != e {
				t.Errorf("wrong size: %v != %v", g, e)
			}
			// TODO show fileModTime as mtime
			// if g, e := fi.ModTime().UTC(), fileModTime; g != e {
			// 	t.Errorf("wrong mtime: %v != %v", g, e)
			// }
			return nil
		},
		scriptName: func(fi os.FileInfo) error {
			// TODO respect scriptPerm
			if g, e := fi.Mode(), os.FileMode(0666); g != e {
				return fmt.Errorf("wrong mode: %v != %v", g, e)
			}
			if g, e := fi.Size(), int64(len(script)); g != e {
				t.Errorf("wrong size: %v != %v", g, e)
			}
			// TODO show fileModTime as mtime
			// if g, e := fi.ModTime().UTC(), fileModTime; g != e {
			// 	t.Errorf("wrong mtime: %v != %v", g, e)
			// }
			return nil
		},
	}); err != nil {
		t.Errorf("wrong directory content: %v", err)
	}
}

func TestWriteAndRead(t *testing.T) {
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
	defer close(quit)
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
}
