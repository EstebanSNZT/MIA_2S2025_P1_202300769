package session

type Session struct {
	IsLoggedIn  bool
	Username    string
	GroupID     int32
	UserID      int32
	PartitionID string
}

func NewSession() *Session {
	return &Session{
		IsLoggedIn:  false,
		Username:    "",
		PartitionID: "",
	}
}

func (s *Session) Login(username string, groupID, userID int32, partitionID string) {
	s.IsLoggedIn = true
	s.Username = username
	s.GroupID = groupID
	s.UserID = userID
	s.PartitionID = partitionID
}

func (s *Session) Logout() {
	s.IsLoggedIn = false
	s.GroupID = -1
	s.UserID = -1
	s.Username = ""
	s.PartitionID = ""
}
