# Backlog

## Inbox

- [ ] Replace Show all hours with an option to collapse disabled hours
- [ ] Add trash bin icon to Delete and remove the tint and make it the same color as Cancel
- [ ] On mobile, when editing response, there shall be no More options because there's only one option - Collapse disabled times
- [ ] Disable the Save button in Edit guest name (and in the form for adding the guest name)
- [ ] Add a placeholder in the required fields
- [ ] Make sure playwright saves traces (tmp/playwright/timed-event-weekly-firefox-26974-ields-through-the-edit-flow-firefox-desktop/error-context.md)
- [x] Move e2e to the root
- [ ] Pin Docker images to SHA in dockerfiles
- [ ] "Editing availability as" - use black font
- [ ] Split timed-event-helpers.ts
- [ ] Why do we need postcss? "<style lang="postcss">"
- [ ] Bug: When editing availability and having selected and highlighed a time slot with the cursor, when I uncollapse the bottom strip, the cursor moves down
- [ ] On mobile, "More options" is too narrow
- [ ] Edit event and Copy link are too big
- [ ] Migrate to StyleX
- [ ] Get rid of ::deep
- [ ] Fix lint warnings
- [ ] GitHub actions - Restore Go module cache doesn't restore anything - <https://github.com/timeful-foss/timeful/actions/runs/34223660393/job/102052462949?pr=25>
- [ ] Rename variables to refer to platform/event visitor identities
- [ ] Switch from swag
- [ ] Why uses SHA-256 for hashing?
- [ ] Why lax, not strict
- [ ] anonymous_event_creation must be allowed by default
- [ ] record the redemption approach (sign out when open a link)
- [ ] What is legacy in the implementation? Why need it?
- [ ] Why transferBoundary makes all these type checks?
- [ ] Get rid of POSTGRES_ANONYMOUS_EVENT_CREATION_ENABLED
- [x] Speed up e2e
- [ ] What is the difference between the default and postgres mode in e2e?
- [ ] Why need postgres-inspect.ts?
- [ ] Deprecate friends?
- [ ] Better hash salting?
- [ ] Use <https://github.com/golang-jwt/jwt>
- [ ] Set VITE_ENABLE_SIGN_IN=true in tests
- [x] timed-event-helpers: dow vs weekly
- [ ] get rid of "legacy" where it refers to Mongo, events stored in Mongo etc
- [ ] remove server/migrations from mongo, consolidate into a single init script for postgres
- [ ] super important rules about which files can be edited when?
- [x] ADR-019, ADR-012 - supersede and don't mention external_user_id
- [x] 24-hex id - use uuidv7 everywhere for platform visitor identity and event visitor identity
- [ ] run e2e selectively - only potentially affected tests, others will run in CI
  - Make this a decision in AGENTS.md? or somewhere else, should be an authoritative decision
- [ ] Mark such authoritative decisions or rely on the contex that a file contains only authoritative decisions?
- [x] Always show logs when running e2e tests. Also write to a file if needed, never use just tail
- [ ] Mention the ADR status (superseded, deprecated) in the adr title
- [ ] In docs/design/readme.md, add a separate table for deprecated records
- [ ] Switch to <https://github.com/DeusData/codebase-memory-mcp>
- [ ] Run the same e2e on all platforms
  - [ ] desktop/mobile may differ
  - [ ] firefox desktop and chrome desktop should be the same
- [ ] task-0178 - consider using another approach because these libraries are deprecated
- [ ] remove "cutover", "legacy", "mongo", "tombstones" mentions
- [ ] reorder e2e and set individual timeouts for slow tests (it's luck that they're one of the first ones)
- [ ] cut over
- [ ] Identify
- [ ] "Each change must be committed together with the task file" - record in /commit skill
- [ ] Add backlog dod
  - [ ] - all relevant changes are staged
  - [ ] - critically assessed the task against the current repo state and suggested changes
  - [ ] - asked a sub-agent to identify any problems and suggest fixes
  - [ ] - Move to completed
    - Declined - doesn't look like renamed because the diff can be too large
- [ ] Analyze tasks completed during the transition
  - [ ] What we did?
  - [ ] What we haven't done?
- [ ] RemoveAttendee - removed
  - Deletion loses the single-row pgx.ErrNoRows semantics for missing removals (batch removal intentionally ignores missing emails); I updated TestAttendeeRepositoryMembershipLifecycle accordingly.
- [ ] Cover the "Down" path in migration tests
- [ ] ADR - we use agents to write code. Affects reliability
- [ ] Show "View event info" to non-owners
- [ ] Add review-commit - what changed vs what had to change?
- [ ] "reviewed_date" in task (needs forking backlog.md)
- [ ] write docs/configuration.md that explains which artifacts we manage
- [ ] why are they called "postgres_events" in sql? No need to emphasize that, I guess
- [ ] Instruct agent to set GOCACHE to a local directory inside ./tmp/ if they really want to
- [ ] Document ./tmp as a local directory for local files that helps bypass sandbox restrictions if any
- [ ] Add template for CAND
- [ ] 224.03 - don't mention guest_id in postgres-event-api-contract
- [ ] consider introducing an ADR and making an interface package for database interaction, like in three-layer cake
- [x] Increase time budget for these tests:
  - [x] ✘   1 [firefox-desktop] › specs/timed-event-access-transfer-firefox.spec.ts:91:3 › Source approves the exact target code for owner access (36.8s)
  - [x] ✘   2 [firefox-desktop] › specs/timed-event-access-transfer-firefox.spec.ts:91:3 › Source approves the exact target code for guest access (36.8s)
- [ ] Set up playwright mcp - npx @playwright/mcp@latest
- [ ] Configure posthog to do something useful
- [ ] Configure backlog.md to check all branches for task ids and avoid same ids - <https://github.com/time-tools/timeful/pull/50>
- [ ] Use openobserve v1.x.x
- [ ] Consider renaming "timeful" - "timepoll"
- [ ] Make all logs structured
- [ ] Schedule on Timeful - available only to the admin
- [ ] Configuration bug: on the server, can't log in with credentials from .env.production and .env.staging
- [ ] Check deployment instructions work end to end
  - [ ] Fresh installation
  - [ ] Updates
- [ ] Unify handoff with task progress

---

Semi-structured TODO list, grouped into thematic classes.
Priority tags carried over from the former MUST/SHOULD/COULD sections appear inline: `[MUST P0]`, `[MUST P?]`, `[SHOULD]`, `[COULD]`; untagged items are regular inbox candidates.
ADR candidates stay in their topical sections.

## Temporal correctness — specific times & time zones

- [ ] Create an event for may 28 with availability from 0 to 4 and timezone +02:00.
  If you open the date picker in +0:00, should you see two days marked in the date picker?
- [ ] Mismatch between event page and specific times <http://127.0.0.1:4173/e/Eb67A>
- [ ] When creating event with specific times and setting timezone, on the specific times page, the timezone should be set to the one specified in the event (during creation).
  For a user, shown in should be the same on the event page and specific times page
- [ ] Lost slots on the specific times page when selected all slots at +7, then switched to +6
- [ ] <http://127.0.0.1:4173/e/EE2Fc> When editing event specific times, when "shown in" is +7, all slots for jun 15 are selected.
  When it's set to +6, jun 14 is duplicated and jun 15 is missing and there's a gap between jun 14 and jun 15.
- [ ] When I set specific times 0-4 on jun 14 and jun 16 when shown in is +5, then switch to +3, jun 15 doesn't appear although it should take some timeslots from jun 16
- [ ] what is the source of truth for enabled slots?
  - date picker may show dates that aren't shown when setting specific times. specific times editor should show these dates even if there are no enabled slots at these dates.
- [ ] should be able to edit specific times again
- [ ] why grey without grid in <http://127.0.0.1:4173/e/Eb67A> at gmt+9?
- [ ] Given an event was scheduled and time zone switched so that event slots shifted to another date, mark them blue in relevant dates
- [x] What did we decide about the calendar - should days before today be selectible?
  - See FR-118
- [ ] "Disabled, change in Edit event" - rename to "Disabled, inside the event dates in the event timezone"
- [ ] [MUST P?] event in +3, edit specific times in +9, some time slots are lost
- [ ] [MUST P?] specific times uses the default 9-18 after unselecting 9-18
- [ ] [MUST P?] For date-specific events, make the dates not disappear
  - For <http://127.0.0.1:4173/e/JTGTEFXY>, the dates disappeared several times, maybe because the agent run a container with another db volume
- [ ] [MUST P?] Remove "Shown in"

## Availability editing & event page UX

- [ ] When adding availability, cancel and save should be aligned to the right
- [ ] Collapse hours when editing response availability
- [ ] At <http://127.0.0.1:4173/e/6df78>, when I hover over Maya Patel, I see if needed (yellow) for jun 20, 13:45-14:30 but it's available (green) when I edit her availability.
- [ ] Move Show all hours to over Overlay availabilities on desktop and mobile
- [ ] On desktop, on the event page, when I scroll the grid and the top border of the grid isn't visible, then the space above sidebar shall be collapsed because it's not needed to separate Schedule event from the time format switch
- [ ] For <http://127.0.0.1:4173/e/C9ZC3WZS>, Edit availability is disabled.
  However, I can edit responses by clicking the pencil icon in responses
- [ ] When there's no response from the guest, "Add availability" shall blink.
  Clicking a timeslot isn't necessary to trigger blinking
- [ ] Change the legend label for red: "Everyone is unavailable at this timeslot; Add/Edit availability"
- [ ] Bug: on mobile, when I add availability and select a timeslot and drag pointer down and my finger is on the collapsed strip then while I hold the finger, the pointer stays at the lowest timeslot
- [ ] When dragging to schedule an event, the tooltip shall show the range of the event, not of a timeslot
- [ ] On mobile, in a timed event, when scheduling an event and dragging, the timeslot pointer shall be visible
  - The bug: click a timeslot at the bottom of the grid, let the overlay panel slide and cover it, see the cursor disappear
- [ ] On the Event Response Editing Page, move Calendar options outside of its section:
  - [ ] on desktop - to under time format and time zone row
  - [ ] on mobile - to under More options row
- [ ] On the Event Response Editing Page:
  - [ ] replace More options with the single option inside it (Show all hours)
- [x] On the Event Response Editing Page:
  - [x] [MUST P?] Show input form for editing the respondent name instead of Editing availability as
    - "Editing availability as" shouldn't be in italic
    - Given I edit availability, I should see Editing availability as - add input field to write the name above Available
- [ ] How to show response selections so that the status is still visible for selected responses?
  - Maybe show the checkboxes to the left of the status?
- [x] Move the "Note: there's no time when ..." from below the grid to the Responses section
- [x] When saving added availability, in the Continue as guest form, the Continue button shall be flat and without glow
- [x] [MUST P0] Make Responses scrollable on desktop
  - [x] Decide at what height to make scrollable
- [ ] [MUST P?] when autofill is disabled, only manually should be enabled
- [ ] [MUST P?] For dates-only events, when I edit event and enable Overlay availability, then I shall see all responses.
  - Respondent's response is overlaid
  - Others' responses are shown

  Problem:
  - Currently, I see none of the responses

## Mobile & responsive layout

- [ ] Make available/If needed on mobile higher to cover the "Adding availability" text
- [x] Handle narrow mobile screens (e.g. iPhone 17)
  - vite doesn't allow to define a custom breakpoint, e.g. xs (~450) so I decided that we keep sm at 640
- [ ] font size of hours and days of week too large on mobile
- [ ] reduce the number of columns on mobile so that the event can fit into that
- [ ] On mobile, on event page, when there are no responses, center Show all hours within its column
- [ ] On mobile, when user added availability and need to write the guest name, the input form shall be above the keyboard, near the top of the page
- [ ] On mobile, make the lower panel with buttons visually separate from the grid (e.g. elevated and white).
  A green panel fuses with the green grid background
- [ ] On mobile, when the reset button is clicked, show a little tooltip (or a banner) that time zone was reset to the event time zone
- [x] On Mobile, on the Event Response Editing Page, make the bottom panel elevated
- [ ] On mobile, On the Event Page, don't show the switch (3d 7d if there are less than 7 days)
- [x] On mobile, on the Event Response Editing Page, in the activity bar, the Save button shall be solid green and without shadow, should be just like Save on desktop
- [x] On mobile, the Legend shall be fully visible and not hidden under the Available/If needed panel
- [ ] Center edit event and new event form on screens 640<
- [ ] On mobile, on the event page, when I click a slot and I see the tooltip and scroll the page down, and I see the Responses panel, then the tooltip shall be below the bottom overlay panels
- [ ] On mobile, when [3 days / 7 days] switch isn't visible, show [Time format [12h / 24h] Time zone [ +N:NN ]]
- [ ] [MUST P?] On mobile, for dates-only events, show the buttons Overlay availability and Start on Monday
- [ ] [MUST P?] landing on mobile - decide which buttons should go into the hamburger menu
- [ ] [MUST P?] On the event page, Edit event and Copy link buttons shall look the same on mobile and desktop

## Styling & design system

- [ ] Use the right palette consistently for dropdowns, selects, buttons, switches
- [ ] make a design system
- [ ] don't modify vuetify internals (deep)
- [ ] Remove magic constants in CSS
- [ ] During the sign-in, use green color for links
- [ ] Use 36 px for (almost) all chips and buttons
- [x] [MUST P0] The built CSS doesn't match the preview.
  E.g, bottom tool row is green
- [ ] [MUST P?] Remove split-gap
- [ ] [MUST P?] On timed event page, Create an event and Give feedback should be bold like on dates-only page
- [ ] [SHOULD] The grid lines should be black, not grey
- [ ] [COULD] use @dicebear/identicon for avatars, not generic head?
  The downside is that after updating the name, the avatar will change too.

## Code quality & frontend tooling

- [ ] Don't mention "legacy" in the code
- [ ] remove `as unknown as`
- [ ] add eslint rule for `as unknown`
- [ ] is `$el` idiomatic modern syntax?
- [ ] Check against composition API
- [ ] refactoring - get rid of duplication
- [ ] make more functions for business logic pure
- [ ] Get rid of eslint-disable-*
  - [ ] `eslint-disable vue/one-component-per-file`
- [ ] use treefmt-nix + oxfmt
- [ ] Replace `tw-z-[60]` in App.vue with a variable in index.css
- [ ] [MUST P0] Switch to pnpm
- [ ] [MUST P?] What is "compact" for?
- [ ] [SHOULD] When `typescript-eslint` supports the installed TypeScript native bridge version in its peer dependency range, verify `npm ci --dry-run` succeeds without overrides and remove `--legacy-peer-deps` from `frontend/Dockerfile`.

## Testing & CI

- [ ] Resolve `schedule-overlap-mobile-scroll:8` under chromium-mobile: read-only grid collapses to a single `data-row="0"` 15px row after the create-dialog flow; unit contract is `ScheduleOverlap.collapsedHours.test.ts` ("collapses the read-only specific-times band to the saved active subset") — decide spec fix vs product fix
- [ ] Stabilize `ensureSpecificTimesEditorMode` / `isSpecificTimesEnabled` E2E toggling in `firefox-touch` and `chromium-mobile`; remove force-click fallback only after reliable native interaction coverage
- [ ] Investigate intermittent `firefox-touch` mobile tooltip/navbar layering failure.
  Verify Responses tap isolation, outside dismissal, and visibility/`mouseleave` behavior on a real Firefox mobile device; Playwright touch emulation is insufficient
- [ ] In `frontend/e2e/sign-up-form-event.spec.ts`, capture a non-OK event POST's status and response body when it recurs, then compare its payload with `server/routes/events.go` timed-payload normalization
- [ ] If unit tests become a CI bottleneck, profile with `npm run test:unit:profile`; retain default workers and `slowTestThreshold: 100` unless target-CI benchmarks justify a change.
  Convert suites away from `happy-dom` only when verified Node-safe
- [ ] enable tests with chromium in addition to Firefox
- [ ] Set up CI/CD (maybe CD on releases only)
- [ ] Track line and branch coverage
- [ ] fix backend CI
- [ ] Does it make sense to assert on source code in tests?
- [ ] [MUST P0] Cache pnpm directory via cache-nix-action
- [ ] [MUST P?] Failing e2e test "sign-up blocks are visible on the event page" — unrelated pre-existing: the spec seeds a sign-up event with legacy fields (duration, dates, timeIncrement, startOnMonday), which POST /api/events has rejected with 400 legacy-timed-event-field:* since commit f7817601 (2026-07-30).
  Reproduced via direct curl against the test API.
- [ ] [MUST P?] Revive the inspect scripts?
  - Restructure the inspect scenario to not require signed-in session
  - "The inspect scenario's prepare needs a signed-in session to open the "New event" dialog."

## Docs, ADRs, requirements & glossary

- [ ] Add adr about using vue 3 and vuetify
- [ ] specify in the docs how the color is calculated at:
  - overlapping slots
  - best times
- [ ] Document the architecture and integration with external systems, e.g. Google Cloud project
- [ ] Introduce a log of non-architectural decisions with SPEC-NNN identifiers
- [ ] Introduce an index that tracks the status of SPECs
- [ ] Introduce UDR - universal decision records along with ADRs?
- [ ] Don't mention a particular ADR in AGENTS.md
- [ ] Check that docs don't contain stale references or too specific references to other files.
  Such references can misguide agents
- [ ] Update backlog instructions to not run e2e and unit tests for docs-only changes
- [ ] Review terms used in FRs and link where the link is missing
- [ ] ADRs should reference QRs, FRs may reference ADRs
- [ ] move adr to architectural-decision-records/adr? and add README.md near adr?
- [ ] Define "instance operator" - someone hosting Timeful
- [ ] ADR - Use stalwart instead of listmonk
- [ ] Rename variables in the code and modules to match the glossary
- [ ] Add glossary term - pointer (box that highlights a cell)
- [ ] Document the traceability structure
- [ ] Automate generating tables in
  - docs/requirements/README.md
  - docs/design/README.md
- [ ] Mark primary FRs as `derived_from: is_primary`, not just an empty array which might mean that the sources are just not determined
- [ ] FR-005 - what are "overlays"?
- [ ] Define slot cursor
- [ ] bug: in Timed Domain Mode section, there's more than one sentence on a line.
  Need to fix formatter to detect such problems
- [x] review glossary:
  - [x] Availability Editing
  - [x] rename: Scheduled event time -> Event Scheduled Time
  - [x] Slot Increment - not an interval between timed slots
  - [x] rename: Timed Grid -> Timed Event Grid
- [ ] Review FR wording
  - [ ] Depending on the context, use timed grid or full "timed event page grid"
  - [ ] Take note of consequtive tems `] [`
- [ ] bug: FR-012 - only owner can edit Scheduled Span, not everyone
- [x] Need to review requirements from the pov of the ownership model
- [ ] list only allowed aliases in glossary, remove rejected
- [ ] Mention DDD - at least for naming, not for design
- [ ] Give names to parts of the Desktop UI
  - [ ] Left part of the header (with event info and buttons)
  - [ ] Right part of the header (where buttons are grouped)
  - [ ] Sidebar (on the right of the grid)
  - [ ] Indicator - thing selecting an option inside a toggle
- [ ] [MUST P?] In glossary, define "guest", "anon"
- [ ] [MUST P?] Refactor ADRs
  - [ ] Move ADRs from `frontend/adr` to `adr` (repo root)
  - [ ] Add README that explains the ADR format
  - [ ] Specify the scope inside the ADR - frontend/backend
  - [ ] in the README, generate tables with ADRs (ID, title) for frontend and backend
    - Maybe use mdsh
  - [ ] Document important quality attributes - performance, maintainability, reliability, security, usability
  - [ ] Select true ADRs that affect important quality attributes
- [ ] [MUST P?] Introduce specs
  - [ ] Identifiers start with `SPEC-`
  - [ ] Current not "true" ADRs can be the first specs
  - [ ] SPECs are affected by ADRs
- [ ] [MUST P?] ADR candidate:
  - (Scope: frontend, backend): backend handles only particular paths for initial HTML with essential metadata
  - Using Crockford base32 encoding (8 characters) with repeated probings for event identifiers (less collisions)
- [ ] [MUST P?] Specify the context for each FR (where is it applicable)

## Self-hosting, deployment & ops

- [ ] Before production deployment, run and record the preflight status of `server/scripts/20260724_canonical_timed_events`; deploy the server before a frontend release that requires derived enabled-slot validation
- [ ] Refactor .env files
  - Group values
  - Shift the most important higher
- [ ] `.env.staging.example` and `.env.production.example` shouldn't have `VITE_PREVIEW_*`
- [ ] Log rotation
- [ ] Send alerts to the platform hoster via email
- [ ] Write a system-manager config
  - [ ] fail2ban
  - [ ] necessary deps like docker and its dns
- [ ] Explain how to update staging, prod
- [ ] [MUST P?] Install buildx on the VM.
  Logs when deploying:

  ```text
  time="2026-08-13T23:29:13+03:00" level=warning msg="Docker Compose is configured to build using Bake, but buildx isn't installed"
  ```

- [ ] [MUST P?] Document features of a good email sending service
- [ ] [MUST P?] Document DNS records

## Auth, identity & feature flags

- [ ] add flag to disable sign-in
  - document flags that disable features
  - make it possible to disable sign-in on backend
- [ ] Flags for Sign up with Google, Outlook.
  - [ ] Disabled by default (not allowed in Russia)
  - [ ] Affects the sign in flow
- [ ] Sign out button
- [ ] Candidate ADR - the platform visitor identity shall not be compromised when the event visitor identity is compromised
- [ ] Enabled scenario: use event visitor id to revoke access for that visitor to responses in that event
- [ ] [MUST P?] Use the following sign-in flow:
  - User enters email and password
  - If user forgot a password:
    - User clicks Forgot password?
    - Timeful sends a magic link to the email
    - Timeful suggests to use it to restore the password
    - User opens the email and clicks the link
    - User is redirected to the Timeful site
    - User enters a new password
  - If email is not recognized, Timeful suggests to sign up
  - If email is recognized, sign in succeeds
- [ ] [COULD] anon identity to save preferences, maybe sign in by password

## Product features & roadmap

- [ ] optional password for restoring.
  edit own responses and open for editing, can click the lock button to enter password and edit others' responses
- [ ] Who are the group respondents?
- [ ] what is group (NewGroup.vue)?
- [ ] Schedule:
  - Phase 1 - everyone
  - Phase 2 - only the event owner, be it a registered or an anon user
- [ ] Cap the number of guests to a smaller number than one tested via the QR, e.g. 200 if tested 500
- [ ] Support adding calendars by link
  - Can't share a link to a calendar in the Google Calendar Android app
  - Service status dashboard - like <https://status.openai.com/>
- [ ] A guest shall see which responses are editable by them and which are protected
- [ ] The description shall be inside "Edit event"
- [ ] Show Help button somewhere at the header.
  Provide there instructions depending on the role
  - For the event owner: Copy link and send it to others
  - For a guest without a response: Click Add availability below to mark in the grid when you're available
  - For a guest with a response: Add availability or Edit availability
- [ ] Only the event owner shall see edit event, guests will see "event info"
- [ ] Blind event (only the event owner sees others' responses)
- [ ] In new event form and using range, Validate the range
- [x] Remove time format from event creation/editing form
  - Decided to keep it for convenience
- [x] The descriptions shall be editable in the event form
- [x] In the new event form, move set specific times per day to the start of the section What times might work, right under the title
- [ ] Given I edit event and I change something, when I click to close the modal and I see Unsaved changes banner and I click Leave page and I open Edit event again, then I shall not see the Unsaved changes banner
- [x] In the new event/edit event form, style the "Name your event" input like the Description input.
  Allow only one line and cap at N characters
- [ ] Change PLUGIN_API_README.md to use not strings?
- [ ] [MUST P?] make the app name configurable and when2meet by default
  - switch to when2meet in the repo
  - Rename to "meetwhen"
  - what is when2meetHref?
- [ ] [MUST P?] add i18n (Russian, German)
- [ ] [SHOULD] User settings for the time format
- [ ] [SHOULD] Cookie consent overlay
- [ ] [SHOULD] Add concurrency control
  - Rule: When one user edits an event, others can only view
  - Scenario: one user edits the event, another one edits the availability
  - Rule: When one user has saved event changes (info, availability), others receive them immediately
- [ ] [SHOULD] In Create event form, I should be able to write the event description:
  - potential problem: the description will be formatted differently than on the event page
  - another one: might distract the event creator
  - maybe just let them know that only they can edit the description?

## Performance & scale

- [ ] Speed up eslint and reduce memory consumption from 11GB
- [ ] load all routes lazily
- [ ] Possibly lazily load timed slots when viewing to avoid browser-storage bloat?
- [ ] [MUST P?] RIIR only if benchmarks justify it
  - [ ] Load-test the Go version and decide based on data
  - [ ] use dbfirst
  - [ ] don't fix sign in functionality in the Go version
  - [ ] keep the original code in comments for line-by-line rewriting

## Landing, demo & marketing

- [ ] rich landing enabled flag - enable more than just the title and demo
- [ ] improve readme
  - [ ] update the site link
  - [ ] add warning about unstability and possible loss of information and under construction
  - [ ] update technologies
- [ ] in FAQ, align text and +
- [ ] in FAQ, don't mention calendars when sign in is disabled
- [ ] How it works section still exists?
- [ ] Summarize feedback on <https://www.reddit.com/r/schej/>
- [ ] [MUST P?] Update demo
  - Make a demo screenshot of <http://127.0.0.1:4173/e/6df78>
  - first convert png to webp <https://picflow.com/convert/png-to-webp>
  - remember scale and screen size and make full page
  - then <https://ezgif.com/webp-maker>
    - 600 ms delay
    - 1200 ms for the last one
    - quality 90
  - <https://ezgif.com/webp-maker/ezgif-62a1b9a1b6704abf-split.html>
  - Hover over participants, then grid, then select best times, then create event on timeful

## Graphify & repo tooling

- [ ] add instructions for the agent to write scripts for the browser and edit it instead of inline scripts?
- [ ] Research how to update graphify semantic index when using OpenCode
- [ ] Make the handoff skill wording prompt to create a handoff more certainly
- [ ] Make create-handoff.sh an app in flake.nix
- [ ] Decompose root .gitignore into per-directory gitignores (graphify, infra, etc.).
  Instruct agents to use per-directory gitignores
- [ ] What connects StageName, StageResult, PropertyGroupName to the rest of the system? (1,745 weakly-connected nodes — possible doc gaps)
- [ ] graphify can't process sql files - tree_sitter_sql — pip install 'graphifyy[sql]'
- [ ] What is the exact relationship between Sign-In Link and Platform Sign-In? (AMBIGUOUS edge)
- [ ] [MUST P?] Set up graphify
  - Preferably add a successfully buildable package to flake.nix' devshell

## Backend & data model

- [ ] Use SQL/PGQ from Postgres 19
- [ ] ADR candidate - Use postgres instead of mongo for better maintainability and speed - clear schemas.
  - Using Postgres needs an ADR?
- [ ] [MUST P?] Specify API response normalization and handling of legacy respondent-name rows
- [ ] [MUST P?] Product questions
  - Anonymous event metadata is publicly editable.
  - Selected schedules are publicly replaceable and clearable.
  - Blind-response viewing accepts guest ID/name without an edit token.
  - Guest edit tokens are stored in plaintext.
  - Dates-only values are currently stored as instants rather than explicit civil dates.
  - Empty weekly events lack a durable anchor week.
  - numResponses is a drift-prone cached value.
  - Anonymous-event adoption after OAuth currently updates Mongo only.
  - Dates-only events should use civil dates, not instants
- [ ] add more instrumentation? (observability)
