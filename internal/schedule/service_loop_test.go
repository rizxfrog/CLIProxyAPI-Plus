package schedule

import (
	"context"
	"testing"
	"time"
)

// The scheduler loop must keep firing interval tasks with no further mutation.
// An earlier version excluded running tasks from the wake computation and never
// signalled the loop when a run finished, so it fired once and then blocked on
// its wake channel forever. These tests drive the real loop, which the
// fireDue-level tests cannot do.
func TestSchedulerLoopKeepsFiringIntervalTasks(t *testing.T) {
	for _, triggerType := range []string{TriggerFixedRate, TriggerFixedDelay} {
		t.Run(triggerType, func(t *testing.T) {
			dir := t.TempDir()
			if _, errPut := NewStore(dir).Put(Group{
				ID: "live", Name: "live", Enabled: true,
				Tasks: []Task{{
					ID: "a", Name: "a", Enabled: true,
					Schedule: &Trigger{Type: triggerType, Value: "40ms"},
					Request:  Request{Method: "GET", URL: "https://example.invalid/never"},
				}},
			}); errPut != nil {
				t.Fatalf("store.Put: %v", errPut)
			}

			completed := make(chan RunEntry, 32)
			service := NewService(NewStore(dir))
			service.runner = &recordingRunner{}
			service.onRunComplete = func(_, _ string, entry RunEntry) { completed <- entry }

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if errStart := service.Start(ctx, dir); errStart != nil {
				t.Fatalf("Start: %v", errStart)
			}
			defer service.Stop()

			for want := 1; want <= 3; want++ {
				select {
				case <-completed:
				case <-time.After(10 * time.Second):
					t.Fatalf("only %d run(s) completed without further mutations; the loop stalled", want-1)
				}
			}
		})
	}
}

// An armed task that is already running still determines when the loop must wake,
// so a long run cannot put the loop to sleep past the next occurrence.
func TestNextWakeDelayIgnoresInFlightRuns(t *testing.T) {
	h := newHarness(t)
	h.put(rateGroup("g", "1h", "a"))

	delay, pending := h.service.nextWakeDelay()
	if !pending || delay <= 0 || delay > time.Hour {
		t.Fatalf("armed task: delay = %s, pending = %t", delay, pending)
	}

	h.service.mu.Lock()
	h.service.state["g/a"].running = true
	h.service.mu.Unlock()

	delay, pending = h.service.nextWakeDelay()
	if !pending {
		t.Fatal("a running task was treated as unarmed; the loop would block instead of waiting for its next occurrence")
	}
	if delay <= 0 || delay > time.Hour {
		t.Fatalf("running task: delay = %s", delay)
	}
}

// A group whose tasks have no trigger of their own must not keep the loop awake.
func TestNextWakeDelayReportsNothingWhenNoTaskIsArmed(t *testing.T) {
	h := newHarness(t)
	group := rateGroup("g", "1h", "manual")
	group.Schedule = nil
	h.put(group)

	if delay, pending := h.service.nextWakeDelay(); pending {
		t.Fatalf("unscheduled group reported a wake in %s", delay)
	}
}
