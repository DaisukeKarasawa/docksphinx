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
