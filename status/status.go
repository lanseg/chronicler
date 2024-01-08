package status

import (
	cm "github.com/lanseg/golang-commons/common"
)

type JobState struct {
}

type ProgressUpdate struct {
}

type Job interface {
	Start() error
	UpdateProgress(u *ProgressUpdate) error
	GetState() *JobState
	Stop() error
}

type NoopJob struct {
	Job

	logger *cm.Logger
}

func (j *NoopJob) Start() error {
	j.logger.Debugf("Job start requested")
	return nil
}

func (j *NoopJob) UpdateProgress(p *ProgressUpdate) error {
	j.logger.Debugf("Job progress requested")
	return nil
}

func (j *NoopJob) GetState() *JobState {
	j.logger.Debugf("Job status requested")
	return &JobState{}
}

func (j *NoopJob) Stop() error {
	j.logger.Debugf("Job stop requested")
	return nil
}

func NewNoopJob(name string) Job {
	return &NoopJob{
		logger: cm.NewLogger("Noop job " + name),
	}
}

type Status interface {
	NewJob(name string) Job
	GetLobs() []Job
}

type NoopStatus struct {
	Status

	logger *cm.Logger
}

func (ns *NoopStatus) NewJob(name string) Job {
	return NewNoopJob(name)
}

func (ns *NoopStatus) GetJobs() []Job {
	return []Job{}
}

func NewNoopStatus() Status {
	return &NoopStatus{
		logger: cm.NewLogger("NoopStatus"),
	}
}
