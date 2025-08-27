package session

type Session struct {
	IsAuthenticated bool
	Username        string
	PartitionID     string
}

func NewSession() *Session {
	return &Session{
		IsAuthenticated: false,
		Username:        "",
		PartitionID:     "",
	}
}

func (s *Session) Login(username, partitionID string) {
	s.IsAuthenticated = true
	s.Username = username
	s.PartitionID = partitionID
}

func (s *Session) Logout() {
	s.IsAuthenticated = false
	s.Username = ""
	s.PartitionID = ""
}
