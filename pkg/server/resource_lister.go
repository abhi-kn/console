package server

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/openshift/console/pkg/serverutils"
)

// ResourceLister ...
type ResourceLister interface {
	HandleResources(w http.ResponseWriter, r *http.Request)
}

// FilterFunction shall filter response before propagating
type FilterFunction func(io.Writer, io.Reader) (int64, error)

// resourceLister determines the list of resources of a particular kind
type resourceLister struct {
	bearerToken    string
	requestURL     *url.URL
	client         *http.Client
	responseFilter FilterFunction
}

func (l *resourceLister) HandleResources(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		serverutils.SendResponse(w, http.StatusMethodNotAllowed, serverutils.ApiError{Err: "invalid method: only GET is allowed"})
		return
	}

	req, err := http.NewRequest("GET", l.requestURL.String(), nil)
	if err != nil {
		serverutils.SendResponse(w, http.StatusInternalServerError, serverutils.ApiError{Err: fmt.Sprintf("failed to create GET request: %v", err)})
		return
	}

	req.Header.Set("Authorization", "Bearer "+l.bearerToken)
	resp, err := l.client.Do(req)
	if err != nil {
		serverutils.SendResponse(w, http.StatusBadGateway, serverutils.ApiError{Err: fmt.Sprintf("GET request failed: %v", err)})
		return
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("console service account cannot list resource: %s", resp.Status)
		serverutils.SendResponse(w, resp.StatusCode, serverutils.ApiError{Err: err.Error()})
		return
	}

	w.WriteHeader(resp.StatusCode)
	l.responseFilter(w, resp.Body)
	resp.Body.Close()
}

// NewResourceLister ....
func NewResourceLister(bearerToken string, requestURL *url.URL, client *http.Client, respFilter FilterFunction) ResourceLister {
	r := &resourceLister{
		bearerToken:    bearerToken,
		requestURL:     requestURL,
		client:         client,
		responseFilter: respFilter,
	}
	if r.responseFilter == nil {
		r.responseFilter = io.Copy
	}

	return r
}
