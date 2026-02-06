# Phase 2: Smart Generation - Context

**Gathered:** 2026-02-06
**Status:** Ready for planning

<domain>
## Phase Boundary

Add deduplication and exclusion logic to ensure generated training pairs are unique within and across splits (train/valid/test). This phase enhances the existing pipeline from Phase 1 without changing core generation or I/O.

</domain>

<decisions>
## Implementation Decisions

### Deduplication Strategy
- **Match criteria:** Exact character-for-character match (case-sensitive)
- **Comparison scope:** Both prompt AND completion must match to be considered a duplicate
- **Deduplication boundary:** Per-document only - check for duplicates within each document's generated pairs, allow duplicates across different documents
- **Tracking mechanism:** In-memory hash set that resets between documents

### Exclusion Mechanics
- **Generation order:** Sequential - generate train first, then valid (excluding train), then test (excluding train+valid)
- **LLM awareness:** Pass previous split pairs as context to LLM so it can avoid creating duplicates (larger prompts but better avoidance)
- **Underfilled splits:** Auto-regenerate missing pairs if exclusion filtering leaves a split below target count
- **Regeneration limit:** Maximum 3 regeneration attempts per split before accepting the achieved count

### Backfill Behavior
- **Target calculation:** Generate exactly the remaining gap (if 7 pairs exist and 10 needed, generate 3 more)
- **Backfill timing:** Per-document basis - after generating all splits for a document, check and backfill immediately
- **Backfill prompting:** Modified prompt that includes previously generated pairs and asks LLM to create different ones
- **Exhaustion handling:** If max attempts reached without filling gap, log warning and continue processing remaining documents

### Quality Validation
- **Validation checks:**
  - Empty field check: Reject pairs where prompt or completion is empty/whitespace-only
  - JSON format check: Verify each pair has exactly 'prompt' and 'completion' keys
- **Validation timing:** Both stages - quick validation after parsing, then full validation after deduplication
- **Invalid pair handling:** Silently discard invalid pairs without logging
- **Backfill on validation failures:** Yes - treat validation failures like deduplication removals and backfill to reach target count

### Claude's Discretion
- Hash function choice for the in-memory hash set
- Exact prompt template modifications for passing context to LLM
- Debug logging implementation details
- Error message wording

</decisions>

<specifics>
## Specific Ideas

No specific requirements — open to standard approaches

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 02-smart-generation*
*Context gathered: 2026-02-06*
