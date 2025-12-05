package enum

type JobStatus string

const (
	JobStatusPending  JobStatus = "pending"
	JobStatusVerified JobStatus = "verified"
	JobStatusRejected JobStatus = "rejected"
)
