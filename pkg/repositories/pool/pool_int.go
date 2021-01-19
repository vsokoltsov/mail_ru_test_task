package pool

// Communication implements actions for pools
type Communication interface {
	StartWorkers()
	listenJobs(id int)
}
