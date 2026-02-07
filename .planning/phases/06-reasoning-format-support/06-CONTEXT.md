# Phase 6: Reasoning Format Support - Context

**Gathered:** 2026-02-07
**Status:** Ready for planning

<domain>
## Phase Boundary

Generate chain-of-thought training data with step-by-step reasoning traces from documentation. Users run `--format reasoning` and get JSONL with structured reasoning output. This is its own distinct format (not an extension of chat format). Deduplication, exclusion, per-doc files, merge, and all existing features work with reasoning format.

</domain>

<decisions>
## Implementation Decisions

### Output structure
- Reasoning is its own `--format reasoning` — distinct from completion and chat, with its own schema
- Claude's Discretion: exact JSONL schema (e.g., thinking tags, separate fields, step arrays — research current ecosystem)
- Claude's Discretion: whether reasoning trace and final answer are separate fields or integrated
- Claude's Discretion: whether to include metadata (step count, difficulty) or keep minimal

### Reasoning style
- Claude's Discretion: reasoning approach (analytical breakdown, problem-solving, Socratic, or mix)
- Claude's Discretion: verbosity level (concise vs detailed, or matched to complexity)
- Claude's Discretion: whether to include self-correction/backtracking in traces
- Claude's Discretion: tone (formal vs casual thinking language)

### Question design
- Questions should be fundamentally different from completion/chat — specifically designed to demand multi-step reasoning ("Why does X lead to Y?", "Compare A and B", "What would happen if...")
- Questions and reasoning must be document-grounded only — no extrapolation beyond what's in the source text
- Skip generating reasoning pairs for very simple/factual content that doesn't warrant multi-step thinking — fewer but higher quality pairs
- Claude's Discretion: specific question type mix (why/how, comparisons, what-if, implications)

### Compatibility targets
- Support multiple reasoning format variants via a flag (e.g., `--reasoning-format`)
- Default to whatever format is most popular in the open-source reasoning model ecosystem (likely DeepSeek-R1 style)
- Claude's Discretion: which specific variants to support (research current ecosystem)
- Claude's Discretion: whether to validate output against known schemas or trust the template

### Claude's Discretion
- Exact JSONL schema design (research best practices)
- Reasoning style choices (approach, verbosity, tone, self-correction)
- Which reasoning format variants to implement and their names
- Output validation approach
- Template design for reasoning-specific prompts

</decisions>

<specifics>
## Specific Ideas

- Questions should specifically demand reasoning — not the same questions as completion/chat with reasoning tacked on
- Quality over quantity — skip content that's too simple for meaningful reasoning traces
- Multiple format variants give users flexibility for different fine-tuning targets

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 06-reasoning-format-support*
*Context gathered: 2026-02-07*
