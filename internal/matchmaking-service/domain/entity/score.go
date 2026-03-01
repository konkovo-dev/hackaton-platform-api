package entity

type TeamScore struct {
	Team  *Team
	Score MatchScore
}

type CandidateScore struct {
	Participation *Participation
	User          *User
	Score         MatchScore
}
