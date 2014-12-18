package main

// ingest is the command line tool for pulling performance data from Google
// Storage and putting in Tiles. See the code in go/ingester for details on how
// ingestion is done.

import (
	"flag"
	"net/http"
	"strings"
	"time"

	"github.com/golang/glog"
	"skia.googlesource.com/buildbot.git/go/auth"
	"skia.googlesource.com/buildbot.git/go/common"
	"skia.googlesource.com/buildbot.git/go/gitinfo"
	"skia.googlesource.com/buildbot.git/perf/go/config"
	"skia.googlesource.com/buildbot.git/perf/go/db"
	"skia.googlesource.com/buildbot.git/perf/go/goldingester"
	"skia.googlesource.com/buildbot.git/perf/go/ingester"
	"skia.googlesource.com/buildbot.git/perf/go/trybot"
)

// flags
var (
	tileDir        = flag.String("tile_dir", "/tmp/tileStore2/", "Path where tiles will be placed.")
	gitRepoDir     = flag.String("git_repo_dir", "../../../skia", "Directory location for the Skia repo.")
	runEvery       = flag.Duration("run_every", 5*time.Minute, "How often the ingester should pull data from Google Storage.")
	runTrybotEvery = flag.Duration("run_trybot_every", 1*time.Minute, "How often the ingester to pull trybot data from Google Storage.")
	run            = flag.String("run", "nano,nano-trybot,golden", "A comma separated list of ingesters to run.")
	graphiteServer = flag.String("graphite_server", "skia-monitoring:2003", "Where is Graphite metrics ingestion server running.")
	doOauth        = flag.Bool("oauth", true, "Run through the OAuth 2.0 flow on startup, otherwise use a GCE service account.")
	oauthCacheFile = flag.String("oauth_cache_file", "/home/perf/google_storage_token.data", "Path to the file where to cache cache the oauth credentials.")
	local          = flag.Bool("local", false, "Running locally if true. As opposed to in production.")
	nCommits       = flag.Int("n_commits", 100, "Minimum number of commits that should be ingested.")
	minDays        = flag.Int("min_days", 7, "Minimum number of days that should be covered by the ingested commits.")
)

// Process is what each ingestion is wrapped up behind.
//
// A Process is expected to never return, and should be called as a Go routine.
type Process func()

// NewIngestionProcess creates a Process for ingesting data.
func NewIngestionProcess(git *gitinfo.GitInfo, tileDir, datasetName string, ri ingester.ResultIngester, gsDir string, every time.Duration, nCommits int, minDuration time.Duration, metricName string) Process {
	i, err := ingester.NewIngester(git, tileDir, datasetName, ri, nCommits, minDuration, gsDir, metricName)
	if err != nil {
		glog.Fatalf("Failed to create Ingester: %s", err)
	}

	glog.Infof("Starting %s ingester. Run every %s. Fetch from %s ", datasetName, every.String(), gsDir)

	// oneStep is a single round of ingestion.
	oneStep := func() {
		glog.Infof("Running ingester: %s", datasetName)
		err := i.Update()
		if err != nil {
			glog.Error(err)
		}
		glog.Infof("Finished running ingester: %s", datasetName)
	}

	return func() {
		oneStep()
		for _ = range time.Tick(every) {
			oneStep()
		}
	}
}

func main() {
	common.InitWithMetrics("ingest", *graphiteServer)

	// Initialize the database. We might not need the oauth dialog if it fails.
	db.Init(db.ProdDatabaseConfig(*local))

	var client *http.Client
	var err error
	if *doOauth {
		config := auth.DefaultOAuthConfig(*oauthCacheFile)
		client, err = auth.RunFlow(config)
		if err != nil {
			glog.Fatalf("Failed to auth: %s", err)
		}
	} else {
		client = nil
		// Add back service account access here when it's fixed.
	}

	ingester.Init(client)

	git, err := gitinfo.NewGitInfo(*gitRepoDir, true, false)
	if err != nil {
		glog.Fatalf("Failed loading Git info: %s\n", err)
	}

	// Get duration equivalent to the number of days.
	minDuration := 24 * time.Hour * time.Duration(*minDays)

	// ingesters is a list of all the types of ingestion we can do.
	ingesters := map[string]Process{
		"nano":        NewIngestionProcess(git, *tileDir, config.DATASET_NANO, ingester.NewNanoBenchIngester(), "nano-json-v1", *runEvery, *nCommits, minDuration, "nano-ingest"),
		"nano-trybot": NewIngestionProcess(git, *tileDir, config.DATASET_NANO_TRYBOT, trybot.NewTrybotResultIngester(), "trybot/nano-json-v1", *runTrybotEvery, *nCommits, minDuration, "nano-trybot"),
		"golden":      NewIngestionProcess(git, *tileDir, config.DATASET_GOLDEN, goldingester.NewGoldIngester(), "dm-json-v1", *runEvery, *nCommits, minDuration, "golden-ingest"),
	}

	for _, name := range strings.Split(*run, ",") {
		glog.Infof("Process name: %s", name)
		if process, ok := ingesters[name]; ok {
			go process()
		} else {
			glog.Fatalf("Not a valid ingester name: %s", name)
		}
	}

	select {}
}
