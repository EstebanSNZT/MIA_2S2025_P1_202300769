package session

type Session struct {
	IsLoggedIn  bool
	Username    string
	PartitionID string
}

func NewSession() *Session {
	return &Session{
		IsLoggedIn:  false,
		Username:    "",
		PartitionID: "",
	}
}

func (s *Session) Login(username, partitionID string) {
	s.IsLoggedIn = true
	s.Username = username
	s.PartitionID = partitionID
}

func (s *Session) Logout() {
	s.IsLoggedIn = false
	s.Username = ""
	s.PartitionID = ""
}
