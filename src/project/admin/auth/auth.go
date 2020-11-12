package auth

import (
	"log"
	"net/http"
	users "project/models"

	"github.com/qor/admin"
	"github.com/qor/auth"
	"github.com/qor/qor"
)

type AdminAuth struct {
	Auth *auth.Auth
}

type AdminRedirector struct {
}

func (AdminRedirector) Redirect(w http.ResponseWriter, req *http.Request, action string) {
	http.Redirect(w, req, "/admin", http.StatusSeeOther)
}

func (AdminAuth) LoginURL(c *admin.Context) string {
	return "/auth/login"
}

func (AdminAuth) LogoutURL(c *admin.Context) string {
	return "/auth/logout"
}

func (a AdminAuth) GetCurrentUser(c *admin.Context) qor.CurrentUser {
	currentUser := a.Auth.GetCurrentUser(c.Request)
	if currentUser != nil {
		qorCurrentUser, ok := currentUser.(qor.CurrentUser)
		if !ok {
			log.Printf("User %#v haven't implement qor.CurrentUser interface\n", currentUser)
		}
		if currentUser.(*users.Admin).Status == users.AdminStatusConfirmed {
			return qorCurrentUser
		}
	}
	return nil
}
