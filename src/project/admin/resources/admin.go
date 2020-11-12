package resources

import (
	"project/models"

	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
)

type Admin struct {
	models.Admin
}

func (Admin) ConfigureQorResource(res resource.Resourcer) {
	if res, ok := res.(*admin.Resource); ok {
		res.Meta(&admin.Meta{Name: "Status", Type: "select_one", Config: &admin.SelectOneConfig{Collection: [][]string{{"0", "Pending"}, {"1", "Active"}}}})
		res.Meta(&admin.Meta{Name: "StatusLabel", Label: "Status", Valuer: func(record interface{}, context *qor.Context) interface{} {
			result := ""
			if user, ok := record.(*Admin); ok {
				switch user.Status {
				case 0:
					result = "Pending"
				case 1:
					result = "Active"
				}
			}
			return result
		}})
		res.IndexAttrs("ID", "Email", "StatusLabel")
		res.NewAttrs("Email", "Status")
		res.EditAttrs("Email", "Status")
	}
}
