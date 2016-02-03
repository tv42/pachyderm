package fuse_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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

func TestRepoReadDir(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(panicOnFail{t})
	defer ctrl.Finish()

	client := mock_pfs.NewMockAPIClient(ctrl)

	const (
		repoName = "foo"
		repoSize = 42
	)
	repoModTime := time.Date(2016, 1, 2, 13, 14, 15, 123456789, time.UTC)

	client.EXPECT().InspectRepo(gomock.Any(), gomock.Eq(&pfs.InspectRepoRequest{
		Repo: &pfs.Repo{Name: repoName},
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

	const (
		commitID   = "blargh"
		commitSize = 13
	)
	commitStartTime := time.Date(2016, 1, 2, 13, 14, 2, 234567890, time.UTC)
	commitFinishTime := time.Date(2016, 1, 2, 13, 14, 7, 345678901, time.UTC)

	client.EXPECT().ListCommit(gomock.Any(), gomock.Eq(&pfs.ListCommitRequest{
		Repo: []*pfs.Repo{
			{Name: repoName},
		},
		CommitType: pfs.CommitType_COMMIT_TYPE_NONE,
		FromCommit: nil,
		Block:      false,
	})).Return(
		&pfs.CommitInfos{
			CommitInfo: []*pfs.CommitInfo{
				{
					Commit: &pfs.Commit{
						Repo: &pfs.Repo{
							Name: repoName,
						},
						Id: commitID,
					},
					CommitType:   pfs.CommitType_COMMIT_TYPE_READ,
					ParentCommit: nil,
					Started:      &google_protobuf.Timestamp{commitStartTime.Unix(), int32(commitStartTime.Nanosecond())},
					Finished:     &google_protobuf.Timestamp{commitFinishTime.Unix(), int32(commitFinishTime.Nanosecond())},
					SizeBytes:    repoSize,
				},
			},
		},
		error(nil),
	)

	client.EXPECT().InspectCommit(gomock.Any(), gomock.Eq(&pfs.InspectCommitRequest{
		Commit: &pfs.Commit{
			Repo: &pfs.Repo{
				Name: repoName,
			},
			Id: commitID,
		},
	})).Return(
		&pfs.CommitInfo{
			Commit: &pfs.Commit{
				Repo: &pfs.Repo{
					Name: repoName,
				},
				Id: commitID,
			},
			CommitType:   pfs.CommitType_COMMIT_TYPE_READ,
			ParentCommit: nil,
			Started:      &google_protobuf.Timestamp{commitStartTime.Unix(), int32(commitStartTime.Nanosecond())},
			Finished:     &google_protobuf.Timestamp{commitFinishTime.Unix(), int32(commitFinishTime.Nanosecond())},
			SizeBytes:    repoSize,
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

	fis, err := ioutil.ReadDir(filepath.Join(mnt.Dir, repoName))
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	if len(fis) != 1 {
		t.Fatalf("expected one commit: %v", fis)
	}
	fi := fis[0]
	if g, e := fi.Name(), commitID; g != e {
		t.Errorf("wrong name: %q != %q", g, e)
	}
	// TODO show commitSize in commit stat?
	if g, e := fi.Size(), int64(0); g != e {
		t.Errorf("wrong size: %v != %v", g, e)
	}
	if g, e := fi.Mode(), os.ModeDir|0555; g != e {
		t.Errorf("wrong mode: %v != %v", g, e)
	}
	// TODO show CommitInfo.StartTime as ctime, CommitInfo.Finished as mtime
	// TODO test ctime via .Sys
	// if g, e := fi.ModTime().UTC(), commitFinishTime; g != e {
	// 	t.Errorf("wrong mtime: %v != %v", g, e)
	// }
}
