# Agent Instructions

This document outlines the rules and guidelines for AI agents working in this repository.

## 1. Task Management with `tasks.md`

- **Create Before Starting:** Before beginning any implementation, you must create or update the `tasks.md` file.
- **Detail Current Task:** Clearly describe the specific task you are currently working on in `tasks.md`.
- **Keep It Current:** Update the file every time you switch tasks. This ensures that anyone (human or AI) can quickly understand the project's status if your work is interrupted.

## 2. Error Handling

- **Don't Just Retry:** If you encounter an error, do not simply retry the same action.
- **Isolate the Problem:** Break down the failing task into smaller sub-tasks to pinpoint the source of the error. This approach facilitates more efficient debugging and resolution.

## 3. `tasks.md` Formatting

`tasks.md` must serve as a complete, historical log of all work. **Do not delete old entries.**

- `[v]` for completed tasks.
- `[x]` for failed tasks.
- `[ ]` for pending tasks.

## 4. Documentation Updates

- **Requirement Changes:** If a user's request involves adding or modifying requirements, you **must** update `requirements.md` to reflect these changes.
- **Usage Changes:** If a requirement change alters how the software is used, you **must** also update `README.md`.