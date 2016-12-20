package tryjobs

import (
	"sort"
	"sync"
	"time"

	"go.skia.org/infra/go/sklog"

	"go.skia.org/infra/task_scheduler/go/db"
	"go.skia.org/infra/task_scheduler/go/window"
)

// jobCache is a struct which provides more useful views of Jobs than the
// database itself can.
type jobCache struct {
	activeTryJobs map[string]*db.Job
	db            db.JobDB
	mtx           sync.RWMutex
	queryId       string
	timeWindow    *window.Window
}

// GetActiveTryJobs returns all active try Jobs. A try Job is
// considered to be active if it has a non-zero Buildbucket lease key.
func (c *jobCache) GetActiveTryJobs() ([]*db.Job, error) {
	c.mtx.RLock()
	defer c.mtx.RUnlock()

	rv := make([]*db.Job, 0, len(c.activeTryJobs))
	for _, j := range c.activeTryJobs {
		rv = append(rv, j.Copy())
	}
	// Sort to maintain deterministic testing.
	sort.Sort(db.JobSlice(rv))
	return rv, nil
}

// expireJobs removes data from c where getJobTimestamp or getRevisionTimestamp
// is before start. Assumes the caller holds a lock. This is a helper for
// expireAndUpdate.
func (c *jobCache) expireJobs(start time.Time) {
	expiredUnfinishedCount := 0
	for _, job := range c.activeTryJobs {
		if job.Created.Before(start) {
			delete(c.activeTryJobs, job.Id)
			if !job.Done() {
				expiredUnfinishedCount++
			}
		}
	}
	if expiredUnfinishedCount > 0 {
		sklog.Infof("Expired %d unfinished jobs created before %s.", expiredUnfinishedCount, start)
	}
}

// insertOrUpdateJob inserts the new/updated job into the cache. Assumes the
// caller holds a lock. This is a helper for expireAndUpdate.
func (c *jobCache) insertOrUpdateJob(job *db.Job) {
	// Active try jobs.
	if job.BuildbucketLeaseKey == 0 {
		delete(c.activeTryJobs, job.Id)
	} else {
		c.activeTryJobs[job.Id] = job
	}
}

// expireAndUpdate removes Jobs before start from the cache and inserts the
// new/updated jobs into the cache. Assumes the caller holds a lock.
func (c *jobCache) expireAndUpdate(start time.Time, jobs []*db.Job) {
	c.expireJobs(start)
	for _, job := range jobs {
		if job.Created.Before(start) {
			sklog.Warningf("Updated job %s after expired. getJobTimestamp returned %s. %#v", job.Id, job.Created, job)
		} else {
			c.insertOrUpdateJob(job.Copy())
		}
	}
}

// reset re-initializes c. Assumes the caller holds a lock.
func (c *jobCache) reset() error {
	if c.queryId != "" {
		c.db.StopTrackingModifiedJobs(c.queryId)
	}
	queryId, err := c.db.StartTrackingModifiedJobs()
	if err != nil {
		return err
	}
	now := time.Now()
	start := c.timeWindow.Start()
	sklog.Infof("Reading Jobs from %s to %s.", start, now)
	jobs, err := c.db.GetJobsFromDateRange(start, now)
	if err != nil {
		c.db.StopTrackingModifiedJobs(queryId)
		return err
	}
	c.activeTryJobs = map[string]*db.Job{}
	c.queryId = queryId
	c.expireAndUpdate(start, jobs)
	return nil
}

// See documentation for JobCache interface.
func (c *jobCache) Update() error {
	newJobs, err := c.db.GetModifiedJobs(c.queryId)
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if db.IsUnknownId(err) {
		sklog.Warningf("Connection to db lost; re-initializing cache from scratch.")
		if err := c.reset(); err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}
	c.expireAndUpdate(c.timeWindow.Start(), newJobs)
	return nil
}

// newJobCache returns a local cache which provides more convenient views of
// job data than the database can provide.
func newJobCache(db db.JobDB, timeWindow *window.Window) (*jobCache, error) {
	tc := &jobCache{
		db:         db,
		timeWindow: timeWindow,
	}
	if err := tc.reset(); err != nil {
		return nil, err
	}
	return tc, nil
}
