# Scheduled tasks

A generic HTTP-request scheduler: task **groups** hold **tasks**, and each group or
task carries a trigger. Definitions are stored as one JSON file per group under the
directory named by `automation.schedule.dir`, and are managed through the
`/v8/management/automation/schedule` API and the panel's **Scheduled tasks** page.

This is deliberately not a provider integration: a scheduled task is an arbitrary HTTP
request the operator authored, executed on a timer by the proxy process itself. It does
not participate in routing, model catalogues, quota accounting, or credential cooldown.

## Configuration

```yaml
automation:
  schedule:
    dir: ./schedules
```

The directory is resolved the same way `auth-dir` is. An empty `dir` disables the
feature: the scheduler never starts, no files are read, and the API reports
`configured: false`. The directory is created on first write; it is never created at
startup, so a typo cannot silently produce an empty schedule store.

## On-disk model

One file per group: `<dir>/<groupId>.json`. Written atomically (temp file plus rename)
with mode `0600`, because `request.headers` routinely carries bearer tokens.

```json
{
  "id": "daily-report",
  "name": "Daily report",
  "description": "",
  "enabled": true,
  "schedule": { "type": "cron", "value": "0 9 * * *" },
  "tasks": [
    {
      "id": "t_7f3a91",
      "name": "Generate report",
      "enabled": true,
      "schedule": null,
      "request": {
        "method": "POST",
        "url": "https://example.com/api/report",
        "headers": { "Authorization": "Bearer REDACTED" },
        "body": "{\"date\":\"today\"}",
        "form": null,
        "timeout_ms": 30000,
        "follow_redirects": false
      }
    }
  ]
}
```

### Group

| Field | Type | Rules |
| --- | --- | --- |
| `id` | string | Required. `^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`, and never `.` or `..`. Doubles as the file name, so it must stay URL- and path-safe. |
| `name` | string | Required, 1–200 characters. Display only. |
| `description` | string | Optional, ≤1000 characters. |
| `enabled` | bool | Disables every task in the group when false. |
| `schedule` | trigger or null | Group-level trigger. Null means each task schedules itself. |
| `tasks` | task[] | May be empty. |

### Task

| Field | Type | Rules |
| --- | --- | --- |
| `id` | string | Required, same pattern as group ids. Unique within the group. |
| `name` | string | Required, 1–200 characters. |
| `enabled` | bool | |
| `schedule` | trigger or null | Overrides the group trigger for this task. |
| `request` | request | Required. |

### Trigger

```json
{ "type": "fixed_rate", "value": "30s" }
```

| `type` | `value` | Fires |
| --- | --- | --- |
| `fixed_rate` | Go duration > 0 (`30s`, `5m`, `1h30m`) | Every interval, measured from the schedule activation anchor, regardless of how long a run takes. |
| `fixed_delay` | Go duration > 0 | Once the previous run has finished, wait the interval, then run again. |
| `cron` | five-field expression | At each matching wall-clock time. |

Neither interval trigger fires before one full interval has elapsed since the schedule was
activated, so activating a schedule never produces a startup burst. Cron follows wall-clock
time and therefore does fire on a matching boundary.

### Request

| Field | Type | Rules |
| --- | --- | --- |
| `method` | string | Uppercase HTTP token, default `GET`. |
| `url` | string | Absolute `http`/`https` URL. |
| `headers` | map[string]string | Optional. Sent verbatim; hop-by-hop headers (`Host`, `Content-Length`, `Connection`) are dropped. |
| `body` | string | Optional raw body. |
| `form` | form[] or null | Optional `application/x-www-form-urlencoded` fields, built per run. Mutually exclusive with `body`. |
| `timeout_ms` | int | 1–600000, default `30000`. |
| `follow_redirects` | bool | Default false, matching `curl`'s own default. |

`form[]` entries are `{ "name": string, "value": string }`.

## Trigger resolution

1. A task with its own `schedule` uses it.
2. Otherwise the task uses the group's `schedule`.
3. Otherwise the task is never scheduled and runs only on demand.

This covers both shapes the page offers: one group-level timer that drives every task in
it, or no group timer with each task timing itself. A task may also override an active
group timer.

A run is skipped when: the group is disabled, the task is disabled, the previous run of
that task is still in flight, or the request is invalid. Skipped runs are recorded as
`skipped` and never retried. Missed fires while the process was down are not backfilled;
the next fire is the next future occurrence.

## Runtime state

Held in memory only; a restart clears it. Summary per group and per task:

```json
{
  "scheduled": true,
  "trigger_source": "group",
  "next_run": "2026-11-11T01:00:00Z",
  "last_run": "2026-11-10T01:00:00Z",
  "last_status": "success",
  "last_status_code": 200,
  "last_duration_ms": 132,
  "last_error": "",
  "run_count": 4,
  "failure_count": 1,
  "runs": [ { "started_at": "...", "finished_at": "...", "status": "success",
              "status_code": 200, "duration_ms": 132, "error": "",
              "response_snippet": "{\"ok\":true}" } ]
}
```

`trigger_source` is `group`, `task`, or `none`. `last_status` is `never`, `success`,
`failed`, or `skipped`. `runs` keeps the most recent 10 entries per task, each response
snippet truncated to 2 KiB.

## HTTP API

All routes live under `/v8/management/automation/schedule` and require the management
credential, like every other v8 route. Errors are `{"error": "..."}` with status 400
(invalid), 404 (absent), or 409 (conflict).

| Method | Path | Body | Result |
| --- | --- | --- | --- |
| GET | `/status` | — | `{configured, dir, scheduler_running, server_time, group_count, task_count, enabled_task_count}` |
| GET | `/groups` | — | `{groups: [group + runtime + tasks[task + runtime]]}` |
| POST | `/groups` | group | 201, created group |
| GET | `/groups/:id` | — | one group with runtime |
| PUT | `/groups/:id` | group | replaced group. `:id` in the path wins over `id` in the body |
| DELETE | `/groups/:id` | — | `{deleted: true}` |
| POST | `/groups/:id/tasks` | `{tasks: task[]}` | 201, `{tasks: [...], skipped: [{index, reason}]}` — batch add |
| PUT | `/groups/:id/tasks/:task` | task | replaced task |
| DELETE | `/groups/:id/tasks/:task` | — | `{deleted: true}` |
| POST | `/groups/:id/tasks/:task/run` | — | `{run: runEntry}` after the request completes |
| POST | `/parse-curl` | `{curl: string}` | `{requests: [{name, request}], errors: [{index, message}]}` |

Every mutating route validates the whole document before writing, then notifies the
scheduler so the new definition takes effect without a restart.

`POST /parse-curl` never executes anything; it only converts text to drafts, so the page
can show the result for review.

### `POST /groups/:id/tasks` skipped entries

Batch add tolerates partial input instead of rejecting the whole batch, so a page that
pastes several `curl` commands can report which one was unusable:

```json
{ "tasks": [ { "id": "a", "name": "A", "request": { "url": "https://a" } } ],
  "skipped": [ { "index": 0, "reason": "duplicate id \"a\"" } ] }
```

A skipped entry means nothing was added for it; the request still succeeds when at least
one task was added, and returns 400 with `added: 0` when none were.

## `curl` import

`parse-curl` accepts one or more commands in a single blob, separated by newlines, `;`, or
`&&`, and maps them onto a group's tasks: a group created from several commands gets one
task per command.

Supported: `-X/--request`, the positional URL, `-H/--header`, `-d/--data`,
`--data-raw`, `--data-binary`, `--data-urlencode`, `--json`, `-F/--form`,
`--form-string`, `-u/--user`, `-b/--cookie`, `-A/--user-agent`, `-e/--referer`,
`-L/--location`, and `\`-continuations. Silently ignored: `-s`, `-S`, `-i`, `-I`, `-v`,
`-k`, `--compressed`, `--http1.1`, `--http2`, `--no-progress-meter`, `-o`, `-w`.

Anything else is reported as an error for that command rather than guessed at.

`--data-urlencode` follows `curl`'s own three forms: `name=value`, `=value`, and a bare
value that is encoded whole. `-u user:pass` becomes an `Authorization: Basic` header.
`-F` produces `form` entries; `--json` sets `Content-Type: application/json` and the body.
