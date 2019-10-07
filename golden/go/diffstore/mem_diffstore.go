package diffstore

import (
	"bytes"
	"fmt"
	"image"
	"math"
	"net/http"
	"runtime"
	"strings"
	"sync"

	"go.skia.org/infra/go/gcs"
	"go.skia.org/infra/go/metrics2"
	"go.skia.org/infra/go/rtcache"
	"go.skia.org/infra/go/skerr"
	"go.skia.org/infra/go/sklog"
	"go.skia.org/infra/go/util"
	"go.skia.org/infra/golden/go/diff"
	"go.skia.org/infra/golden/go/diffstore/common"
	"go.skia.org/infra/golden/go/diffstore/failurestore"
	"go.skia.org/infra/golden/go/diffstore/mapper"
	"go.skia.org/infra/golden/go/diffstore/metricsstore"
	"go.skia.org/infra/golden/go/types"
	"go.skia.org/infra/golden/go/validation"
)

const (
	// DEFAULT_IMG_DIR_NAME is the directory where the images are stored.
	DEFAULT_IMG_DIR_NAME = "images"

	// DEFAULT_DIFFIMG_DIR_NAME is the directory where the diff images are stored.
	DEFAULT_DIFFIMG_DIR_NAME = "diffs"

	// DEFAULT_GCS_IMG_DIR_NAME is the default image directory in GCS.
	DEFAULT_GCS_IMG_DIR_NAME = "dm-images-v1"

	// BYTES_PER_IMAGE is the estimated number of bytes an uncompressed images consumes.
	// Used to conservatively estimate the maximum number of items in the cache.
	BYTES_PER_IMAGE = 1024 * 1024

	// BYTES_PER_DIFF_METRIC is the estimated number of bytes per diff metric.
	// Used to conservatively estimate the maximum number of items in the cache.
	BYTES_PER_DIFF_METRIC = 100

	// maxGoRoutines is the maximum number of Go-routines we allow in MemDiffStore.
	// This was determined empirically: we had instances running with 500k go-routines without problems.
	maxGoRoutines = 500000
)

// MemDiffStore implements the diff.DiffStore interface.
type MemDiffStore struct {
	// diffMetricsCache caches and calculates diff metrics.
	diffMetricsCache rtcache.ReadThroughCache

	// imgLoader fetches and caches images.
	imgLoader *ImageLoader

	// metricsStore persists diff metrics.
	metricsStore metricsstore.MetricsStore

	// wg is used to synchronize background operations like saving files. Used for testing.
	wg sync.WaitGroup

	// mapper contains various functions for creating image IDs, paths and diff metrics.
	mapper mapper.Mapper

	// maxGoRoutinesCh is a buffered channel that is used to limit the number of goroutines for diffing.
	maxGoRoutinesCh chan bool
}

// NewMemDiffStore returns a new instance of MemDiffStore.
// 'gigs' is the approximate number of gigs to use for caching. This is not the
// exact amount memory that will be used, but a tuning parameter to increase
// or decrease memory used. If 'gigs' is 0 nothing will be cached in memory.
// If diffFn is not specified, the diff.DefaultDiffFn will be used. If codec is
// not specified, a JSON codec for the diff.DiffMetrics struct will be used.
// If mapper is not specified, GoldIDPathMapper will be used.
func NewMemDiffStore(client gcs.GCSClient, gsImageBaseDir string, gigs int, m mapper.Mapper, mStore metricsstore.MetricsStore, fStore failurestore.FailureStore) (diff.DiffStore, error) {
	imageCacheCount, diffCacheCount := getCacheCounts(gigs)

	// Set up image retrieval, caching and serving.
	sklog.Debugf("Creating img loader with cache of size %d", imageCacheCount)
	imgLoader, err := NewImgLoader(client, fStore, gsImageBaseDir, imageCacheCount, m)
	if err != nil {
		return nil, skerr.Fmt("Could not create img loader %s", err)
	}

	ret := &MemDiffStore{
		imgLoader:       imgLoader,
		metricsStore:    mStore,
		mapper:          m,
		maxGoRoutinesCh: make(chan bool, maxGoRoutines),
	}

	sklog.Debugf("Creating diffMetricsCache of size %d", diffCacheCount)
	if ret.diffMetricsCache, err = rtcache.New(ret.diffMetricsWorker, diffCacheCount, runtime.NumCPU()); err != nil {
		return nil, skerr.Fmt("Could not create diffMetricsCache: %s", err)
	}
	return ret, nil
}

// WarmDigests fetches images based on the given list of digests. It does
// not cache the images but makes sure they are downloaded from GCS.
func (d *MemDiffStore) WarmDigests(priority int64, digests types.DigestSlice, sync bool) {
	missingDigests := make(types.DigestSlice, 0, len(digests))
	for _, digest := range digests {
		if !d.imgLoader.Contains(digest) {
			missingDigests = append(missingDigests, digest)
		}
	}
	if len(missingDigests) > 0 {
		d.imgLoader.Warm(rtcache.PriorityTimeCombined(priority), missingDigests, sync)
	}
}

// WarmDiffs puts the diff metrics for the cross product of leftDigests x rightDigests into the cache for the
// given diff metric and with the given priority. This means if there are multiple subsets of the digests
// with varying priority (ignored vs "regular") we can call this multiple times.
func (d *MemDiffStore) WarmDiffs(priority int64, leftDigests types.DigestSlice, rightDigests types.DigestSlice) {
	priority = rtcache.PriorityTimeCombined(priority)
	diffIDs := getDiffIds(leftDigests, rightDigests)
	sklog.Infof("Warming %d diffs", len(diffIDs))
	d.wg.Add(len(diffIDs))
	for _, id := range diffIDs {
		d.maxGoRoutinesCh <- true

		go func(diffId string) {
			defer func() {
				d.wg.Done()
				<-d.maxGoRoutinesCh
			}()

			if err := d.diffMetricsCache.Warm(priority, diffId); err != nil {
				sklog.Errorf("Unable to warm diff %s. Got error: %s", diffId, err)
			}
		}(id)
	}
}

func (d *MemDiffStore) sync() {
	d.wg.Wait()
}

// See DiffStore interface.
func (d *MemDiffStore) Get(priority int64, mainDigest types.Digest, rightDigests types.DigestSlice) (map[types.Digest]interface{}, error) {
	if mainDigest == "" {
		return nil, fmt.Errorf("Received empty dMain digest.")
	}

	diffMap := make(map[types.Digest]interface{}, len(rightDigests))
	var wg sync.WaitGroup
	var mutex sync.Mutex
	for _, right := range rightDigests {
		// Don't compare the digest to itself.
		if mainDigest != right {
			wg.Add(1)
			d.maxGoRoutinesCh <- true

			go func(right types.Digest) {
				defer func() {
					wg.Done()
					<-d.maxGoRoutinesCh
				}()
				id := common.DiffID(mainDigest, right)
				ret, err := d.diffMetricsCache.Get(priority, id)
				if err != nil {
					sklog.Errorf("Unable to calculate diff for %s. Got error: %s", id, err)
					return
				}
				mutex.Lock()
				defer mutex.Unlock()
				diffMap[right] = ret
			}(right)
		}
	}
	wg.Wait()
	return diffMap, nil
}

// UnavailableDigests implements the DiffStore interface.
func (m *MemDiffStore) UnavailableDigests() map[types.Digest]*diff.DigestFailure {
	return m.imgLoader.failureStore.UnavailableDigests()
}

// PurgeDigests implements the DiffStore interface.
func (m *MemDiffStore) PurgeDigests(digests types.DigestSlice, purgeGCS bool) error {
	// We remove the given digests from the various places where they might
	// be stored. None of the purge steps should return an error if the digests
	// related information is missing. So any error indicates a bigger problem in the
	// underlying system, i.e.issues with disk etc., that needs to be investigated
	// by hand. Since we remove the digests from the failureStore last, we will
	// not loose the vital information of what digests failed in the first place.

	// Remove the images from, the image cache, disk and GCS if necessary.
	if err := m.imgLoader.PurgeImages(digests, purgeGCS); err != nil {
		return err
	}

	// Remove the diff metrics from the cache if they exist.
	digestSet := make(types.DigestSet, len(digests))
	for _, d := range digests {
		digestSet[d] = true
	}
	removeKeys := make([]string, 0, len(digests))
	for _, key := range m.diffMetricsCache.Keys() {
		d1, d2 := common.SplitDiffID(key)
		if digestSet[d1] || digestSet[d2] {
			removeKeys = append(removeKeys, key)
		}
	}
	m.diffMetricsCache.Remove(removeKeys)

	if err := m.metricsStore.PurgeDiffMetrics(digests); err != nil {
		return err
	}

	return m.imgLoader.failureStore.PurgeDigestFailures(digests)
}

// ImageHandler implements the DiffStore interface.
func (m *MemDiffStore) ImageHandler(urlPrefix string) (http.Handler, error) {
	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		// Go's image package has no color profile support and we convert to 8-bit NRGBA to diff, but
		// our source images may have embedded color profiles and be up to 16-bit. So we must at least
		// take care to serve the original .pngs unaltered.
		//
		// TODO(lovisolo): Diff in NRGBA64?
		// TODO(lovisolo): Make sure each pair of images is in the same color space before diffing?
		//                 (They probably are today but it'd be a good sanity check to make sure.)

		dotExt := "." + common.IMG_EXTENSION
		urlPath := r.URL.Path
		sklog.Debugf("diffstore handling %s", urlPath)
		idx := strings.Index(urlPath, "/")
		if idx == -1 {
			noCacheNotFound(w, r)
			return
		}
		dir := urlPath[:idx]

		// Limit the requests to directories with the images and diff images.
		if (dir != DEFAULT_DIFFIMG_DIR_NAME) && (dir != DEFAULT_IMG_DIR_NAME) {
			noCacheNotFound(w, r)
			return
		}

		// Get the file that was requested and verify that it's a valid PNG file.
		file := urlPath[idx+1:]
		if (len(file) <= len(dotExt)) || (!strings.HasSuffix(file, dotExt)) {
			noCacheNotFound(w, r)
			return
		}

		// Trim the image extension to get the image ID.
		imgID := urlPath[idx+1 : len(urlPath)-len(dotExt)]
		imgDigest := types.Digest(imgID)

		// Cache images for 12 hours.
		w.Header().Set("Cache-Control", "public, max-age=43200")

		if dir == DEFAULT_IMG_DIR_NAME {
			// Validate the requested image ID.
			if !validation.IsValidDigest(imgID) {
				noCacheNotFound(w, r)
				return
			}

			// Retrieve the image from the in-memory cache.
			imgs, err := m.imgLoader.Get(diff.PRIORITY_NOW, types.DigestSlice{imgDigest})
			if err != nil {
				sklog.Errorf("Error retrieving digest: %s", imgID)
				noCacheNotFound(w, r)
				return
			}

			// Write output image to the http.ResponseWriter. Content-Type is set automatically based on
			// the first 512 bytes of written data. See docs for ResponseWriter.Write() for details.
			if _, err := w.Write(imgs[0]); err != nil {
				sklog.Errorf("Error writing image to http.ResponseWriter: %s", err)
				noCacheNotFound(w, r)
			}
		} else {
			// Validate the requested diff image ID.
			if !validation.IsValidDiffImgID(imgID) {
				noCacheNotFound(w, r)
				return
			}

			// Extract the left and right image digests.
			leftImgDigest, rightImgDigest := common.SplitDiffID(imgID)

			// Retrieve the images from the in-memory cache.
			imgs, err := m.imgLoader.Get(diff.PRIORITY_NOW, types.DigestSlice{leftImgDigest, rightImgDigest})
			if err != nil {
				sklog.Errorf("Error retrieving digests to compute diff: %s", imgID)
				noCacheNotFound(w, r)
				return
			}

			// Decode images.
			leftImg, rightImg, err := decodeImages(imgs[0], imgs[1])
			if err != nil {
				sklog.Errorf("Error decoding left/right images for diff %s: %s", imgID, err)
				noCacheNotFound(w, r)
				return
			}

			// Compute the diff image.
			_, diffImg := diff.PixelDiff(leftImg, rightImg)

			// Write output image to the http.ResponseWriter. Content-Type is set automatically based on
			// the first 512 bytes of written data. See docs for ResponseWriter.Write() for details.
			//
			// The encoding step below does not take color profiles into account. This is fine since both
			// the left and right images used to compute the diff are in the same color space, and also
			// because the resulting diff image is just a visual approximation of the differences between
			// the left and right images.
			if err := common.EncodeImg(w, diffImg); err != nil {
				sklog.Errorf("Error encoding diff image: %s", err)
				noCacheNotFound(w, r)
			}
		}
	}

	sklog.Infof("Created diffstore")

	// The above function relies on the URL prefix being stripped.
	return http.StripPrefix(urlPrefix, http.HandlerFunc(handlerFunc)), nil
}

// decodeImages takes two images (left and right) represented as byte slices, decodes them and
// returns two image.NRGBA pointers with the results.
func decodeImages(leftBytes, rightBytes []byte) (leftImg, rightImg *image.NRGBA, err error) {
	var leftErr, rightErr error

	var wg sync.WaitGroup
	wg.Add(2)

	// Decode left image.
	go func() {
		defer wg.Done()
		leftImg, leftErr = common.DecodeImg(bytes.NewReader(leftBytes))
	}()

	// Decode right image.
	go func() {
		defer wg.Done()
		rightImg, rightErr = common.DecodeImg(bytes.NewReader(rightBytes))
	}()

	wg.Wait()

	if leftErr != nil {
		return nil, nil, skerr.Wrap(leftErr)
	}
	if rightErr != nil {
		return nil, nil, skerr.Wrap(rightErr)
	}
	return leftImg, rightImg, nil
}

// noCacheNotFound disables caching and returns a 404.
func noCacheNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	http.NotFound(w, r)
}

// diffMetricsWorker calculates the diff if it's not in the cache.
func (d *MemDiffStore) diffMetricsWorker(priority int64, id string) (interface{}, error) {
	defer metrics2.FuncTimer().Stop()
	leftDigest, rightDigest := common.SplitDiffID(id)

	// Load it from disk cache if necessary.
	if dm, err := d.metricsStore.LoadDiffMetrics(id); err != nil {
		sklog.Errorf("Error trying to load diff metric: %s", err)
	} else if dm != nil {
		return dm, nil
	}

	// Get the images.
	imgs, err := d.imgLoader.Get(priority, types.DigestSlice{leftDigest, rightDigest})
	if err != nil {
		return nil, skerr.Wrapf(err, "Failed retrieving the following digests from ImageLoader: %s, %s.", leftDigest, rightDigest)
	}

	// Decode images. We are guaranteed to have two images at this point.
	leftImg, rightImg, err := decodeImages(imgs[0], imgs[1])
	if err != nil {
		return nil, skerr.Wrapf(err, "Error decoding left/right images while computing diff metrics for %s", id)
	}

	// We are guaranteed to have two images at this point.
	diffMetrics := d.mapper.DiffFn(leftImg, rightImg)

	// Save the diffMetrics.
	d.saveDiffMetricsAsync(id, leftDigest, rightDigest, diffMetrics.(*diff.DiffMetrics))
	return diffMetrics, nil
}

// saveDiffMetricsAsync saves the given diff metrics to disk asynchronously.
func (d *MemDiffStore) saveDiffMetricsAsync(diffID string, leftDigest, rightDigest types.Digest, diffMetrics *diff.DiffMetrics) {
	d.wg.Add(1)
	d.maxGoRoutinesCh <- true
	go func() {
		defer func() {
			d.wg.Done()
			<-d.maxGoRoutinesCh
		}()
		if err := d.metricsStore.SaveDiffMetrics(diffID, diffMetrics); err != nil {
			sklog.Errorf("Error saving diff metric: %s", err)
		}
	}()
}

// Returns all combinations of leftDigests and rightDigests using the given
// DiffID function of the DiffStore's mapper.
func getDiffIds(leftDigests, rightDigests types.DigestSlice) []string {
	// reminder: diffIds are "digest1:digest2"
	diffIDsSet := make(util.StringSet, len(leftDigests)*len(rightDigests))
	for _, left := range leftDigests {
		for _, right := range rightDigests {
			if left != right {
				diffIDsSet[common.DiffID(left, right)] = true
			}
		}
	}
	return diffIDsSet.Keys()
}

// getCacheCounts returns the number of images and diff metrics to cache
// based on the number of GiB provided.
// We are assume that we want to store x images and x^2 diffmetrics and
// solve the corresponding quadratic equation.
func getCacheCounts(gigs int) (int, int) {
	if gigs <= 0 {
		return 0, 0
	}

	// We are looking for x (number of images we can cache) where x is found by solving
	// a * x^2 + b * x = c
	diffSize := float64(BYTES_PER_DIFF_METRIC)             // a
	imgSize := float64(BYTES_PER_IMAGE)                    // b
	bytesGig := float64(uint64(gigs) * 1024 * 1024 * 1024) // c
	// To solve, use the quadratic formula and round to an int
	imgCount := int((-imgSize + math.Sqrt(imgSize*imgSize+4*diffSize*bytesGig)) / (2 * diffSize))
	return imgCount, imgCount * imgCount
}
