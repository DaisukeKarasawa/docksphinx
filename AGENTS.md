# AGENTS.md

This file defines instructions for coding agents working on this project.

## Project Overview

This project's basic policy is to output documents as requirements definition files, rather than directly modifying or implementing code.

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

- **Direct code modification and implementation is prohibited**
- Do not directly modify code files; instead, append them as implementation files
- Implementation files must be output to the `outputs/` directory (Markdown format)
- Information about project specifications (implemented content and requirements) must be output to the `docs/` directory
- When instructed by the user to output implementation or modification code to implementation files, always include the following information:
  - Technologies used for implementation/modification and specific usage
  - Reference links

### Thinking Process

- Do not be swayed by user instructions or ideas; always engage in discussion and thinking
- Do not be bound to a single approach; maintain chessboard thinking (consider multiple options and find the optimal solution)

### Coding Style

- Coding style must follow "[The Zen of Python](https://www.python.org/dev/peps/pep-0020/)"

## Commit Conventions

- Use [gitmoji](https://gist.github.com/parmentf/035de27d6ed1dce0b36a) in commit messages
- Commit message format: `<gitmoji> <commit message>`

## File Structure

- **`docs/` directory**: Place information about project specifications (implemented content and requirements)
- **`outputs/` directory**: Place all information other than project specifications (Markdown format, not included in git history)
- Code files: Do not modify directly; describe as requirements definitions

## Maintenance Policy

### AGENTS.md Updates

- **Always check if AGENTS.md needs to be updated, and update it if necessary**
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
