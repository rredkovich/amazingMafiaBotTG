package game

import "sync"

type State int

const (
	Prepairing State = iota
	Night
	Day
	DayVoting
	Stopped
)

type SafeState struct {
	//isPlaying  bool
	//isStarting bool
	//dayORNight bool //true - day, false - night
	state State
	mux   sync.Mutex
}

func (s *SafeState) SetStarted() {
	s.mux.Lock()
	defer s.mux.Unlock()

	//s.isPlaying = true
	//s.isStarting = false
	s.state = Night
}

func (s *SafeState) SetPrepairing() {
	s.mux.Lock()
	defer s.mux.Unlock()

	//s.isPlaying = false
	//s.isStarting = true
	s.state = Prepairing
}

func (s *SafeState) SetStopped() {
	s.mux.Lock()
	defer s.mux.Unlock()

	//s.isPlaying = false
	//s.isStarting = false
	s.state = Stopped
}

// IsStopped marks if game is alive
func (s *SafeState) IsStopped() bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	//return s.isPlaying || s.isStarting
	return s.state == Stopped

}

func (s *SafeState) SetDay() {
	s.mux.Lock()
	defer s.mux.Unlock()

	//s.dayORNight = true
	s.state = Day
}

func (s *SafeState) SetNight() {
	s.mux.Lock()
	defer s.mux.Unlock()

	//s.dayORNight = false
	s.state = Night
}

func (s *SafeState) IsItDay() bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	return s.state == Day || s.state == DayVoting
}

func (s *SafeState) IsPlaying() bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	return s.state != Prepairing || s.state != Stopped
}

func (s *SafeState) HasStarted() bool {
	return s.state != Prepairing
}

/*
MoveToNextState moves game through possible states:

Prepairing -> Night
Night -> Day
Day -> DayVoting
DayVoting -> Night
?could be Night -> Stopped if no more members alive !!! not responsibility of State to determine
*/
func (s *SafeState) MoveToNextState() {
	s.mux.Lock()
	defer s.mux.Unlock()

	switch s.state {
	case Prepairing:
		s.state = Night
	case Night:
		s.state = Day
	case Day:
		s.state = DayVoting
	case DayVoting:
		s.state = Night
	}
}
