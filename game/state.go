package game

import "sync"

type SafeState struct {
	isPlaying  bool
	isStarting bool
	dayORNight bool //true - day, false - night
	mux        sync.Mutex
}

func (s *SafeState) SetStarted() {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.isPlaying = true
	s.isStarting = false
}

func (s *SafeState) SetPrepairing() {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.isPlaying = false
	s.isStarting = true
}

func (s *SafeState) SetStopped() {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.isPlaying = false
	s.isStarting = false
}

func (s *SafeState) IsActive() bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	return s.isPlaying || s.isStarting

}

func (s *SafeState) SetDay() {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.dayORNight = true
}

func (s *SafeState) SetNight() {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.dayORNight = false
}

func (s *SafeState) GetDayNight() Daytime {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.dayORNight {
		return Day
	}

	return Night
}
