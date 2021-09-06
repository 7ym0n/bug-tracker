package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/bug-tracker/app"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/xanzy/go-gitlab"
)

var (
	//go:embed templates/*
	fs           embed.FS
	AllowedHosts []string = []string{}
)

// initGitlab
func initGitlab() error {
	git, err := gitlab.NewClient(app.Config.AccessToken, gitlab.WithBaseURL(app.Config.Repos))
	if err != nil {
		return err
	}
	app.Gitlab = git
	return nil
}

// getConfig ...
func initConfig() error {
	viper.SetConfigName("config")                 // name of config file (without extension)
	viper.SetConfigType("yaml")                   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/" + app.NAME + "/") // path to look for the config file in
	viper.AddConfigPath("$HOME/.appname")         // call multiple times to add many search paths
	viper.AddConfigPath(".")                      // optionally look for config in the working directory
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if err := viper.Unmarshal(&app.Config); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := initConfig(); err != nil {
		fmt.Println("Init config failed.", err)
		return
	}

	if err := initGitlab(); err != nil {
		fmt.Println("Init gitlab failed.", err)
		return
	}

	router := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsAllowOrigins := []string{}
	secureAllowedHosts := []string{}
	for _, host := range AllowedHosts {
		host = strings.TrimPrefix(strings.TrimPrefix(host, "http://"), "https://")
		secureAllowedHosts = append(secureAllowedHosts, host)
		corsAllowOrigins = append(corsAllowOrigins, fmt.Sprintf("http://%s", host), fmt.Sprintf("https://%s", host))
	}
	if len(corsAllowOrigins) > 0 {
		corsConfig.AllowOrigins = corsAllowOrigins
	} else {
		corsConfig.AllowAllOrigins = true
	}

	router.Use(cors.New(corsConfig))
	router.Use(secure.New(secure.Config{
		AllowedHosts:       secureAllowedHosts,
		FrameDeny:          true,
		ContentTypeNosniff: true,
		BrowserXssFilter:   true,
		IENoOpen:           true,
		ReferrerPolicy:     "strict-origin-when-cross-origin",
		// 限制加载第三方站点资源，参考 https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Headers/Content-Security-Policy
		// ContentSecurityPolicy: "default-src 'self'",
	}))
	templ := template.Must(template.New("").ParseFS(fs, "templates/*.tmpl"))
	router.SetHTMLTemplate(templ)
	router.StaticFS("/public", http.FS(fs))

	// routes
	router.GET("/", func(c *gin.Context) {
		app.Render(c, "index", app.NewResponse(nil))
	})
	router.POST("/upload", app.Upload)
	router.POST("/comments", app.GetComments)
	router.POST("/members", app.GetMembers)
	router.GET("/issue/:pid/:id", app.GetIssue)
	router.GET("/show/:pid/:id", app.ShowIssue)
	router.POST("/issues", app.GetIssues)
	router.POST("/issue", app.CreateIssue)
	router.PUT("/issue", app.UpdateIssue)
	router.DELETE("/issue", app.RemoveIssue)
	router.Run(":80")
}
