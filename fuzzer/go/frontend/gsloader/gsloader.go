package gsloader

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/skia-dev/glog"
	"go.skia.org/infra/fuzzer/go/common"
	"go.skia.org/infra/fuzzer/go/config"
	"go.skia.org/infra/fuzzer/go/fuzz"
	"go.skia.org/infra/fuzzer/go/fuzzcache"
	"go.skia.org/infra/go/gs"
	"google.golang.org/cloud/storage"
)

// LoadFromBoltDB loads the fuzz.FuzzReport from FuzzReportCache associated with the given hash.
// The FuzzReport is first put into the staging fuzz cache, and then into the current.
// If a cache for the commit does not exist, or there are other problems with the retrieval,
// an error is returned.
func LoadFromBoltDB(cache *fuzzcache.FuzzReportCache) error {
	glog.Infof("Looking into cache for revision %s", config.FrontEnd.SkiaVersion.Hash)
	for _, category := range common.FUZZ_CATEGORIES {
		if staging, err := cache.LoadTree(category, config.FrontEnd.SkiaVersion.Hash); err != nil {
			return fmt.Errorf("Problem decoding existing from bolt db: %s", err)
		} else {
			fuzz.SetStaging(category, *staging)
			glog.Infof("Successfully loaded %s fuzzes from bolt db cache", category)
		}
	}
	fuzz.StagingToCurrent()
	return nil
}

// GSLoader is a struct that handles downloading fuzzes from Google Storage.
type GSLoader struct {
	storageClient *storage.Client
	Cache         *fuzzcache.FuzzReportCache

	// completedCounter is the number of fuzzes that have been downloaded from GCS, used for logging.
	completedCounter int32
}

// New creates a GSLoader and returns it.
func New(s *storage.Client, c *fuzzcache.FuzzReportCache) *GSLoader {
	return &GSLoader{
		storageClient: s,
		Cache:         c,
	}
}

// LoadFreshFromGoogleStorage pulls all fuzzes out of GCS and loads them into memory.
// The "fresh" in the name refers to the fact that all other loaded fuzzes (if any)
// are written over, including in the cache.
// Upon completion, the full results are cached to a boltDB instance and moved from staging
// to the current copy.
func (g *GSLoader) LoadFreshFromGoogleStorage() error {
	revision := config.FrontEnd.SkiaVersion.Hash
	fuzz.ClearStaging()
	fuzzNames := make([]string, 0, 100)
	for _, cat := range common.FUZZ_CATEGORIES {
		badPath := fmt.Sprintf("%s/%s/bad", cat, revision)
		reports, err := g.getBinaryReportsFromGS(badPath, nil)
		if err != nil {
			return err
		}
		n := 0
		for report := range reports {
			fuzz.NewFuzzFound(cat, report)
			fuzzNames = append(fuzzNames, report.FuzzName)
			n++
		}
		glog.Infof("%d bad fuzzes freshly loaded from gs://%s/%s", n, config.GS.Bucket, badPath)
	}

	fuzz.StagingToCurrent()
	for _, category := range common.FUZZ_CATEGORIES {
		if err := g.Cache.StoreTree(fuzz.StagingCopy(category), category, revision); err != nil {
			glog.Errorf("Problem storing category %s to boltDB: %s", category, err)
		}
	}
	return g.Cache.StoreFuzzNames(fuzzNames, revision)
}

// LoadBinaryFuzzesFromGoogleStorage pulls all fuzzes out of GCS that are on the given whitelist
// and loads them into memory (as staging).  After loading them, it updates the cache
// and moves them from staging to the current copy.
func (g *GSLoader) LoadBinaryFuzzesFromGoogleStorage(whitelist []string) error {
	revision := config.FrontEnd.SkiaVersion.Hash
	fuzz.StagingFromCurrent()
	sort.Strings(whitelist)

	fuzzNames := make([]string, 0, 100)
	for _, cat := range common.FUZZ_CATEGORIES {
		badPath := fmt.Sprintf("%s/%s/bad", cat, revision)
		reports, err := g.getBinaryReportsFromGS(badPath, whitelist)
		if err != nil {
			return err
		}
		n := 0
		for report := range reports {
			fuzz.NewFuzzFound(cat, report)
			fuzzNames = append(fuzzNames, report.FuzzName)
			n++
		}
		glog.Infof("%d bad fuzzes freshly loaded from gs://%s/%s", n, config.GS.Bucket, badPath)
	}
	fuzz.StagingToCurrent()

	oldBinaryFuzzNames, err := g.Cache.LoadFuzzNames(revision)
	if err != nil {
		glog.Warningf("Could not read old binary fuzz names from cache.  Continuing...", err)
		oldBinaryFuzzNames = []string{}
	}
	for _, category := range common.FUZZ_CATEGORIES {
		if err := g.Cache.StoreTree(fuzz.StagingCopy(category), category, revision); err != nil {
			glog.Errorf("Problem storing category %s to boltDB: %s", category, err)
		}
	}
	return g.Cache.StoreFuzzNames(append(oldBinaryFuzzNames, whitelist...), revision)
}

// A fuzzPackage contains all the information about a fuzz, mostly the paths to the files that
// need to be downloaded.
type fuzzPackage struct {
	FuzzName        string
	DebugDumpName   string
	DebugErrName    string
	ReleaseDumpName string
	ReleaseErrName  string
}

// getBinaryReportsFromGS pulls all files in baseFolder from the skia-fuzzer bucket and
// groups them by fuzz.  It parses these groups of files into a BinaryFuzzReport and returns
// a channel through whcih all reports generated in this way will be streamed.
// The channel will be closed when all reports are done being sent.
func (g *GSLoader) getBinaryReportsFromGS(baseFolder string, whitelist []string) (<-chan fuzz.FuzzReport, error) {
	reports := make(chan fuzz.FuzzReport, 10000)

	fuzzPackages, err := g.fetchFuzzPackages(baseFolder)
	if err != nil {
		close(reports)
		return reports, err
	}

	toDownload := make(chan fuzzPackage, len(fuzzPackages))
	g.completedCounter = 0

	var wg sync.WaitGroup
	for i := 0; i < config.FrontEnd.NumDownloadProcesses; i++ {
		wg.Add(1)
		go g.download(toDownload, reports, &wg)
	}

	for _, d := range fuzzPackages {
		if whitelist != nil {
			name := d.FuzzName
			if i := sort.SearchStrings(whitelist, name); i < len(whitelist) && whitelist[i] == name {
				// is on the whitelist
				toDownload <- d
			}
		} else {
			// no white list
			toDownload <- d
		}
	}
	close(toDownload)

	go func() {
		wg.Wait()
		close(reports)
	}()

	return reports, nil
}

// fetchFuzzPackages scans for all fuzzes in the given folder and returns a
// slice of all of the metadata for each fuzz, as a fuzz package.  It returns
// error if it cannot access Google Storage.
func (g *GSLoader) fetchFuzzPackages(baseFolder string) (fuzzPackages []fuzzPackage, err error) {

	fuzzNames, err := common.GetAllFuzzNamesInFolder(g.storageClient, baseFolder)
	if err != nil {
		return nil, fmt.Errorf("Problem getting fuzz packages from %s: %s", baseFolder, err)
	}
	for _, fuzzName := range fuzzNames {
		prefix := fmt.Sprintf("%s/%s/%s", baseFolder, fuzzName, fuzzName)
		fuzzPackages = append(fuzzPackages, fuzzPackage{
			FuzzName:        fuzzName,
			DebugDumpName:   fmt.Sprintf("%s_debug.dump", prefix),
			DebugErrName:    fmt.Sprintf("%s_debug.err", prefix),
			ReleaseDumpName: fmt.Sprintf("%s_release.dump", prefix),
			ReleaseErrName:  fmt.Sprintf("%s_release.err", prefix),
		})
	}
	return fuzzPackages, nil
}

// emptyStringOnError returns a string of the passed in bytes or empty string if err is nil.
func emptyStringOnError(b []byte, err error) string {
	if err != nil {
		glog.Warningf("Ignoring error when fetching file contents: %v", err)
		return ""
	}
	return string(b)
}

// download waits for fuzzPackages to appear on the toDownload channel and then downloads
// the four pieces of the package.  It then parses them into a BinaryFuzzReport and sends
// the binary to the passed in channel.  When there is no more work to be done, this function.
// returns and writes out true to the done channel.
func (g *GSLoader) download(toDownload <-chan fuzzPackage, reports chan<- fuzz.FuzzReport, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range toDownload {
		debugDump := emptyStringOnError(gs.FileContentsFromGS(g.storageClient, config.GS.Bucket, job.DebugDumpName))
		debugErr := emptyStringOnError(gs.FileContentsFromGS(g.storageClient, config.GS.Bucket, job.DebugErrName))
		releaseDump := emptyStringOnError(gs.FileContentsFromGS(g.storageClient, config.GS.Bucket, job.ReleaseDumpName))
		releaseErr := emptyStringOnError(gs.FileContentsFromGS(g.storageClient, config.GS.Bucket, job.ReleaseErrName))
		reports <- fuzz.ParseReport(job.FuzzName, debugDump, debugErr, releaseDump, releaseErr)
		atomic.AddInt32(&g.completedCounter, 1)
		if g.completedCounter%100 == 0 {
			glog.Infof("%d fuzzes downloaded", g.completedCounter)
		}
	}
}
