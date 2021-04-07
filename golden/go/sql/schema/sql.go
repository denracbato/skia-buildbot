package schema

// Generated by //golden/go/sql/exporter/tosql
// DO NOT EDIT

const Schema = `CREATE TABLE IF NOT EXISTS Changelists (
  changelist_id STRING PRIMARY KEY,
  system STRING NOT NULL,
  status STRING NOT NULL,
  owner_email STRING NOT NULL,
  subject STRING NOT NULL,
  last_ingested_data TIMESTAMP WITH TIME ZONE NOT NULL,
  INDEX system_status_ingested_idx (system, status, last_ingested_data)
);
CREATE TABLE IF NOT EXISTS CommitsWithData (
  commit_id STRING PRIMARY KEY,
  tile_id INT4 NOT NULL
);
CREATE TABLE IF NOT EXISTS DiffMetrics (
  left_digest BYTES,
  right_digest BYTES,
  num_pixels_diff INT4 NOT NULL,
  percent_pixels_diff FLOAT4 NOT NULL,
  max_rgba_diffs INT2[] NOT NULL,
  max_channel_diff INT2 NOT NULL,
  combined_metric FLOAT4 NOT NULL,
  dimensions_differ BOOL NOT NULL,
  ts TIMESTAMP WITH TIME ZONE NOT NULL,
  PRIMARY KEY (left_digest, right_digest)
);
CREATE TABLE IF NOT EXISTS ExpectationDeltas (
  expectation_record_id UUID,
  grouping_id BYTES,
  digest BYTES,
  label_before CHAR NOT NULL,
  label_after CHAR NOT NULL,
  PRIMARY KEY (expectation_record_id, grouping_id, digest)
);
CREATE TABLE IF NOT EXISTS ExpectationRecords (
  expectation_record_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  branch_name STRING,
  user_name STRING NOT NULL,
  triage_time TIMESTAMP WITH TIME ZONE NOT NULL,
  num_changes INT4 NOT NULL
);
CREATE TABLE IF NOT EXISTS Expectations (
  grouping_id BYTES,
  digest BYTES,
  label CHAR NOT NULL,
  expectation_record_id UUID,
  PRIMARY KEY (grouping_id, digest)
);
CREATE TABLE IF NOT EXISTS GitCommits (
  git_hash STRING PRIMARY KEY,
  commit_id STRING NOT NULL,
  commit_time TIMESTAMP WITH TIME ZONE NOT NULL,
  author_email STRING NOT NULL,
  subject STRING NOT NULL,
  INDEX commit_idx (commit_id)
);
CREATE TABLE IF NOT EXISTS Groupings (
  grouping_id BYTES PRIMARY KEY,
  keys JSONB NOT NULL
);
CREATE TABLE IF NOT EXISTS IgnoreRules (
  ignore_rule_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  creator_email STRING NOT NULL,
  updated_email STRING NOT NULL,
  expires TIMESTAMP WITH TIME ZONE NOT NULL,
  note STRING,
  query JSONB NOT NULL
);
CREATE TABLE IF NOT EXISTS Options (
  options_id BYTES PRIMARY KEY,
  keys JSONB NOT NULL
);
CREATE TABLE IF NOT EXISTS Patchsets (
  patchset_id STRING PRIMARY KEY,
  system STRING NOT NULL,
  changelist_id STRING NOT NULL REFERENCES Changelists (changelist_id),
  ps_order INT2 NOT NULL,
  git_hash STRING NOT NULL,
  commented_on_cl BOOL NOT NULL,
  last_checked_if_comment_necessary TIMESTAMP WITH TIME ZONE NOT NULL,
  INDEX cl_order_idx (changelist_id, ps_order)
);
CREATE TABLE IF NOT EXISTS PrimaryBranchParams (
  tile_id INT4,
  key STRING,
  value STRING,
  PRIMARY KEY (tile_id, key, value)
);
CREATE TABLE IF NOT EXISTS ProblemImages (
  digest STRING PRIMARY KEY,
  num_errors INT2 NOT NULL,
  latest_error STRING NOT NULL,
  error_ts TIMESTAMP WITH TIME ZONE NOT NULL
);
CREATE TABLE IF NOT EXISTS SecondaryBranchExpectations (
  branch_name STRING,
  grouping_id BYTES,
  digest BYTES,
  label CHAR NOT NULL,
  expectation_record_id UUID NOT NULL,
  PRIMARY KEY (branch_name, grouping_id, digest)
);
CREATE TABLE IF NOT EXISTS SecondaryBranchParams (
  branch_name STRING,
  version_name STRING,
  key STRING,
  value STRING,
  PRIMARY KEY (branch_name, version_name, key, value)
);
CREATE TABLE IF NOT EXISTS SecondaryBranchValues (
  branch_name STRING,
  version_name STRING,
  secondary_branch_trace_id BYTES,
  digest BYTES NOT NULL,
  grouping_id BYTES NOT NULL,
  options_id BYTES NOT NULL,
  source_file_id BYTES NOT NULL,
  tryjob_id string,
  PRIMARY KEY (branch_name, version_name, secondary_branch_trace_id)
);
CREATE TABLE IF NOT EXISTS SourceFiles (
  source_file_id BYTES PRIMARY KEY,
  source_file STRING NOT NULL,
  last_ingested TIMESTAMP WITH TIME ZONE NOT NULL
);
CREATE TABLE IF NOT EXISTS TiledTraceDigests (
  trace_id BYTES,
  tile_id INT4,
  digest BYTES NOT NULL,
  grouping_id BYTES NOT NULL,
  PRIMARY KEY (trace_id, tile_id, digest),
  INDEX grouping_digest_idx (grouping_id, digest)
);
CREATE TABLE IF NOT EXISTS TrackingCommits (
  repo STRING PRIMARY KEY,
  last_git_hash STRING NOT NULL
);
CREATE TABLE IF NOT EXISTS TraceValues (
  shard INT2,
  trace_id BYTES,
  commit_id STRING,
  digest BYTES NOT NULL,
  grouping_id BYTES NOT NULL,
  options_id BYTES NOT NULL,
  source_file_id BYTES NOT NULL,
  PRIMARY KEY (shard, commit_id, trace_id)
);
CREATE TABLE IF NOT EXISTS Traces (
  trace_id BYTES PRIMARY KEY,
  corpus STRING AS (keys->>'source_type') STORED NOT NULL,
  grouping_id BYTES NOT NULL,
  keys JSONB NOT NULL,
  matches_any_ignore_rule BOOL,
  INDEX grouping_ignored_idx (grouping_id, matches_any_ignore_rule),
  INDEX ignored_grouping_idx (matches_any_ignore_rule, grouping_id),
  INVERTED INDEX keys_idx (keys)
);
CREATE TABLE IF NOT EXISTS Tryjobs (
  tryjob_id STRING PRIMARY KEY,
  system STRING NOT NULL,
  changelist_id STRING NOT NULL REFERENCES Changelists (changelist_id),
  patchset_id STRING NOT NULL REFERENCES Patchsets (patchset_id),
  display_name STRING NOT NULL,
  last_ingested_data TIMESTAMP WITH TIME ZONE NOT NULL
);
CREATE TABLE IF NOT EXISTS ValuesAtHead (
  trace_id BYTES PRIMARY KEY,
  most_recent_commit_id STRING NOT NULL,
  digest BYTES NOT NULL,
  options_id BYTES NOT NULL,
  grouping_id BYTES NOT NULL,
  corpus STRING AS (keys->>'source_type') STORED NOT NULL,
  keys JSONB NOT NULL,
  matches_any_ignore_rule BOOL,
  INDEX ignored_grouping_idx (matches_any_ignore_rule, grouping_id)
);
CREATE TABLE IF NOT EXISTS DeprecatedIngestedFiles (
  source_file_id BYTES PRIMARY KEY,
  source_file STRING NOT NULL,
  last_ingested TIMESTAMP WITH TIME ZONE NOT NULL
);
CREATE TABLE IF NOT EXISTS DeprecatedExpectationUndos (
  id SERIAL PRIMARY KEY,
  expectation_id STRING NOT NULL,
  user_id STRING NOT NULL,
  ts TIMESTAMP WITH TIME ZONE NOT NULL
);
`
