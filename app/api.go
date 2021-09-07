package app

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xanzy/go-gitlab"
)

// Error ...
func Error(c *gin.Context, code int, err error) {
	c.AbortWithStatusJSON(http.StatusOK, Response{
		Code: code,
		Msg:  err.Error(),
		Data: nil,
	})
}

type ListIssueOptions struct {
	ProjectID int
	Options   gitlab.ListProjectIssuesOptions
}

type IssueAndProjectID struct {
	ProjectID int
	IssueID   int
}

func Upload(c *gin.Context) {
	var (
		err error
		pid int
	)
	_pid, ok := c.GetQuery("pid")
	if ok {
		pid, err = strconv.Atoi(_pid)
		if err != nil {
			Error(c, 500, err)
			return
		}
	}
	if !CheckProject(pid) {
		Error(c, 403, errors.New("No permission."))
		return
	}
	var projectUrl string
	for _, p := range Config.Projects {
		if p.ID == pid {
			projectUrl = strings.TrimSuffix(p.ProjectURL, "/")
		}
	}

	uploads := []*gitlab.ProjectFile{}
	form, err := c.MultipartForm()
	if err != nil {
		Error(c, 500, err)
		return
	}
	files := form.File["issue-attach[]"]
	for _, file := range files {

		fileNameInt := time.Now().Unix()
		fileNameStr := strconv.FormatInt(fileNameInt, 10)
		ext := path.Ext(file.Filename)
		dst := path.Join("upload", "/", fileNameStr+ext)
		err = c.SaveUploadedFile(file, dst)
		if err != nil {
			fmt.Println(err)
		}
		upload, _, err := Gitlab.Projects.UploadFile(pid, dst, nil)
		if err != nil {
			fmt.Println("upload " + dst + " file failed.")
		}
		err = os.Remove(dst) //删除文件test.txt
		if err != nil {
			fmt.Println("remove " + dst + "file failed")
			Error(c, 500, err)
			return
		}
		upload.Markdown = strings.ReplaceAll(upload.Markdown, upload.URL, projectUrl+upload.URL)
		upload.URL = strings.ReplaceAll(upload.URL, upload.URL, projectUrl+upload.URL)
		uploads = append(uploads, upload)
	}
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "",
		Data: uploads,
	})
}

func GetIssue(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, 500, errors.New("id must be is integer."))
		return
	}
	pid, err := strconv.Atoi(c.Param("pid"))
	if err != nil {
		Error(c, 500, errors.New("pid must be is integer."))
		return
	}

	if !CheckProject(pid) {
		Error(c, 403, errors.New("No permission."))
		return
	}

	issue, _, err := Gitlab.Issues.GetIssue(pid, id, nil)
	if err != nil {
		Error(c, 500, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "",
		Data: issue,
	})
}

type ProjectsIssueComments struct {
	IssueAndProjectID
	Options *gitlab.ListIssueNotesOptions
}

type ProjectMembers struct {
	ProjectID int
	Options   *gitlab.ListProjectMembersOptions
}

// GetMembers ...
func GetMembers(c *gin.Context) {
	var projectMembers ProjectMembers
	err := c.ShouldBindJSON(&projectMembers)
	if err != nil {
		Error(c, 500, errors.New("Params error."))
	}

	if !CheckProject(projectMembers.ProjectID) {
		Error(c, 403, errors.New("No permission."))
		return
	}
	members, response, err := Gitlab.ProjectMembers.ListAllProjectMembers(projectMembers.ProjectID, projectMembers.Options, nil)
	if err != nil {
		Error(c, 500, err)
		return
	}

	c.JSON(http.StatusOK, PageResponse{
		Response: Response{
			Code: 0,
			Msg:  "",
			Data: members,
		},
		Pagination: Pagination{
			ItemsPerPage: response.ItemsPerPage,
			CurrentPage:  response.CurrentPage,
			TotalItems:   response.TotalItems,
			TotalPages:   response.TotalPages,
			NextPage:     response.NextPage,
			PreviousPage: response.PreviousPage,
		},
	})
}

// GetComments ...
func GetComments(c *gin.Context) {
	var projectsIssueComments ProjectsIssueComments
	if err := c.ShouldBindJSON(&projectsIssueComments); err != nil {
		Error(c, 500, errors.New("Params error."))
		return
	}

	if !CheckProject(projectsIssueComments.ProjectID) {
		Error(c, 403, errors.New("No permission."))
		return
	}
	notes, response, err := Gitlab.Notes.ListIssueNotes(projectsIssueComments.ProjectID,
		projectsIssueComments.IssueID,
		projectsIssueComments.Options,
		nil)
	if err != nil {
		Error(c, 500, err)
		return
	}

	c.JSON(http.StatusOK, PageResponse{
		Response: Response{
			Code: 0,
			Msg:  "",
			Data: notes,
		},
		Pagination: Pagination{
			ItemsPerPage: response.ItemsPerPage,
			CurrentPage:  response.CurrentPage,
			TotalItems:   response.TotalItems,
			TotalPages:   response.TotalPages,
			NextPage:     response.NextPage,
			PreviousPage: response.PreviousPage,
		},
	})
}

// ShowIssue ...
func ShowIssue(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, 500, errors.New("id must be is integer."))
		return
	}
	pid, err := strconv.Atoi(c.Param("pid"))
	if err != nil {
		Error(c, 500, errors.New("pid must be is integer."))
		return
	}

	if !CheckProject(pid) {
		Error(c, 403, errors.New("No permission."))
		return
	}

	issue, _, err := Gitlab.Issues.GetIssue(pid, id, nil)
	if err != nil {
		Error(c, 500, err)
		return
	}

	Render(c, "show", NewResponse(issue))
}

// GetIssues ...
func GetIssues(c *gin.Context) {
	var issueOptions ListIssueOptions
	if err := c.ShouldBindJSON(&issueOptions); err != nil {
		Error(c, 404, errors.New("Params error."))
		return
	}

	if !CheckProject(issueOptions.ProjectID) {
		Error(c, 403, errors.New("No permission."))
		return
	}

	issues, response, err := Gitlab.Issues.ListProjectIssues(issueOptions.ProjectID, &issueOptions.Options, nil)
	if err != nil {
		Error(c, 500, err)
		return
	}

	// for _, issue := range issues {
	// 	realNameAndTitle := strings.Split(issue.Title, "-")
	// 	if len(realNameAndTitle) >= 2 {
	// 		issue.Author.Username = realNameAndTitle[0]
	// 		issue.Title = strings.Join(realNameAndTitle[1:], "")
	// 	}
	// }

	c.JSON(http.StatusOK, PageResponse{
		Response{
			Code: 0,
			Msg:  "",
			Data: issues,
		},
		Pagination{
			ItemsPerPage: response.ItemsPerPage,
			CurrentPage:  response.CurrentPage,
			TotalItems:   response.TotalItems,
			TotalPages:   response.TotalPages,
			NextPage:     response.NextPage,
			PreviousPage: response.PreviousPage,
		},
	})
}

type CreateIssueOptions struct {
	ProjectID int
	Options   gitlab.CreateIssueOptions
}

// CreateIssue ...
func CreateIssue(c *gin.Context) {

	var issueOptions CreateIssueOptions
	if err := c.ShouldBindJSON(&issueOptions); err != nil {
		Error(c, 404, errors.New("Params error."))
		return
	}

	if !CheckProject(issueOptions.ProjectID) {
		Error(c, 403, errors.New("No permission."))
		return
	}
	issue, _, err := Gitlab.Issues.CreateIssue(issueOptions.ProjectID, &issueOptions.Options, nil)
	if err != nil {
		Error(c, 500, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "",
		Data: issue,
	})
}

// RemoveIssue ...
func RemoveIssue(c *gin.Context) {
	var (
		pid int
		id  int
		err error
	)
	_pid, ok := c.GetQuery("pid")
	if ok {
		pid, err = strconv.Atoi(_pid)
		if err != nil {
			Error(c, 500, err)
			return
		}
	} else {
		Error(c, 500, err)
		return
	}
	_id, ok := c.GetQuery("id")
	if ok {
		id, err = strconv.Atoi(_id)
		if err != nil {
			Error(c, 500, err)
			return
		}
	} else {
		Error(c, 500, err)
		return
	}

	if !CheckProject(pid) {
		Error(c, 403, errors.New("No permission."))
		return
	}
	if _, err := Gitlab.Issues.DeleteIssue(pid, id, nil); err != nil {
		Error(c, 500, err)
		return
	}
}

type UpdateIssueOptions struct {
	ProjectID int
	Options   gitlab.UpdateIssueOptions
	IssueID   int
}

func UpdateIssue(c *gin.Context) {
	var updateIssue UpdateIssueOptions
	if err := c.ShouldBindJSON(&updateIssue); err != nil {
		Error(c, 404, errors.New("Params error."))
		return
	}

	if !CheckProject(updateIssue.ProjectID) {
		Error(c, 403, errors.New("No permission."))
		return
	}

	oldIssue, _, err := Gitlab.Issues.GetIssue(updateIssue.ProjectID, updateIssue.IssueID, nil)
	if err != nil {
		Error(c, 500, err)
		return
	}
	if strings.ToLower(oldIssue.State) == "closed" {
		Error(c, 302, errors.New("closed issue don't modify."))
		return
	}
	issue, _, err := Gitlab.Issues.UpdateIssue(updateIssue.ProjectID, updateIssue.IssueID, &updateIssue.Options, nil)
	if err != nil {
		Error(c, 500, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "",
		Data: issue,
	})
}
