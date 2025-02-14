package report

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/slack-go/slack"
)

const (
	reportTimeLayout = "2006-01-02T150405.MST"

	cpuCommentFmt       = ":rotating_light:[CPU] usage (*%.2f%%*) > threshold (*%.2f%%*)"
	memCommentFmt       = ":rotating_light:[MEM] usage (*%.2f%%*) > threshold (*%.2f%%*)"
	goroutineCommentFmt = ":rotating_light:[GOROUTINE] count (*%d*) > threshold (*%d*)"
)

// SlackReporter is the reporter to send the profiling report to the
// specific Slack channel.
type SlackReporter struct {
	app     string
	channel string

	client *slack.Client
}

// SlackReporterOption is the option for the Slack reporter.
type SlackReporterOption struct {
	App     string
	Token   string
	Channel string
}

// NewSlackReporter returns the new SlackReporter.
func NewSlackReporter(opt *SlackReporterOption) *SlackReporter {
	return &SlackReporter{
		app:     opt.App,
		channel: opt.Channel,
		client:  slack.New(opt.Token),
	}
}

// uploadProfile is a helper function that handles the common logic for uploading profile data to Slack.
func (s *SlackReporter) uploadProfile(
	ctx context.Context,
	r io.Reader,
	filename string,
	comment string,
) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("autopprof: failed to read the profile file: %w", err)
	}
	filesize := len(data)
	if filesize == 0 {
		return fmt.Errorf("autopprof: profile file size is 0")
	}

	params := slack.UploadFileV2Parameters{
		Reader:         bytes.NewReader(data),
		Filename:       filename,
		FileSize:       filesize,
		Title:          filename,
		InitialComment: comment,
		Channel:        s.channel,
	}

	if _, err := s.client.UploadFileV2Context(ctx, params); err != nil {
		return fmt.Errorf("autopprof: failed to upload a file to Slack channel: %w", err)
	}
	return nil
}

// ReportCPUProfile sends the CPU profiling data to the Slack.
func (s *SlackReporter) ReportCPUProfile(
	ctx context.Context, r io.Reader, ci CPUInfo,
) error {
	hostname, _ := os.Hostname()
	now := time.Now().Format(reportTimeLayout)
	filename := fmt.Sprintf(CPUProfileFilenameFmt, s.app, hostname, now)
	comment := fmt.Sprintf(cpuCommentFmt, ci.UsagePercentage, ci.ThresholdPercentage)

	return s.uploadProfile(ctx, r, filename, comment)
}

// ReportHeapProfile sends the heap profiling data to the Slack.
func (s *SlackReporter) ReportHeapProfile(
	ctx context.Context, r io.Reader, mi MemInfo,
) error {
	hostname, _ := os.Hostname()
	now := time.Now().Format(reportTimeLayout)
	filename := fmt.Sprintf(HeapProfileFilenameFmt, s.app, hostname, now)
	comment := fmt.Sprintf(memCommentFmt, mi.UsagePercentage, mi.ThresholdPercentage)

	return s.uploadProfile(ctx, r, filename, comment)
}

// ReportGoroutineProfile sends the goroutine profiling data to the Slack.
func (s *SlackReporter) ReportGoroutineProfile(
	ctx context.Context, r io.Reader, gi GoroutineInfo,
) error {
	hostname, _ := os.Hostname()
	now := time.Now().Format(reportTimeLayout)
	filename := fmt.Sprintf(GoroutineProfileFilenameFmt, s.app, hostname, now)
	comment := fmt.Sprintf(goroutineCommentFmt, gi.Count, gi.ThresholdCount)

	return s.uploadProfile(ctx, r, filename, comment)
}
