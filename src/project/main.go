package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rosberry/ginlog"

	"project/admin"
	"project/bindatafs"
	cm "project/common"
	"project/config"
	"project/controllers/http/temp"
	"project/models"
)

// Debug mode flag
var debugMode bool

func main() {
	cmdLine := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	compileTemplate := cmdLine.Bool("compile-templates", false, "Compile Templates")
	cmdLine.Parse(os.Args[1:])

	switch config.App.Mode {
	case config.ModeRelease:
		gin.SetMode(gin.ReleaseMode)
	case config.ModeDebug:
		debugMode = true
	}

	db := models.GetDB()
	_ = db

	router := gin.New()
	router.Use(
		gin.Recovery(),
	)

	system := router.Group("/")
	{
		system.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, &cm.EmptyResponse{Result: true})
		})
	}

	api := router.Group("/v1/", ginlog.Logger(debugMode), addCORSHeaders)
	{
		// Return Status Ok for all OPTIONS requests, required for CORS
		api.OPTIONS("/*path", func(c *gin.Context) {
			c.JSON(http.StatusOK, &cm.EmptyResponse{Result: false})
		})

		api.POST("/auth", temp.Pong) //users.Auth)

		auth := api.Group("", authRequired)
		{
			auth.GET("/ping", temp.Pong)
		}
	}

	//router.Run(config.App.Backend.Listen)
	admin.HandleGin(router)

	if *compileTemplate {
		bindatafs.AssetFS.Compile()
	} else {
		router.Run(config.App.Backend.Listen)
	}
}

func authRequired(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := authHeader[7:]
			if len(token) > 0 {
				if device := models.GetDeviceByToken(token); device != nil {
					if user := models.GetUser(device.UserID); user != nil {
						if debugMode {
							ginlog.AddDebugValue(c, fmt.Sprintf("User:%d", user.ID))
						}
						c.Set("Device", device)
						c.Set("User", user)
						c.Next()
						return
					}
				}
			}
		}
	}
	c.JSON(http.StatusUnauthorized, cm.Error[cm.ErrNotAuth])
	c.Abort()
}

func addCORSHeaders(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, PATCH")
	c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, dpi")
	c.Next()
}
