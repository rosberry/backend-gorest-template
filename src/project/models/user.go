package models

type (
	// User is the user model of the mobile application.
	User struct {
		BaseModelWithSoftDelete
	}
)

func (u *User) Save() (ok bool) {
	return GetDB().Save(u).Error == nil
}

func NewUser() *User {
	user := &User{}
	user.Save()
	return user
}

func GetUser(ID uint) *User {
	var user User
	GetDB().First(&user, ID)
	if user.ID == 0 {
		return nil
	}
	return &user
}
