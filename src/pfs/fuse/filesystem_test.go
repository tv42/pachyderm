package fuse_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"bazil.org/fuse/fs/fstestutil"
	"github.com/golang/mock/gomock"
	"github.com/pachyderm/pachyderm/src/internal/mock_pfs"
	"github.com/pachyderm/pachyderm/src/pfs"
	pfuse "github.com/pachyderm/pachyderm/src/pfs/fuse"
	"go.pedge.io/pb/go/google/protobuf"
)

// panicOnFail is a gomock TestReporter that panics instead of calling
// t.Fatal. This lets the FUSE serve loop handle the panic, instead of
// t.FailNow calling runtime.Goexit, as that loses the response and
// causes a hang.
type panicOnFail struct {
	testing.TB
}

var _ gomock.TestReporter = panicOnFail{}

func (p panicOnFail) Fatalf(format string, args ...interface{}) {
	p.Errorf(format, args...)
	panic(fmt.Errorf(format, args...))
}

func TestMount(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock_pfs.NewMockAPIClient(ctrl)
	shard := &pfs.Shard{}
	commitMounts := []*pfuse.CommitMount{}
	filesys := pfuse.NewFilesystem(client, shard, commitMounts)

	mnt, err := fstestutil.MountedT(t, filesys, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer mnt.Close()
}

func TestRootReadDir(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(panicOnFail{t})
	defer ctrl.Finish()

	client := mock_pfs.NewMockAPIClient(ctrl)

	const (
		repoName = "foo"
		repoSize = 42
	)
	repoModTime := time.Date(2016, 1, 2, 13, 14, 15, 123456789, time.UTC)

	client.EXPECT().ListRepo(gomock.Any(), gomock.Eq(&pfs.ListRepoRequest{})).Return(
		&pfs.RepoInfos{
			RepoInfo: []*pfs.RepoInfo{
				{
					Repo: &pfs.Repo{
						Name: repoName,
					},
					Created:   &google_protobuf.Timestamp{repoModTime.Unix(), int32(repoModTime.Nanosecond())},
					SizeBytes: repoSize,
				},
			},
		},
		error(nil),
	)

	client.EXPECT().InspectRepo(gomock.Any(), gomock.Eq(&pfs.InspectRepoRequest{
		Repo: &pfs.Repo{
			Name: repoName,
		},
	})).Return(
		&pfs.RepoInfo{
			Repo: &pfs.Repo{
				Name: repoName,
			},
			Created:   &google_protobuf.Timestamp{repoModTime.Unix(), int32(repoModTime.Nanosecond())},
			SizeBytes: repoSize,
		},
		error(nil),
	)

	shard := &pfs.Shard{}
	commitMounts := []*pfuse.CommitMount{}
	filesys := pfuse.NewFilesystem(client, shard, commitMounts)

	mnt, err := fstestutil.MountedT(t, filesys, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer mnt.Close()

	fis, err := ioutil.ReadDir(mnt.Dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	if len(fis) != 1 {
		t.Fatalf("expected one repo: %v", fis)
	}
	fi := fis[0]
	if g, e := fi.Name(), repoName; g != e {
		t.Errorf("wrong name: %q != %q", g, e)
	}
	// TODO show repoSize in repo stat?
	if g, e := fi.Size(), int64(0); g != e {
		t.Errorf("wrong size: %v != %v", g, e)
	}
	if g, e := fi.Mode(), os.ModeDir|0555; g != e {
		t.Errorf("wrong mode: %v != %v", g, e)
	}
	// TODO show RepoInfo.Created as time
	// if g, e := fi.ModTime().UTC(), repoModTime; g != e {
	// 	t.Errorf("wrong mtime: %v != %v", g, e)
	// }
}
