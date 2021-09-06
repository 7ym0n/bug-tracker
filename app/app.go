package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xanzy/go-gitlab"
)

const (
	NAME = "bugtracker"
)

type Project struct {
	ID         int
	Name       string
	ProjectURL string
}

type config struct {
	Projects    []Project
	AccessToken string
	Repos       string
}

var (
	Title  string  = "BugTracker for gitflow"
	Config *config = &config{}
	Gitlab *gitlab.Client
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type Pagination struct {
	PerPage int `json:"perPage"`
	Page    int `json:"page"`
	Total   int `json:"total"`
}

type PageResponse struct {
	Response
	Pagination
}

type Results struct {
	Title  string
	Config *config
	Data   interface{}
}

// NewResponse ...
func NewResponse(data interface{}) Results {
	return Results{
		Title:  Title,
		Config: Config,
		Data:   data,
	}
}

// Render ...
func Render(c *gin.Context, name string, data interface{}) {
	c.HTML(http.StatusOK, "layout.tmpl", gin.H{
		"bugtrack":  data,
		"container": name,
	})
}
