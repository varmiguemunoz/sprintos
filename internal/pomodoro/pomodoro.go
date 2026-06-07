package pomodoro

import (
	"fmt"
	"sync"
	"time"
)

type State int

const (
	Idle State = iota
	Running
	Grace
)

const graceDuration = 15 * time.Second

type Session struct {
	mu          sync.Mutex
	duration    time.Duration
	startedAt   time.Time
	graceAt     time.Time
	state       State
	stopCh      chan struct{}
	onExpired   func()
	onAutoStart func()
}

func New(onExpired, onAutoStart func()) *Session {
	return &Session{
		onExpired:   onExpired,
		onAutoStart: onAutoStart,
	}
}

func (s *Session) Start(minutes int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopCh != nil {
		close(s.stopCh)
	}

	s.duration = time.Duration(minutes) * time.Minute
	s.startedAt = time.Now()
	s.state = Running
	s.stopCh = make(chan struct{})

	go s.tick(s.stopCh)
}

func (s *Session) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopCh != nil {
		close(s.stopCh)
		s.stopCh = nil
	}
	s.state = Idle
}

func (s *Session) tick(stopCh chan struct{}) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			s.mu.Lock()
			now := time.Now()

			switch s.state {
			case Running:
				if now.Sub(s.startedAt) >= s.duration {
					s.state = Grace
					s.graceAt = now
					go s.onExpired()
				}
			case Grace:
				if now.Sub(s.graceAt) >= graceDuration {
					s.startedAt = now
					s.state = Running
					go s.onAutoStart()
				}
			}
			s.mu.Unlock()
		}
	}
}

func (s *Session) GetState() State {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}

func (s *Session) Remaining() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == Idle {
		return 0
	}
	if s.state == Running {
		rem := s.duration - time.Since(s.startedAt)
		if rem < 0 {
			return 0
		}
		return rem
	}
	return 0
}

func (s *Session) GraceRemaining() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != Grace {
		return 0
	}
	rem := graceDuration - time.Since(s.graceAt)
	if rem < 0 {
		return 0
	}
	return rem
}

func FormatDuration(d time.Duration) string {
	total := int(d.Seconds())
	if total < 0 {
		total = 0
	}
	m := total / 60
	sec := total % 60
	return fmt.Sprintf("%02d:%02d", m, sec)
}
