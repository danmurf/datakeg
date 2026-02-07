# Phase 5: Chat Format Support - Context

**Gathered:** 2026-02-07
**Status:** Ready for planning

<domain>
## Phase Boundary

Generate chat-style training data with messages/roles JSONL output via `--format chat`. Single-turn conversations (one user message, one assistant response) in OpenAI-compatible format. Existing completion format remains unchanged. Reasoning format is a separate phase.

</domain>

<decisions>
## Implementation Decisions

### Conversation structure
- Single-turn only: one user message + one assistant response per entry
- No multi-turn conversations
- System message is optional, enabled via `--system-message` flag
- Claude's Discretion: whether system message content is user-provided or auto-generated from document context

### Template instructions
- Mix of question styles: factual, clarifying, how-to, and conceptual questions for diversity
- User messages are context-free (no reference to "the docs" or "the guide")
- Assistant responses must mirror the voice, tone, dialect, and conversational style of the source document — as if the document's author is speaking
- Claude's Discretion: whether to explicitly instruct the LLM to match document style or let it infer naturally
- Separate templates per split (train, valid, test) — same as completion format

### Output shape
- OpenAI-compatible messages format: `{"messages": [{"role": "user", "content": "..."}, {"role": "assistant", "content": "..."}]}`
- When system message included: `{"messages": [{"role": "system", "content": "..."}, {"role": "user", "content": "..."}, {"role": "assistant", "content": "..."}]}`
- Messages only — no metadata fields (no source doc name, no timestamps)
- Claude's Discretion: system message position in array (standard is first)
- Per-document files use same naming pattern as completion format (docname_train.jsonl)

### Format flag behavior
- One format per run: `--format chat` or `--format completion`, not both
- Claude's Discretion: default format when `--format` not specified (likely completion for backward compatibility)
- Claude's Discretion: merge subcommand format detection (auto-detect vs require flag)
- Claude's Discretion: behavior when format mismatch with existing files in output directory

### Claude's Discretion
- System message implementation (user-provided vs auto-generated)
- Template wording for style matching
- Default format value
- Merge format detection approach
- Format coexistence/overwrite behavior
- Deduplication adaptation for messages format

</decisions>

<specifics>
## Specific Ideas

- "The response should be written in the same style as the document. It should be assumed that the document is written as if a person spoke it, so the same conversational style, dialect, etc. should be used." — This is a core requirement for chat template design.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 05-chat-format-support*
*Context gathered: 2026-02-07*
