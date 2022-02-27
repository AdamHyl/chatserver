package worker

import (
	"github.com/AdamHyl/chatserver/common/timer"
	"sync"
	"sync/atomic"
	"time"

	"github.com/AdamHyl/chatserver/common/log"
)

const (
	DefaultChanLen            = 100000
	DefaultTimerDispatcherLen = 1000
)

type callInfo struct {
	name string
	cb   func()
}

type Worker struct {
	callInfos  chan *callInfo
	dispatcher *timer.Dispatcher

	wg        sync.WaitGroup
	stopped   bool
	active    int64
	mainStack []byte
}

func New(chanLen int, dispatcherLen int) *Worker {
	if chanLen <= 0 {
		chanLen = DefaultChanLen
	}

	if dispatcherLen <= 0 {
		dispatcherLen = DefaultTimerDispatcherLen
	}

	return &Worker{
		callInfos:  make(chan *callInfo, chanLen),
		dispatcher: timer.NewDispatcher(DefaultTimerDispatcherLen),
		active:     time.Now().Unix(),
	}
}

func (s *Worker) Run() {
	s.wg.Add(1)
	defer s.wg.Done()

	for !s.stopped {
		select {
		case ci := <-s.callInfos:
			if ci.cb != nil {
				// todo 使用耗时 recover 等
				// log.Debug("cb execute %v", ci.name)
				ci.cb()
			}

		case t := <-s.dispatcher.ChanTimer:
			if t != nil {
				// todo 使用耗时 recover 等
				//log.Debug("timer execute %v", t.Name)
				t.Cb()
			}
		}
	}
}

func (s *Worker) updateActive() {
	atomic.StoreInt64(&s.active, time.Now().Unix())
}

func (s *Worker) Stop() {
	s.Post("Stop", func() {
		if s.stopped {
			return
		}

		s.stopped = true
	})

	s.wg.Wait()
}

func (s *Worker) Post(name string, f func()) {
	if f == nil {
		return
	}

	select {
	case s.callInfos <- &callInfo{name: name, cb: f}:
	default:
		log.Error("Worker is full")
	}
}

func (s *Worker) GetRPCTaskNum() int {
	return len(s.callInfos)
}

func (s *Worker) NewTicker(name string, d time.Duration, cb func()) *timer.Ticker {
	return s.dispatcher.NewTicker(name, d, cb)
}
