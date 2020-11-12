package admin

import (
	"html/template"
	"net/http"
	"project/bindatafs"
	"project/models"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/auth"
	"github.com/qor/i18n"
	"github.com/qor/i18n/backends/database"
	"github.com/qor/roles"
	"github.com/qor/sorting"
	"github.com/qor/validations"

	adminAuth "project/admin/auth"
	"project/admin/resources"
	"project/auth_themes/clean"
)

func HandleGin(router *gin.Engine) {
	qorAdmin, qorAdminHandlers := initResources()

	mux := http.NewServeMux()
	qorAdmin.MountTo("/admin", mux)
	router.Any("/admin/*resources", gin.WrapH(mux))

	mux2 := http.NewServeMux()
	mux2.Handle("/auth/", qorAdminHandlers)
	router.Any("/auth/*resources", gin.WrapH(mux2))
}

func noescape(str string) template.HTML {
	return template.HTML(str)
}

func initResources() (*admin.Admin, http.Handler) {
	db, _ := getDB()
	sorting.RegisterCallbacks(db)
	adminAuth, adminAuthHandle := getAdminAuth(db)
	I18n := i18n.New(database.New(db))
	Admin := admin.New(&admin.AdminConfig{DB: db, Auth: adminAuth, I18n: I18n, AssetFS: bindatafs.AssetFS})

	Admin.AddResource(&resources.Admin{}, &admin.Config{
		Menu: []string{"User Management"},
		Permission: roles.
			//Deny(roles.Delete, roles.Anyone).
			//Deny(roles.Update, roles.Anyone).
			Deny(roles.Create, roles.Anyone),
	})

	Admin.RegisterFuncMap("noescape", func(str template.JS) template.HTML {
		return template.HTML(string(str))
	})

	//Admin.RegisterViewPath("project/admin/views")
	//Admin.RegisterViewPath("project/admin/assets")

	return Admin, adminAuthHandle
}

func getDB() (db *gorm.DB, ok bool) {
	db = models.GetDB()
	validations.RegisterCallbacks(db)
	return db, true
}

func getAdminAuth(db *gorm.DB) (admin.Auth, http.Handler) {
	auth := clean.New(&auth.Config{
		DB:         db,
		UserModel:  models.Admin{},
		Redirector: adminAuth.AdminRedirector{},
	})
	a := adminAuth.AdminAuth{Auth: auth}
	h := auth.NewServeMux()

	return a, h
}
