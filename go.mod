module github.com/bug-tracker

go 1.16

replace github.com/bug-tracker/app => ./app

require (
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-contrib/secure v0.0.1
	github.com/gin-gonic/gin v1.7.4
	github.com/spf13/viper v1.8.1
	github.com/xanzy/go-gitlab v0.50.4
)
