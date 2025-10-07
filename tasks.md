# Schedule Management Feature Enhancement

- [x] **Milestone 1: Implement Multi-Participant Functionality**
  - [x] Update data models (`internal/model/schedule.go`) to include participant information.
    - [x] Add `ParticipantIDs` to `CreateScheduleRequest`.
    - [x] Add `ParticipantIDs` to `UpdateScheduleRequest`.
    - [x] Add a `Participants` slice of `UserResponse` to `ScheduleResponse`.
  - [x] Enhance repository (`internal/repository/schedule_repository.go`) to manage participants.
    - [x] Modify `CreateSchedule` to save participants in `schedule_participants` table.
    - [x] Modify `UpdateSchedule` to update participants.
    - [x] Modify schedule query methods to retrieve and attach participant data.
  - [x] Update handler (`internal/handler/schedule_handler.go`) to process participant data from requests.

- [x] **Milestone 2: Implement Schedule Viewing and Permissions**
  - [x] Create a new handler and route to allow users to view schedules of other users.
  - [x] Implement permission logic in the repository or handler to ensure only the creator of a schedule can edit or delete it.

- [x] **Milestone 3: Finalization**
  - [x] Review and test all new features.
  - [x] Update `requirements.md` if any clarifications or changes were made during implementation.