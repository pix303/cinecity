package batch

import (
	"log/slog"
	"sync"
	"time"

	"github.com/pix303/cinecity/pkg/actor"
)

type Batcher struct {
	messages         []actor.Message
	maxNumMessages   uint
	timeout          time.Duration
	timer            *time.Timer
	mutex            sync.Mutex
	processMessageFn func(actor.Message)
}

type MessageProcessHandler = func(msg actor.Message)

func NewBatcher(timeoutMs uint, maxMessages uint, fn MessageProcessHandler) *Batcher {
	b := Batcher{
		timeout:          time.Duration(timeoutMs) * time.Millisecond,
		messages:         make([]actor.Message, 0),
		mutex:            sync.Mutex{},
		maxNumMessages:   maxMessages,
		processMessageFn: fn,
	}
	slog.Info("Batcher created", "timeout", timeoutMs, "maxMessages", maxMessages)
	return &b
}

func (batcher *Batcher) Add(msg actor.Message) {
	batcher.mutex.Lock()
	if len(batcher.messages) == 0 {
		slog.Info("batch timer started", slog.Time("time", time.Now()), slog.Duration("timeout", batcher.timeout))
		batcher.timer = time.AfterFunc(batcher.timeout, batcher.process)
	}
	batcher.messages = append(batcher.messages, msg)
	slog.Info("batch msg added", slog.Int("totalMsg", len(batcher.messages)), slog.Int("max", int(batcher.maxNumMessages)))
	batcher.mutex.Unlock()

	if len(batcher.messages) >= int(batcher.maxNumMessages) {
		slog.Info("batch max messages reached")
		batcher.process()
	}
}

func (batcher *Batcher) process() {
	slog.Info("batch process started", slog.Time("time", time.Now()))
	for _, msg := range batcher.messages {
		batcher.processMessageFn(msg)
	}
	batcher.Stop()
	slog.Info("batch process end")
}

func (batcher *Batcher) Stop() {
	batcher.timer.Stop()
	batcher.messages = make([]actor.Message, 0)
}
