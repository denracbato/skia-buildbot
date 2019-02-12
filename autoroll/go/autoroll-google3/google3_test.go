package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
	"go.skia.org/infra/autoroll/go/recent_rolls"
	"go.skia.org/infra/autoroll/go/roller"
	"go.skia.org/infra/go/autoroll"
	"go.skia.org/infra/go/deepequal"
	"go.skia.org/infra/go/ds"
	"go.skia.org/infra/go/ds/testutil"
	"go.skia.org/infra/go/git"
	git_testutils "go.skia.org/infra/go/git/testutils"
	gitiles_testutils "go.skia.org/infra/go/gitiles/testutils"
	"go.skia.org/infra/go/jsonutils"
	"go.skia.org/infra/go/mockhttpclient"
	"go.skia.org/infra/go/testutils"
)

func setup(t *testing.T) (context.Context, *AutoRoller, *git_testutils.GitBuilder, *gitiles_testutils.MockRepo, func()) {
	testutils.LargeTest(t)
	ctx := context.Background()
	testutil.InitDatastore(t, ds.KIND_AUTOROLL_ROLL, ds.KIND_AUTOROLL_STATUS)
	gb := git_testutils.GitInit(t, ctx)
	urlmock := mockhttpclient.NewURLMock()
	mockChild := gitiles_testutils.NewMockRepo(t, gb.RepoUrl(), git.GitDir(gb.Dir()), urlmock)
	a, err := NewAutoRoller(ctx, "", &roller.AutoRollerConfig{
		ChildName: "test-child",
		Google3RepoManager: &roller.Google3FakeRepoManagerConfig{
			ChildBranch: "master",
			ChildRepo:   gb.RepoUrl(),
		},
		ParentName: "test-parent",
		RollerName: "test-roller",
	}, urlmock.Client())
	assert.NoError(t, err)
	a.Start(ctx, time.Second, time.Second)
	return ctx, a, gb, mockChild, func() {
		gb.Cleanup()
	}
}

func makeIssue(num int64, commit string) *autoroll.AutoRollIssue {
	now := time.Now().UTC()
	return &autoroll.AutoRollIssue{
		Closed:      false,
		Committed:   false,
		CommitQueue: true,
		Created:     now,
		Issue:       num,
		Modified:    now,
		Patchsets:   nil,
		Result:      autoroll.ROLL_RESULT_IN_PROGRESS,
		RollingFrom: "prevrev",
		RollingTo:   commit,
		Subject:     fmt.Sprintf("%d", num),
		TryResults: []*autoroll.TryResult{
			&autoroll.TryResult{
				Builder:  "Test Summary",
				Category: autoroll.TRYBOT_CATEGORY_CQ,
				Created:  now,
				Result:   "",
				Status:   autoroll.TRYBOT_STATUS_STARTED,
				Url:      "http://example.com/",
			},
		},
	}
}

func closeIssue(issue *autoroll.AutoRollIssue, result string) {
	issue.Closed = true
	issue.CommitQueue = false
	issue.Modified = time.Now().UTC()
	issue.Result = result
	issue.TryResults[0].Status = autoroll.TRYBOT_STATUS_COMPLETED
	issue.TryResults[0].Result = autoroll.TRYBOT_RESULT_FAILURE
	if result == autoroll.ROLL_RESULT_SUCCESS {
		issue.Committed = true
		issue.TryResults[0].Result = autoroll.TRYBOT_RESULT_SUCCESS
	}
}

func TestStatus(t *testing.T) {
	ctx, a, gb, mockChild, cleanup := setup(t)
	defer cleanup()

	commits := []string{gb.CommitGen(ctx, "a.txt")}

	issue1 := makeIssue(1, commits[0])
	assert.NoError(t, a.AddOrUpdateIssue(ctx, issue1, http.MethodPost))
	closeIssue(issue1, autoroll.ROLL_RESULT_SUCCESS)
	assert.NoError(t, a.AddOrUpdateIssue(ctx, issue1, http.MethodPut))

	// Ensure that repo update occurs when updating status.
	commits = append(commits, gb.CommitGen(ctx, "a.txt"))

	mockChild.MockGetCommit(ctx, "master")
	mockChild.MockLog(ctx, commits[0], commits[1])
	assert.NoError(t, a.UpdateStatus(ctx, "", true))
	status := a.status.Get()
	assert.Equal(t, 0, status.NumFailedRolls)
	assert.Equal(t, 1, status.NumNotRolledCommits)
	assert.Equal(t, issue1.RollingTo, status.LastRollRev)
	assert.Nil(t, status.CurrentRoll)
	deepequal.AssertDeepEqual(t, issue1, status.LastRoll)
	deepequal.AssertDeepEqual(t, []*autoroll.AutoRollIssue{issue1}, status.Recent)

	// Ensure that repo update occurs when adding an issue.
	commits = append(commits, gb.CommitGen(ctx, "a.txt"))

	issue2 := makeIssue(2, commits[2])
	assert.NoError(t, a.AddOrUpdateIssue(ctx, issue2, http.MethodPost))
	closeIssue(issue2, autoroll.ROLL_RESULT_FAILURE)
	assert.NoError(t, a.AddOrUpdateIssue(ctx, issue2, http.MethodPut))

	issue3 := makeIssue(3, commits[2])
	assert.NoError(t, a.AddOrUpdateIssue(ctx, issue3, http.MethodPost))
	closeIssue(issue3, autoroll.ROLL_RESULT_FAILURE)
	assert.NoError(t, a.AddOrUpdateIssue(ctx, issue3, http.MethodPut))

	issue4 := makeIssue(4, commits[2])
	assert.NoError(t, a.AddOrUpdateIssue(ctx, issue4, http.MethodPost))

	recent := []*autoroll.AutoRollIssue{issue4, issue3, issue2, issue1}
	mockChild.MockGetCommit(ctx, "master")
	mockChild.MockLog(ctx, commits[0], commits[2])
	assert.NoError(t, a.UpdateStatus(ctx, "error message", false))
	status = a.status.Get()
	assert.Equal(t, 2, status.NumFailedRolls)
	assert.Equal(t, 2, status.NumNotRolledCommits)
	assert.Equal(t, issue1.RollingTo, status.LastRollRev)
	assert.Equal(t, "error message", status.Error)
	deepequal.AssertDeepEqual(t, issue4, status.CurrentRoll)
	deepequal.AssertDeepEqual(t, issue3, status.LastRoll)
	deepequal.AssertDeepEqual(t, recent, status.Recent)

	// Test preserving error.
	mockChild.MockGetCommit(ctx, "master")
	mockChild.MockLog(ctx, commits[0], commits[2])
	assert.NoError(t, a.UpdateStatus(ctx, "", true))
	status = a.status.Get()
	assert.Equal(t, "error message", status.Error)

	// Overflow recent_rolls.RECENT_ROLLS_LENGTH.
	closeIssue(issue4, autoroll.ROLL_RESULT_FAILURE)
	assert.NoError(t, a.AddOrUpdateIssue(ctx, issue4, http.MethodPut))
	recent = []*autoroll.AutoRollIssue{issue4, issue3}
	// Rolls 3 and 4 failed, so we need 5 thru recent_rolls.RECENT_ROLLS_LENGTH + 3 to also fail for
	// overflow.
	for i := int64(5); i < recent_rolls.RECENT_ROLLS_LENGTH+3; i++ {
		issueI := makeIssue(i, commits[2])
		assert.NoError(t, a.AddOrUpdateIssue(ctx, issueI, http.MethodPost))
		closeIssue(issueI, autoroll.ROLL_RESULT_FAILURE)
		assert.NoError(t, a.AddOrUpdateIssue(ctx, issueI, http.MethodPut))
		recent = append([]*autoroll.AutoRollIssue{issueI}, recent...)
	}
	mockChild.MockGetCommit(ctx, "master")
	mockChild.MockLog(ctx, commits[0], commits[2])
	assert.NoError(t, a.UpdateStatus(ctx, "error message", false))
	status = a.status.Get()
	assert.Equal(t, recent_rolls.RECENT_ROLLS_LENGTH+1, status.NumFailedRolls)
	assert.Equal(t, 2, status.NumNotRolledCommits)
	assert.Equal(t, issue1.RollingTo, status.LastRollRev)
	assert.Equal(t, "error message", status.Error)
	assert.Nil(t, status.CurrentRoll)
	deepequal.AssertDeepEqual(t, recent[0], status.LastRoll)
	deepequal.AssertDeepEqual(t, recent, status.Recent)
}

func TestAddOrUpdateIssue(t *testing.T) {
	ctx, a, gb, mockChild, cleanup := setup(t)
	defer cleanup()

	commits := []string{gb.CommitGen(ctx, "a.txt"), gb.CommitGen(ctx, "a.txt"), gb.CommitGen(ctx, "a.txt")}

	issue1 := makeIssue(1, commits[0])
	assert.NoError(t, a.AddOrUpdateIssue(ctx, issue1, http.MethodPost))
	closeIssue(issue1, autoroll.ROLL_RESULT_SUCCESS)
	assert.NoError(t, a.AddOrUpdateIssue(ctx, issue1, http.MethodPut))

	// Test adding an issue that is already closed.
	issue2 := makeIssue(2, commits[1])
	closeIssue(issue2, autoroll.ROLL_RESULT_SUCCESS)
	assert.NoError(t, a.AddOrUpdateIssue(ctx, issue2, http.MethodPut))
	mockChild.MockGetCommit(ctx, "master")
	mockChild.MockLog(ctx, commits[1], commits[2])
	assert.NoError(t, a.UpdateStatus(ctx, "", true))
	deepequal.AssertDeepEqual(t, []*autoroll.AutoRollIssue{issue2, issue1}, a.status.Get().Recent)

	// Test adding a two issues without closing the first one.
	issue3 := makeIssue(3, commits[2])
	assert.NoError(t, a.AddOrUpdateIssue(ctx, issue3, http.MethodPost))
	issue4 := makeIssue(4, commits[2])
	assert.NoError(t, a.AddOrUpdateIssue(ctx, issue4, http.MethodPost))
	mockChild.MockGetCommit(ctx, "master")
	mockChild.MockLog(ctx, commits[1], commits[2])
	assert.NoError(t, a.UpdateStatus(ctx, "", true))
	issue3.Closed = true
	issue3.Result = autoroll.ROLL_RESULT_FAILURE
	deepequal.AssertDeepEqual(t, []*autoroll.AutoRollIssue{issue4, issue3, issue2, issue1}, a.status.Get().Recent)

	// Test both situations at the same time.
	issue5 := makeIssue(5, commits[2])
	closeIssue(issue5, autoroll.ROLL_RESULT_SUCCESS)
	assert.NoError(t, a.AddOrUpdateIssue(ctx, issue5, http.MethodPut))
	mockChild.MockGetCommit(ctx, "master")
	mockChild.MockLog(ctx, commits[2], commits[2])
	assert.NoError(t, a.UpdateStatus(ctx, "", true))
	issue4.Closed = true
	issue4.Result = autoroll.ROLL_RESULT_FAILURE
	deepequal.AssertDeepEqual(t, []*autoroll.AutoRollIssue{issue5, issue4, issue3, issue2, issue1}, a.status.Get().Recent)
}

func makeRoll(now time.Time) Roll {
	return Roll{
		ChangeListNumber: 1,
		Closed:           false,
		Created:          jsonutils.Time(now),
		Modified:         jsonutils.Time(now),
		Result:           autoroll.ROLL_RESULT_IN_PROGRESS,
		RollingTo:        "rev",
		RollingFrom:      "prevrev",
		Subject:          "1",
		Submitted:        false,
		TestSummaryUrl:   "http://example.com/",
	}
}

func TestRollAsIssue(t *testing.T) {
	testutils.SmallTest(t)

	expected := makeIssue(1, "rev")
	now := expected.Created
	roll := makeRoll(now)

	actual, err := roll.AsIssue()
	assert.NoError(t, err)
	deepequal.AssertDeepEqual(t, expected, actual)

	roll.TestSummaryUrl = ""
	savedTryResults := expected.TryResults
	expected.TryResults = []*autoroll.TryResult{}
	actual, err = roll.AsIssue()
	assert.NoError(t, err)
	deepequal.AssertDeepEqual(t, expected, actual)

	roll.Closed = true
	expected.Closed = true
	expected.CommitQueue = false
	roll.Result = autoroll.ROLL_RESULT_FAILURE
	expected.Result = autoroll.ROLL_RESULT_FAILURE
	roll.TestSummaryUrl = "http://example.com/"
	expected.TryResults = savedTryResults
	expected.TryResults[0].Result = autoroll.TRYBOT_RESULT_FAILURE
	expected.TryResults[0].Status = autoroll.TRYBOT_STATUS_COMPLETED
	actual, err = roll.AsIssue()
	assert.NoError(t, err)
	deepequal.AssertDeepEqual(t, expected, actual)

	roll.Submitted = true
	roll.Result = autoroll.ROLL_RESULT_SUCCESS
	expected.Committed = true
	expected.Result = autoroll.ROLL_RESULT_SUCCESS
	expected.TryResults[0].Result = autoroll.TRYBOT_RESULT_SUCCESS
	actual, err = roll.AsIssue()
	assert.NoError(t, err)
	deepequal.AssertDeepEqual(t, expected, actual)

	roll = makeRoll(now)
	roll.Created = jsonutils.Time{}
	_, err = roll.AsIssue()
	assert.EqualError(t, err, "Missing parameter.")

	roll = makeRoll(now)
	roll.RollingFrom = ""
	_, err = roll.AsIssue()
	assert.EqualError(t, err, "Missing parameter.")

	roll = makeRoll(now)
	roll.RollingTo = ""
	_, err = roll.AsIssue()
	assert.EqualError(t, err, "Missing parameter.")

	roll = makeRoll(now)
	roll.Closed = true
	_, err = roll.AsIssue()
	assert.EqualError(t, err, "Inconsistent parameters: result must be set.")

	roll = makeRoll(now)
	roll.Submitted = true
	_, err = roll.AsIssue()
	assert.EqualError(t, err, "Inconsistent parameters: submitted but not closed.")

	roll = makeRoll(now)
	roll.Result = ""
	_, err = roll.AsIssue()
	assert.EqualError(t, err, "Unsupported value for result.")

	roll = makeRoll(now)
	roll.TestSummaryUrl = ":http//example.com"
	_, err = roll.AsIssue()
	assert.EqualError(t, err, "Invalid testSummaryUrl parameter.")
}
