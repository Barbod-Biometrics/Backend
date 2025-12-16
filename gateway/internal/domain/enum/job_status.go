package enum

type JobStatus string

const (
	JobStatusPending JobStatus = "pending"
	JobStatusDone    JobStatus = "done"
	JobStatusFailed  JobStatus = "failed"
)
