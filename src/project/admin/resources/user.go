package resources

import (
	"project/models"

	"github.com/qor/qor/resource"
)

type User struct {
	models.User
}

func (User) ConfigureQorResource(res resource.Resourcer) {
}
