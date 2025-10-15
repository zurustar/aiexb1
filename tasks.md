# Project Task List

- [v] **Milestone 0: Project Setup**
  - [v] Initialize Go module and basic project structure.
  - [v] Set up the web server using `net/http`.
  - [v] Select and integrate a Pure-Go SQLite driver (`modernc.org/sqlite`).
  - [v] Implement database initialization logic to create tables.
  - [v] Configure database connection to use `/tmp/schedule.db`.
  - [v] Implement user registration endpoint (`/api/register`).
  - [v] Implement user login endpoint (`/api/login`) with JWT generation.
  - [v] Create middleware for JWT authentication to protect routes.
  - [v] Define user data models and repository interfaces.

# Schedule Management Feature Enhancement

- [v] **Milestone 1: Implement Multi-Participant Functionality**
  - [v] Update data models (`internal/model/schedule.go`) to include participant information.
    - [v] Add `ParticipantIDs` to `CreateScheduleRequest`.
    - [v] Add `ParticipantIDs` to `UpdateScheduleRequest`.
    - [v] Add a `Participants` slice of `UserResponse` to `ScheduleResponse`.
  - [v] Enhance repository (`internal/repository/schedule_repository.go`) to manage participants.
    - [v] Modify `CreateSchedule` to save participants in `schedule_participants` table.
    - [v] Modify `UpdateSchedule` to update participants.
    - [v] Modify schedule query methods to retrieve and attach participant data.
  - [v] Update handler (`internal/handler/schedule_handler.go`) to process participant data from requests.

- [v] **Milestone 2: Implement Schedule Viewing and Permissions**
  - [v] Create a new handler and route to allow users to view schedules of other users.
  - [v] Implement permission logic in the repository or handler to ensure only the creator of a schedule can edit or delete it.

- [v] **Milestone 3: Finalization**
  - [v] Review and test all new features.
  - [v] Update `requirements.md` if any clarifications or changes were made during implementation.

# Project Refinement and Hardening

- [v] **Milestone 4: Code Refactoring and Improvement**
  - [v] Refactor handler utilities to avoid function redeclaration.
    - [v] Create `internal/handler/handler_utils.go` for shared helper functions.
    - [v] Move `writeJSON` and `errorJSON` to the new utility file.
    - [v] Update `user_handler.go` and `schedule_handler.go` to use the shared utilities.
  - [v] Implement input validation.
    - [v] Add validation for user registration (username, email, password).
    - [v] Add validation for schedule creation (title, ownerID).
  - [v] Optimize database queries.
    - [v] Fix N+1 query problem in `schedule_repository.go`'s `FindByOwnerID` method.

# Documentation Update

- [v] **Milestone 5: Update and Translate `AGENTS.md`**
  - [v] Translate `AGENTS.md` from Japanese to English.
  - [v] Add the new rule for updating `requirements.md` and `README.md`.
  - [v] Verify `README.md` is already in English and requires no changes.

# UI/UX Improvement

- [v] **Milestone 6: Enhance Schedule Visualization**
  - [v] Change the schedule display from a list to a weekly calendar view.
  - [v] Add a vertical time axis to the weekly calendar.