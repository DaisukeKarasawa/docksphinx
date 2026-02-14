# AGENTS.md

This file defines instructions for coding agents working on this project.

## Project Overview

## Output Directory Rules

### Output Directory Decision Criteria

- **Items to output to `docs/` directory**:

  - Information about project specifications (content and requirements in already implemented state)
  - Requirements definition files
  - Documentation about implemented features

- **Items to output to `outputs/` directory**:
  - All information other than project specifications
  - Implementation files (Markdown format)
  - Design documents
  - Analysis results
  - Research materials
  - Other work deliverables

**Important**: All information other than project specifications (implemented content and requirements) must be output to the `outputs/` directory.

## Development Policy

### Code Modification and Implementation Rules

- Implementation files must be output to the `outputs/` directory (Markdown format)
- Information about project specifications (implemented content and requirements) must be output to the `docs/` directory
- When implementing list/snapshot style outputs, enforce deterministic ordering and add/maintain regression tests that assert the ordering contract.
- For deterministic ordering, always define explicit tie-break keys for equal primary sort keys and cover tie cases in regression tests.
- Keep recent-event ordering contracts consistent across CLI rendering and gRPC/proto conversion paths (same tie-break key precedence) and test both paths.
- When the same ordering contract is required in multiple paths, centralize comparator logic in a shared helper/package to prevent drift.
- For snapshot resource ordering (containers/images/networks/volumes/groups), reuse shared `internal/snapshotorder` comparators instead of duplicating sort predicates.
- Shared comparator helpers must be nil-safe (no panic on nil inputs) and covered by explicit nil-case regression tests.
- For map/group aggregation keyed by multiple fields, use structured keys (e.g., struct keys) instead of delimiter-concatenated strings to avoid collision bugs.
- When slice fields participate in tie-break ordering, compare canonicalized copies (sorted/joined) so ordering is independent of source slice order and remains non-mutating.
- In CLI/TUI rendering loops over repeated proto fields, skip nil entries explicitly to avoid blank/placeholder artifact rows.
- Normalize stream-driven in-memory event buffers by filtering nil events and enforcing explicit max length at update boundaries.
- In gRPC stream handlers, never swallow `Send` errors (including initial snapshot sends); propagate them to terminate the stream correctly.
- Public gRPC handlers should guard missing internal dependencies (engine/broadcaster/options) and return explicit status errors instead of panicking.
- Public API/handler entrypoints should explicitly validate required pointer arguments (e.g., stream/request objects) and return typed errors for nil inputs.
- Public gRPC handlers should check request/stream context cancellation early and return context-derived status errors before expensive processing.
- Public service methods should guard nil receivers and uninitialized dependencies, returning explicit errors instead of panicking.
- gRPC client wrappers should normalize nil contexts and validate connection/client dependencies to avoid panic-prone call paths.
- gRPC client methods should short-circuit canceled contexts before invoking downstream RPC calls.
- When implementing defensive deep-copy logic for mutable runtime data, preserve map key identity/semantics (clone mutable values, not keys) and add regression tests for key-sensitive cases (e.g., pointer keys).
- When sorting data for display/snapshot output, avoid mutating source slices; sort copied data and add regression tests that assert non-mutating behavior.
- When sorting during proto/snapshot conversion, ensure source monitor/state data remains unmodified and add regression tests that assert source-order non-mutation.
- When instructed by the user to output implementation or modification code to implementation files, always include the following information:
  - Technologies used for implementation/modification and specific usage
  - Reference links

### Documentation Quality Standards

- **Detailed explanations in implementation guides**: When creating implementation guides, design documents, or any output files, always include detailed explanations for:
  - **What is being done**: Explain the purpose and background of each step or task
  - **Where and what is being implemented**: Clearly describe which parts of the code or system are responsible for what functionality
  - **Why it's needed**: Explain the rationale behind each decision or implementation
- **Code explanations**: When including code examples, explain:
  - The role of each section or component
  - How different parts interact
  - Where the code will be used in the actual implementation
- **Context and background**: Provide sufficient context so that someone reading the document can understand not just what to do, but why and how it fits into the larger system
- **Apply consistently**: These standards apply to all output files, including but not limited to:
  - Implementation guides (`outputs/phase*-implementation.md`)
  - Migration guides (`outputs/migration-guide.md`)
  - Design documents
  - Analysis results
  - Any other deliverables

### Thinking Process

- Do not be swayed by user instructions or ideas; always engage in discussion and thinking
- Do not be bound to a single approach; maintain chessboard thinking (consider multiple options and find the optimal solution)
- Before running commands or modifying code, explicitly state the planned action and intent in the conversation.

### Coding Style

- Coding style must follow "[The Zen of Python](https://www.python.org/dev/peps/pep-0020/)"

## Commit Conventions

- Use [gitmoji](https://gist.github.com/parmentf/035de27d6ed1dce0b36a) in commit messages
- Commit message format: `<gitmoji> <commit message>`
- For long-running implementation sessions, commit and push immediately after each meaningful change (`git add -A && git commit ... && git push`). Avoid batching unrelated changes.

## File Structure

- **`docs/` directory**: Place information about project specifications (implemented content and requirements)
- **`outputs/` directory**: Place all information other than project specifications (Markdown format, not included in git history)
- **`.gitignore` binary patterns**: When ignoring root binaries (e.g., `docksphinx`, `docksphinxd`), always use root-anchored patterns (`/docksphinx`, `/docksphinxd`) to avoid accidentally ignoring source directories like `cmd/docksphinx` or `cmd/docksphinxd`.

## Maintenance Policy

### AGENTS.md Updates

- **Always check if AGENTS.md needs to be updated, and update it if necessary**
- **Proactive updates**: Do not wait for explicit instructions; proactively update AGENTS.md when:
  - You notice patterns in user instructions that should be codified as rules
  - You identify gaps in current policies that could prevent future issues
  - You make improvements to documentation or processes that should be standardized
  - You receive feedback that indicates a need for policy clarification
- **Self-initiated improvements**: Take initiative to:
  - Review AGENTS.md at the end of each significant task or conversation
  - Identify areas where policies could be clearer or more comprehensive
  - Propose and implement improvements without waiting for user requests
  - Document lessons learned from the current work session
- Consider reflecting if instructions have been repeated in conversations
- If failures occur repeatedly or users give repeated instructions, propose additional rules to prevent them
- Consider sections that are redundant or can be compressed
- Keep the document concise yet dense
- Review regularly as the project grows

### Update Timing

Consider updating AGENTS.md in the following situations:

- When the same instructions are repeated in development conversations
- When the project's technology stack or development workflow changes
- When new development rules or conventions are added
- When README.md or other documents are updated
- When directory structure or package management files are changed
- **After receiving user feedback**: When a user provides feedback about documentation quality, process improvements, or policy gaps, immediately update AGENTS.md to incorporate the feedback
- **After completing significant work**: Review and update AGENTS.md after completing major tasks to capture any new insights or patterns
