package workload

import (
	"fmt"
	"io"
	"math/rand"

	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/ppsutil"
)

func RunWorkload(
	pfsClient pfs.APIClient,
	ppsClient pps.APIClient,
	rand *rand.Rand,
	size int,
) error {
	worker := newWorker(rand)
	for i := 0; i < size; i++ {
		if err := worker.work(pfsClient, ppsClient); err != nil {
			return err
		}
	}
	for _, job := range worker.startedJobs {
		jobInfo, err := ppsClient.InspectJob(
			context.Background(),
			&pps.InspectJobRequest{
				Job:        job,
				BlockState: true,
			},
		)
		if err != nil {
			return err
		}
		if jobInfo.State != pps.JobState_JOB_STATE_SUCCESS {
			return fmt.Errorf("job %s failed", job.Id)
		}
	}
	return nil
}

type worker struct {
	repos       []*pfs.Repo
	finished    []*pfs.Commit
	started     []*pfs.Commit
	files       []*pfs.File
	startedJobs []*pps.Job
	jobs        []*pps.Job
	pipelines   []*pps.Pipeline
	rand        *rand.Rand
}

func newWorker(rand *rand.Rand) *worker {
	return &worker{
		rand: rand,
	}
}

const (
	repo     float64 = .02
	commit           = .1
	file             = .9
	job              = .98
	pipeline         = 1.0
)

const maxStartedCommits = 6
const maxStartedJobs = 6

func (w *worker) work(pfsClient pfs.APIClient, ppsClient pps.APIClient) error {
	opt := w.rand.Float64()
	switch {
	case opt < repo:
		repoName := w.randString(10)
		if err := pfsutil.CreateRepo(pfsClient, repoName); err != nil {
			return err
		}
		w.repos = append(w.repos, &pfs.Repo{Name: repoName})
		commit, err := pfsutil.StartCommit(pfsClient, repoName, "")
		if err != nil {
			return err
		}
		w.started = append(w.started, commit)
	case opt < commit:
		if len(w.started) >= maxStartedCommits || len(w.finished) == 0 {
			if len(w.started) == 0 {
				return nil
			}
			i := w.rand.Intn(len(w.started))
			commit := w.started[i]
			if err := pfsutil.FinishCommit(pfsClient, commit.Repo.Name, commit.Id); err != nil {
				return err
			}
			w.started = append(w.started[:i], w.started[i+1:]...)
			w.finished = append(w.finished, commit)
		} else {
			if len(w.finished) == 0 {
				return nil
			}
			commit := w.finished[w.rand.Intn(len(w.finished))]
			commit, err := pfsutil.StartCommit(pfsClient, commit.Repo.Name, commit.Id)
			if err != nil {
				return err
			}
			w.started = append(w.started, commit)
		}
	case opt < file:
		if len(w.started) == 0 {
			return nil
		}
		commit := w.started[w.rand.Intn(len(w.started))]
		if _, err := pfsutil.PutFile(pfsClient, commit.Repo.Name, commit.Id, w.randString(10), 0, w.reader()); err != nil {
			return err
		}
	case opt < job:
		if len(w.startedJobs) >= maxStartedJobs {
			job := w.startedJobs[0]
			w.startedJobs = w.startedJobs[1:]
			jobInfo, err := ppsClient.InspectJob(
				context.Background(),
				&pps.InspectJobRequest{
					Job:        job,
					BlockState: true,
				},
			)
			if err != nil {
				return err
			}
			if jobInfo.State != pps.JobState_JOB_STATE_SUCCESS {
				return fmt.Errorf("job %s failed", job.Id)
			}
			w.jobs = append(w.jobs, job)
		} else {
			if len(w.finished) == 0 {
				return nil
			}
			inputs := [5]string{}
			var jobInputs []*pps.JobInput
			repoSet := make(map[string]bool)
			for i := range inputs {
				commit := w.finished[w.rand.Intn(len(w.finished))]
				if _, ok := repoSet[commit.Repo.Name]; ok {
					continue
				}
				repoSet[commit.Repo.Name] = true
				inputs[i] = commit.Repo.Name
				jobInputs = append(jobInputs, &pps.JobInput{Commit: commit})
			}
			var parentJobID string
			if len(w.jobs) > 0 {
				parentJobID = w.jobs[w.rand.Intn(len(w.jobs))].Id
			}
			outFilename := w.randString(10)
			job, err := ppsutil.CreateJob(
				ppsClient,
				"",
				[]string{"bash"},
				w.grepCmd(inputs, outFilename),
				1,
				jobInputs,
				parentJobID,
			)
			if err != nil {
				return err
			}
			w.startedJobs = append(w.startedJobs, job)
		}
	case opt < pipeline:
		if len(w.repos) == 0 {
			return nil
		}
		inputs := [5]string{}
		var pipelineInputs []*pps.PipelineInput
		repoSet := make(map[string]bool)
		for i := range inputs {
			repo := w.repos[w.rand.Intn(len(w.repos))]
			if _, ok := repoSet[repo.Name]; ok {
				continue
			}
			repoSet[repo.Name] = true
			inputs[i] = repo.Name
			pipelineInputs = append(pipelineInputs, &pps.PipelineInput{Repo: repo})
		}
		pipelineName := w.randString(10)
		outFilename := w.randString(10)
		if err := ppsutil.CreatePipeline(
			ppsClient,
			pipelineName,
			"",
			[]string{"bash"},
			w.grepCmd(inputs, outFilename),
			1,
			pipelineInputs,
		); err != nil {
			return err
		}
		w.pipelines = append(w.pipelines, ppsutil.NewPipeline(pipelineName))
	}
	return nil
}

const letters = "abcdefghijklmnopqrstuvwxyz"
const lettersAndSpaces = "abcdefghijklmnopqrstuvwxyz      "

func (w *worker) randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[w.rand.Intn(len(letters))]
	}
	return string(b)
}

type reader struct {
	rand  *rand.Rand
	bytes int
}

func NewReader(rand *rand.Rand, bytes int) io.Reader {
	return &reader{
		rand:  rand,
		bytes: bytes,
	}
}

func (r *reader) Read(p []byte) (int, error) {
	for i := range p {
		if i%128 == 127 {
			p[i] = '\n'
		} else {
			p[i] = lettersAndSpaces[r.rand.Intn(len(lettersAndSpaces))]
		}
	}
	p[len(p)-1] = '\n'
	r.bytes -= len(p)
	if r.bytes <= 0 {
		return len(p), io.EOF
	}
	return len(p), nil
}

func (w *worker) reader() io.Reader {
	return NewReader(w.rand, 1000)
}

func (w *worker) grepCmd(inputs [5]string, outFilename string) []string {
	return []string{
		fmt.Sprintf(
			"grep %s /pfs/{%s,%s,%s,%s,%s}/* >/pfs/out/%s; true",
			w.randString(4),
			inputs[0],
			inputs[1],
			inputs[2],
			inputs[3],
			inputs[4],
			outFilename,
		),
	}
}
