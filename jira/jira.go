package jira

import (
	"encoding/json"
	"net/http"
	"net/url"
)

// JiraAPI specifies how to communicate with the Jira API
type JiraAPI struct {
	// JiraUrl is required and specified the base url of the Jira instance in
	// the form "https://{jira-cloud-instance-name}.atlassian.net/"
	JiraUrl string

	// Email used for authenticating into Jira
	Email string

	// ApiToken used for authenticating to Jira
	ApiToken string
}

// JiraUser defines the user data parsed from Jira API
type JiraUser struct {
	AccountId string
}

// Issue
type Issue struct {
	Key string
	Url string `json:"self"`
	Fields struct {
		Summary string
	}
}

// GetCurrentUser returns the JiraUser user representation from the API
func (j JiraAPI) GetCurrentUser() (user JiraUser, err error) {
	request, err := http.NewRequest(http.MethodGet, j.getApiEndpoint("myself"), nil)
	if err != nil {
		return 
	}
	request.SetBasicAuth(j.Email, j.ApiToken)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return 
	}
	defer response.Body.Close()
	dec := json.NewDecoder(response.Body)
	err = dec.Decode(&user)
	return
}

func (j JiraAPI) GetWorkableIssues() (issues []Issue, err error) {	
	n, err := j.getTotalNumberOfIssues()
	if err != nil {
		return 
	}
	issuesChan := make(chan []Issue)
	errorChan := make(chan error)
	for i := 0; i < n; i+=100 {
		go j.getIssues(i, issuesChan, errorChan)
	}
	for i := 0; i < n; i+=100 {
		if err = <-errorChan; err != nil {
			return
		}
		issues = append(issues, <-issuesChan...)
	}
	return
}

func (j JiraAPI) getIssues(startAt int, issues chan []Issue, errors chan error) {
	v := url.Values{}
	v.Set("jql", `project in projectsWhereUserHasPermission("Work on issues")`)
	v.Set("fields", "summary")
	v.Set("maxResults", "100")
	request, err := http.NewRequest(http.MethodGet, j.getApiEndpoint("search?"+v.Encode()), nil)
	if err != nil {
		errors <- err
		return 
	}
	request.SetBasicAuth(j.Email, j.ApiToken)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		errors <- err
		return 
	}
	defer response.Body.Close()
	var jsonResponse struct {
		Issues []Issue
	}
	dec := json.NewDecoder(response.Body)
	errors <- dec.Decode(&jsonResponse)
	issues <- jsonResponse.Issues
	return
}

func (j JiraAPI) getTotalNumberOfIssues() (n int, err error) {
	v := url.Values{}
	v.Set("jql", `project in projectsWhereUserHasPermission("Work on issues")`)
	v.Set("maxResults", "0")
	request, err := http.NewRequest(http.MethodGet, j.getApiEndpoint("search?"+v.Encode()), nil)
	if err != nil {
		return 
	}
	request.SetBasicAuth(j.Email, j.ApiToken)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return 
	}
	defer response.Body.Close()
	var header struct {
		Total int
	}
	dec := json.NewDecoder(response.Body)
	err = dec.Decode(&header)
	n = header.Total
	return
}

func (j JiraAPI) getApiEndpoint(endpoint string) string {
	return j.JiraUrl + "/rest/api/3/" + endpoint
}