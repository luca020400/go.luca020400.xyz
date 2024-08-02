package util

type User struct {
	ID   int64
	Nick string
}

type Todo struct {
	UserID    int64
	ID        int64
	Name      string
	Completed bool
}
