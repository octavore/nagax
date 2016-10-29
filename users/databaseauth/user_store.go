package databaseauth

type UserStore interface {
	Get(email string) (id, hashedPassword string, err error)
}
