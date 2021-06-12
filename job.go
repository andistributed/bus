package bus

import (
	"context"
	"errors"
	"sync"
)

type Job interface {
	Execute(ctx context.Context, params string) (string, error)
}

type JobSessions struct {
	lock     sync.RWMutex
	sessions map[string]JobSession // {snapshot.Id:JobSession}
}

func NewJobSessions() *JobSessions {
	return &JobSessions{
		sessions: make(map[string]JobSession),
	}
}

func (s *JobSessions) Add(snapshotId string, job Job) JobSession {
	s.lock.Lock()
	session, ok := s.sessions[snapshotId]
	if ok {
		session.Cancel()
	}
	session = NewJobSession(job)
	s.sessions[snapshotId] = session
	s.lock.Unlock()
	return session
}

func (s *JobSessions) Done(snapshotId string) {
	s.lock.Lock()
	delete(s.sessions, snapshotId)
	s.lock.Unlock()
}

func (s *JobSessions) Cancel(snapshotId string) {
	s.lock.Lock()
	session, ok := s.sessions[snapshotId]
	if ok {
		session.Cancel()
		delete(s.sessions, snapshotId)
	}
	s.lock.Unlock()
}

type JobSession interface {
	Job() Job
	Execute(params string) (string, error)
	Cancel()
	Context() context.Context
}

var ErrJobCancelled = errors.New("the job was canceled")

type JobWithCancel struct {
	job    Job
	ctx    context.Context
	cancel context.CancelFunc
}

func (j JobWithCancel) Execute(params string) (r string, err error) {
	for {
		select {
		case <-j.ctx.Done():
			err = ErrJobCancelled
			return
		default:
			r, err = j.job.Execute(j.ctx, params)
			j.cancel()
			return
		}
	}
}

func (job JobWithCancel) Cancel() {
	job.cancel()
}

func (job JobWithCancel) Context() context.Context {
	return job.ctx
}

func (job JobWithCancel) Job() Job {
	return job.job
}

func NewJobSession(job Job) JobSession {
	ctx, cancel := context.WithCancel(context.Background())
	return &JobWithCancel{
		job:    job,
		ctx:    ctx,
		cancel: cancel,
	}
}
