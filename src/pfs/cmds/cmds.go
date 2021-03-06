package cmds

import (
	"fmt"
	"os"
	"path"
	"strings"
	"text/tabwriter"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"github.com/pachyderm/pachyderm/src/pfs/pretty"
	"github.com/spf13/cobra"
	//"go.pedge.io/lion"
	"go.pedge.io/pkg/cobra"
	"google.golang.org/grpc"
)

func Cmds(address string) []*cobra.Command {
	var fileNumber int
	var fileModulus int
	var blockNumber int
	var blockModulus int
	shard := func() *pfs.Shard {
		return &pfs.Shard{
			FileNumber:   uint64(fileNumber),
			FileModulus:  uint64(fileModulus),
			BlockNumber:  uint64(blockNumber),
			BlockModulus: uint64(blockModulus),
		}
	}

	addShardFlags := func(cmd *cobra.Command) {
		cmd.Flags().IntVarP(&fileNumber, "file-shard", "s", 0, "file shard to read")
		cmd.Flags().IntVarP(&fileModulus, "file-modulus", "m", 1, "modulus of file shard")
		cmd.Flags().IntVarP(&blockNumber, "block-shard", "b", 0, "block shard to read")
		cmd.Flags().IntVarP(&blockModulus, "block-modulus", "n", 1, "modulus of block shard")
	}

	repo := &cobra.Command{
		Use:   "repo",
		Short: "Docs for repos.",
		Long: `Repos, short for repository, are the top level data object in Pachyderm.

Repos are created with create-repo.`,
		Run: pkgcobra.RunFixedArgs(0, func(args []string) error {
			return nil
		}),
	}

	createRepo := &cobra.Command{
		Use:   "create-repo repo-name",
		Short: "Create a new repo.",
		Long:  "Create a new repo.",
		Run: pkgcobra.RunFixedArgs(1, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			return pfsutil.CreateRepo(apiClient, args[0])
		}),
	}

	inspectRepo := &cobra.Command{
		Use:   "inspect-repo repo-name",
		Short: "Return info about a repo.",
		Long:  "Return info about a repo.",
		Run: pkgcobra.RunFixedArgs(1, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			repoInfo, err := pfsutil.InspectRepo(apiClient, args[0])
			if err != nil {
				return err
			}
			if repoInfo == nil {
				return fmt.Errorf("repo %s not found", args[0])
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintRepoHeader(writer)
			pretty.PrintRepoInfo(writer, repoInfo)
			return writer.Flush()
		}),
	}

	listRepo := &cobra.Command{
		Use:   "list-repo",
		Short: "Return all repos.",
		Long:  "Reutrn all repos.",
		Run: pkgcobra.RunFixedArgs(0, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			repoInfos, err := pfsutil.ListRepo(apiClient)
			if err != nil {
				return err
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintRepoHeader(writer)
			for _, repoInfo := range repoInfos {
				pretty.PrintRepoInfo(writer, repoInfo)
			}
			return writer.Flush()
		}),
	}

	deleteRepo := &cobra.Command{
		Use:   "delete-repo repo-name",
		Short: "Delete a repo.",
		Long:  "Delete a repo.",
		Run: pkgcobra.RunFixedArgs(1, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			return pfsutil.DeleteRepo(apiClient, args[0])
		}),
	}

	commit := &cobra.Command{
		Use:   "commit",
		Short: "Docs for commits.",
		Long: `Commits are atomic transactions on the content of a repo.

Creating a commit is a multistep process: 
- start a new commit with start-commit
- write files to it through fuse or with put-file
- finish the new commit with finish-commit

Commits that have been started but not finished are NOT durable storage.
Commits become reliable (and immutable) when they are finished.

Commits can be created with another commit as a parent.
This layers the data in the commit over the data in the parent.`,
		Run: pkgcobra.RunFixedArgs(0, func(args []string) error {
			return nil
		}),
	}

	startCommit := &cobra.Command{
		Use:   "start-commit repo-name [parent-commit-id]",
		Short: "Start a new commit.",
		Long:  "Start a new commit with parent-commit-id as the parent.",
		Run: pkgcobra.RunBoundedArgs(pkgcobra.Bounds{Min: 1, Max: 2}, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			parentCommitID := ""
			if len(args) == 2 {
				parentCommitID = args[1]
			}
			commit, err := pfsutil.StartCommit(apiClient, args[0], parentCommitID)
			if err != nil {
				return err
			}
			fmt.Println(commit.Id)
			return nil
		}),
	}

	finishCommit := &cobra.Command{
		Use:   "finish-commit repo-name commit-id",
		Short: "Finish a started commit.",
		Long:  "Finish a started commit. Commit-id must be a writeable commit.",
		Run: pkgcobra.RunFixedArgs(2, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			return pfsutil.FinishCommit(apiClient, args[0], args[1])
		}),
	}

	inspectCommit := &cobra.Command{
		Use:   "inspect-commit repo-name commit-id",
		Short: "Return info about a commit.",
		Long:  "Return info about a commit.",
		Run: pkgcobra.RunFixedArgs(2, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			commitInfo, err := pfsutil.InspectCommit(apiClient, args[0], args[1])
			if err != nil {
				return err
			}
			if commitInfo == nil {
				return fmt.Errorf("commit %s not found", args[1])
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintCommitInfoHeader(writer)
			pretty.PrintCommitInfo(writer, commitInfo)
			return writer.Flush()
		}),
	}

	listCommit := &cobra.Command{
		Use:   "list-commit repo-name",
		Short: "Return all commits on a repo.",
		Long:  "Return all commits on a repo.",
		Run: pkgcobra.RunFixedArgs(1, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			commitInfos, err := pfsutil.ListCommit(apiClient, args)
			if err != nil {
				return err
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintCommitInfoHeader(writer)
			for _, commitInfo := range commitInfos {
				pretty.PrintCommitInfo(writer, commitInfo)
			}
			return writer.Flush()
		}),
	}

	deleteCommit := &cobra.Command{
		Use:   "delete-commit repo-name commit-id",
		Short: "Delete a commit.",
		Long:  "Delete a commit.",
		Run: pkgcobra.RunFixedArgs(2, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			return pfsutil.DeleteCommit(apiClient, args[0], args[1])
		}),
	}

	file := &cobra.Command{
		Use:   "file",
		Short: "Docs for files.",
		Long: `Files are the lowest level data object in Pachyderm.

Files can be written to started (but not finished) commits with put-file.
Files can be read from finished commits with get-file.`,
		Run: pkgcobra.RunFixedArgs(0, func(args []string) error {
			return nil
		}),
	}

	putFile := &cobra.Command{
		Use:   "put-file repo-name commit-id path/to/file",
		Short: "Put a file from stdin",
		Long:  "Put a file from stdin. Directories must exist. commit-id must be a writeable commit.",
		Run: pkgcobra.RunFixedArgs(3, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			_, err = pfsutil.PutFile(apiClient, args[0], args[1], args[2], 0, os.Stdin)
			return err
		}),
	}

	getFile := &cobra.Command{
		Use:   "get-file repo-name commit-id path/to/file",
		Short: "Return the contents of a file.",
		Long:  "Return the contents of a file.",
		Run: pkgcobra.RunFixedArgs(3, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			return pfsutil.GetFile(apiClient, args[0], args[1], args[2], 0, 0, "", shard(), os.Stdout)
		}),
	}
	addShardFlags(getFile)

	inspectFile := &cobra.Command{
		Use:   "inspect-file repo-name commit-id path/to/file",
		Short: "Return info about a file.",
		Long:  "Return info about a file.",
		Run: pkgcobra.RunFixedArgs(3, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			fileInfo, err := pfsutil.InspectFile(apiClient, args[0], args[1], args[2], "", shard())
			if err != nil {
				return err
			}
			if fileInfo == nil {
				return fmt.Errorf("file %s not found", args[2])
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintFileInfoHeader(writer)
			pretty.PrintFileInfo(writer, fileInfo)
			return writer.Flush()
		}),
	}
	addShardFlags(inspectFile)

	listFile := &cobra.Command{
		Use:   "list-file repo-name commit-id path/to/dir",
		Short: "Return the files in a directory.",
		Long:  "Return the files in a directory.",
		Run: pkgcobra.RunBoundedArgs(pkgcobra.Bounds{Min: 2, Max: 3}, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			var path string
			if len(args) == 3 {
				path = args[2]
			}
			fileInfos, err := pfsutil.ListFile(apiClient, args[0], args[1], path, "", shard())
			if err != nil {
				return err
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			pretty.PrintFileInfoHeader(writer)
			for _, fileInfo := range fileInfos {
				pretty.PrintFileInfo(writer, fileInfo)
			}
			return writer.Flush()
		}),
	}
	addShardFlags(listFile)

	deleteFile := &cobra.Command{
		Use:   "delete-file repo-name commit-id path/to/file",
		Short: "Delete a file.",
		Long:  "Delete a file.",
		Run: pkgcobra.RunFixedArgs(2, func(args []string) error {
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			return pfsutil.DeleteFile(apiClient, args[0], args[1], args[2])
		}),
	}

	var mountPoint string
	mount := &cobra.Command{
		Use:   "mount [repo/commit:alias...]",
		Short: "Mount pfs locally.",
		Long:  "Mount pfs locally.",
		Run: pkgcobra.Run(func(args []string) error {
			//lion.SetLevel(lion.LevelDebug)
			apiClient, err := getAPIClient(address)
			if err != nil {
				return err
			}
			mounter := fuse.NewMounter(address, apiClient)
			return mounter.Mount(mountPoint, shard(), parseCommitMounts(args), nil)
		}),
	}
	mount.Flags().StringVarP(&mountPoint, "mount-point", "p", "/pfs", "root of mounted filesystem")
	addShardFlags(mount)

	var result []*cobra.Command
	result = append(result, repo)
	result = append(result, createRepo)
	result = append(result, inspectRepo)
	result = append(result, listRepo)
	result = append(result, deleteRepo)
	result = append(result, commit)
	result = append(result, startCommit)
	result = append(result, finishCommit)
	result = append(result, inspectCommit)
	result = append(result, listCommit)
	result = append(result, deleteCommit)
	result = append(result, file)
	result = append(result, putFile)
	result = append(result, getFile)
	result = append(result, inspectFile)
	result = append(result, listFile)
	result = append(result, deleteFile)
	result = append(result, mount)
	return result
}

func getAPIClient(address string) (pfs.APIClient, error) {
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return pfs.NewAPIClient(clientConn), nil
}

func getDriveAPIClient(address string) (pfs.BlockAPIClient, error) {
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return pfs.NewBlockAPIClient(clientConn), nil
}

func parseCommitMounts(args []string) []*fuse.CommitMount {
	var result []*fuse.CommitMount
	for _, arg := range args {
		commitMount := &fuse.CommitMount{Commit: pfsutil.NewCommit("", "")}
		repo, commitAlias := path.Split(arg)
		commitMount.Commit.Repo.Name = path.Clean(repo)
		split := strings.Split(commitAlias, ":")
		if len(split) > 0 {
			commitMount.Commit.Id = split[0]
		}
		if len(split) > 1 {
			commitMount.Alias = split[1]
		}
		result = append(result, commitMount)
	}
	return result
}
