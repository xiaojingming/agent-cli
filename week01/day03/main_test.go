package main

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStdout runs f with os.Stdout redirected to a pipe and returns
// everything f printed to it.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()

	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	f()

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func TestTaskStatusValues(t *testing.T) {
	tests := []struct {
		status TaskStatus
		want   string
	}{
		{Pending, "pending"},
		{Running, "running"},
		{Retrying, "retrying"},
		{Succeeded, "succeeded"},
		{Failed, "failed"},
	}
	for _, tt := range tests {
		if got := string(tt.status); got != tt.want {
			t.Errorf("TaskStatus = %q, want %q", got, tt.want)
		}
	}
}

func TestRunTaskFailsAfterRetries(t *testing.T) {
	task := Task{name: "deploy", maxRetries: 3, status: Pending}

	var runErr error
	out := captureStdout(t, func() { runErr = runTask(&task) })

	if runErr == nil {
		t.Fatal("runTask returned nil, want error from toolCall")
	}
	if runErr.Error() != "temporary tool failure" {
		t.Errorf("runTask error = %q, want %q", runErr, "temporary tool failure")
	}
	if task.status != Failed {
		t.Errorf("final status = %q, want %q", task.status, Failed)
	}
	if got := strings.Count(out, "任务执行中"); got != 3 {
		t.Errorf("attempts = %d, want 3 (maxRetries)\noutput:\n%s", got, out)
	}
	for _, want := range []string{
		"任务进行第 1 次重试，失败原因 temporary tool failure",
		"任务进行第 2 次重试，失败原因 temporary tool failure",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\noutput:\n%s", want, out)
		}
	}
	if strings.Contains(out, "第 3 次重试") {
		t.Errorf("output contains a 3rd retry, want only maxRetries-1 = 2 retries\noutput:\n%s", out)
	}
	if want := "任务 deploy 失败，失败原因 temporary tool failure"; !strings.Contains(out, want) {
		t.Errorf("output missing %q\noutput:\n%s", want, out)
	}
}

func TestRunTaskSingleAttemptFailsImmediately(t *testing.T) {
	task := Task{name: "deploy", maxRetries: 1, status: Pending}

	var runErr error
	out := captureStdout(t, func() { runErr = runTask(&task) })

	if runErr == nil {
		t.Fatal("runTask returned nil, want error from toolCall")
	}
	if task.status != Failed {
		t.Errorf("final status = %q, want %q", task.status, Failed)
	}
	if got := strings.Count(out, "任务执行中"); got != 1 {
		t.Errorf("attempts = %d, want 1\noutput:\n%s", got, out)
	}
	if strings.Contains(out, "重试") {
		t.Errorf("no retry expected with maxRetries=1\noutput:\n%s", out)
	}
}

func TestRunTaskInvalidRetriesNoop(t *testing.T) {
	for _, maxRetries := range []int{0, -1} {
		task := Task{name: "deploy", maxRetries: maxRetries, status: Pending}

		var runErr error
		out := captureStdout(t, func() { runErr = runTask(&task) })

		if runErr != nil {
			t.Errorf("maxRetries=%d: error = %v, want nil", maxRetries, runErr)
		}
		if task.status != Pending {
			t.Errorf("maxRetries=%d: status = %q, want %q", maxRetries, task.status, Pending)
		}
		if out != "" {
			t.Errorf("maxRetries=%d: unexpected output %q", maxRetries, out)
		}
	}
}

func TestPrintTaskStatus(t *testing.T) {
	err := errors.New("boom")
	tests := []struct {
		name    string
		status  TaskStatus
		current int
		want    string
	}{
		{"running", Running, 1, "任务执行中"},
		{"retrying", Retrying, 2, "任务进行第 2 次重试，失败原因 boom"},
		{"failed", Failed, 3, "任务 deploy 失败，失败原因 boom"},
		{"succeeded", Succeeded, 1, "任务成功"},
		{"pending is silent", Pending, 1, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := Task{name: "deploy", maxRetries: 3, status: tt.status}
			out := captureStdout(t, func() { printTaskStatus(tt.current, task, err) })
			if tt.want == "" {
				if out != "" {
					t.Errorf("expected no output, got %q", out)
				}
				return
			}
			if !strings.Contains(out, tt.want) {
				t.Errorf("output %q does not contain %q", out, tt.want)
			}
		})
	}
}

func TestMainWithTaskName(t *testing.T) {
	t.Setenv("TASK_NAME", "build")

	out := captureStdout(t, main)

	if strings.Contains(out, "任务执行完成") {
		t.Errorf("toolCall always fails, so completion must not be printed\noutput:\n%s", out)
	}
	if !strings.Contains(out, "任务 build 失败，失败原因 temporary tool failure") {
		t.Errorf("missing failure output\noutput:\n%s", out)
	}
}

func TestMainEmptyTaskName(t *testing.T) {
	t.Setenv("TASK_NAME", "")

	out := captureStdout(t, main)

	if !strings.Contains(out, "任务名称不能为空") {
		t.Errorf("missing empty-name guard output\noutput:\n%s", out)
	}
	if strings.Contains(out, "任务执行中") {
		t.Errorf("unexpected run output for empty task name\noutput:\n%s", out)
	}
}
