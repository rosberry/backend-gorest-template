package clean

import (
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/qor/auth"
	"github.com/qor/auth/auth_identity"
	"github.com/qor/auth/claims"
	"github.com/qor/auth/providers/password"
	"github.com/qor/i18n"
	"github.com/qor/i18n/backends/database"
	"github.com/qor/i18n/backends/yaml"
	"github.com/qor/qor"
	"github.com/qor/qor/utils"
	"github.com/qor/render"
	"github.com/qor/session"
)

// ErrPasswordConfirmationNotMatch password confirmation not match error
var ErrPasswordConfirmationNotMatch = errors.New("password confirmation doesn't match password")
var ErrOnReview = errors.New("you account is on review")

// New initialize clean theme
func New(config *auth.Config) *auth.Auth {
	if config == nil {
		config = &auth.Config{}
	}
	config.ViewPaths = append(config.ViewPaths, "auth_themes/clean/views")

	if config.DB == nil {
		log.Print("Please configure *gorm.DB for Auth theme clean")
	}

	if config.Render == nil {
		yamlBackend := yaml.New()
		databaseBackend := database.New(config.DB)
		I18n := i18n.New(databaseBackend)
		for idx, gopath := range append([]string{filepath.Join(utils.AppRoot, "vendor")}, utils.GOPATH()...) {
			var filePath string
			if idx > 0 {
				filePath = filepath.Join(gopath, "src", "project", "auth_themes/clean/locales/en-US.yml") // XXX added project, removed github...
			} else {
				filePath = filepath.Join(gopath, "auth_themes/clean/locales/en-US.yml")
			}
			if content, err := ioutil.ReadFile(filePath); err == nil {
				translations, _ := yamlBackend.LoadYAMLContent(content)
				for _, translation := range translations {
					I18n.SaveTranslation(translation)
				}
				break
			}
		}

		config.Render = render.New(&render.Config{
			FuncMapMaker: func(render *render.Render, req *http.Request, w http.ResponseWriter) template.FuncMap {
				return template.FuncMap{
					"t": func(key string, args ...interface{}) template.HTML {
						return I18n.T(utils.GetLocale(&qor.Context{Request: req}), key, args...)
					},
				}
			},
		})
	}

	Auth := auth.New(config)

	Auth.RegisterProvider(password.New(&password.Config{
		Confirmable:      false,
		AuthorizeHandler: AuthorizeHandler,
		RegisterHandler:  RegisterHandler,
	}))

	if Auth.Config.DB != nil {
		// Migrate Auth Identity model
		Auth.Config.DB.AutoMigrate(Auth.Config.AuthIdentityModel)
	}
	return Auth
}

//AuthorizeHandler is a replacement for DefaultAuthorizeHandler. It contains fixes of original password package
func AuthorizeHandler(context *auth.Context) (*claims.Claims, error) {
	var (
		authInfo    auth_identity.Basic
		req         = context.Request
		tx          = context.Auth.GetDB(req)
		provider, _ = context.Provider.(*password.Provider)
	)

	req.ParseForm()
	authInfo.Provider = provider.GetName()
	authInfo.UID = strings.TrimSpace(req.Form.Get("login"))

	if tx.Model(context.Auth.Config.AuthIdentityModel).Where(
		map[string]interface{}{
			"provider": authInfo.Provider,
			"uid":      authInfo.UID,
		}).Scan(&authInfo).RecordNotFound() {
		return nil, auth.ErrInvalidAccount
	}

	if provider.Config.Confirmable && authInfo.ConfirmedAt == nil {
		currentUser, _ := context.Auth.UserStorer.Get(authInfo.ToClaims(), context)
		provider.Config.ConfirmMailer(authInfo.UID, context, authInfo.ToClaims(), currentUser)

		return nil, password.ErrUnconfirmed
	}

	if err := provider.Encryptor.Compare(authInfo.EncryptedPassword, strings.TrimSpace(req.Form.Get("password"))); err == nil {
		return authInfo.ToClaims(), err
	}

	return nil, auth.ErrInvalidPassword
}

//RegisterHandler is a replacement for DefaultRegisterHandler. It contains fixes of original password package
func RegisterHandler(context *auth.Context) (*claims.Claims, error) {
	context.Request.ParseForm()

	if context.Request.Form.Get("confirm_password") != context.Request.Form.Get("password") {
		return nil, ErrPasswordConfirmationNotMatch
	}

	var (
		err         error
		currentUser interface{}
		schema      auth.Schema
		authInfo    auth_identity.Basic
		req         = context.Request
		tx          = context.Auth.GetDB(req)
		provider, _ = context.Provider.(*password.Provider)
	)

	req.ParseForm()
	if req.Form.Get("login") == "" {
		return nil, auth.ErrInvalidAccount
	}

	if req.Form.Get("password") == "" {
		return nil, auth.ErrInvalidPassword
	}

	authInfo.Provider = provider.GetName()
	authInfo.UID = strings.TrimSpace(req.Form.Get("login"))

	if !tx.Model(context.Auth.Config.AuthIdentityModel).Where(
		map[string]interface{}{
			"provider": authInfo.Provider,
			"uid":      authInfo.UID,
		}).Scan(&authInfo).RecordNotFound() {
		return nil, auth.ErrInvalidAccount
	}

	if authInfo.EncryptedPassword, err = provider.Encryptor.Digest(strings.TrimSpace(req.Form.Get("password"))); err == nil {
		schema.Provider = authInfo.Provider
		schema.UID = authInfo.UID
		schema.Email = authInfo.UID
		schema.RawInfo = req

		currentUser, authInfo.UserID, err = context.Auth.UserStorer.Save(&schema, context)
		if err != nil {
			return nil, err
		}
		log.Printf("%+v", authInfo)
		// create auth identity
		authIdentity := reflect.New(utils.ModelType(context.Auth.Config.AuthIdentityModel)).Interface()
		if err = tx.Where(map[string]interface{}{
			"provider":           authInfo.Provider,
			"uid":                authInfo.UID,
			"encrypted_password": authInfo.EncryptedPassword,
			"user_id":            authInfo.UserID,
		}).FirstOrCreate(authIdentity).Error; err == nil {
			if provider.Config.Confirmable {
				context.SessionStorer.Flash(context.Writer, req, session.Message{Message: password.ConfirmFlashMessage, Type: "success"})
				err = provider.Config.ConfirmMailer(schema.Email, context, authInfo.ToClaims(), currentUser)
			}

			return nil, ErrOnReview
		}
	}

	return nil, err
}
