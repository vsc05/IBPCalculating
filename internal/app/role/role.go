package role

type Role bool

const (
	IsModerator Role = true  // 0
	User        Role = false // 1
)
