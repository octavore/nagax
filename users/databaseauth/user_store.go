package databaseauth

type UserStore interface {
	Create(email, password string) (id string, err error)
	Get(email string) (id, hashedPassword string, err error)
}
