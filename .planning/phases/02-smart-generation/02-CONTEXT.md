# Phase 2: Smart Generation - Context

**Gathered:** 2026-02-06
**Status:** Ready for planning

<domain>
## Phase Boundary

Ensuring generated training data has no duplicates and maintains clean boundaries between train/valid/test splits through exclusion logic, deduplication, and intelligent fill mechanism. This phase builds on the core pipeline to add quality controls.

</domain>

<decisions>
## Implementation Decisions

### Exclusion Strategy
- Pass full pairs (prompt + completion) to LLM in JSONL format
- If exclusion list exceeds context window limits, fail generation with clear error
- No token estimation or warnings needed — trust defaults are safe, fail fast if actual error
- Exclusion pairs formatted as JSONL (one pair per line, consistent with output format)

### Deduplication Approach
- Exact match only — both prompt and completion must be character-for-character identical
- Deduplicate per-document only (not at final merge time)
- Enforce exclusion across splits: valid can't contain train pairs, test can't contain train or valid pairs
- Keep first occurrence, remove subsequent duplicates (chronological order)
- Detailed logging: log each duplicate found with what was removed
- Include `--no-dedup` flag to disable deduplication for debugging LLM output quality
- If ALL pairs are duplicates, error and stop — indicates serious LLM or config problem
- Compare content only — ignore source document, metadata, timestamps

### Fill Mechanism
- Single LLM call for the shortage amount when dedup reduces counts below target
- If fill generation produces duplicates, retry up to 3 times (4 total attempts)
- Pass ALL existing pairs (train + valid + test) as exclusions during fill to prevent any duplicates
- After 3 failed retries, fail with error indicating LLM quality issue

### Claude's Discretion
- Exact error messages and formatting for context overflow failures
- Implementation of duplicate tracking data structures
- Logging format and verbosity details
- Retry backoff timing (if any)

</decisions>

<specifics>
## Specific Ideas

No specific requirements — open to standard approaches.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 02-smart-generation*
*Context gathered: 2026-02-06*
