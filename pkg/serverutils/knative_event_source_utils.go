package serverutils

import (
	"encoding/json"
	"io"
	"net/http"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KnativeEventSourceSpec describes how a user wants their resource to appear
type KnativeEventSourceSpec struct {
	Group   string                                      `json:"group" protobuf:"bytes,1,opt,name=group"`
	Version string                                      `json:"version,omitempty" protobuf:"bytes,2,opt,name=version"`
	Names   apiextensions.CustomResourceDefinitionNames `json:"names" protobuf:"bytes,3,opt,name=names"`
}

// KnativeEventSourceMeta is metadata that all persisted resources must have, which includes all objects users must create
type KnativeEventSourceMeta struct {
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
}

// KnativeEventSourceDefinition represents a resource that should be exposed on the API server.
type KnativeEventSourceDefinition struct {
	metav1.TypeMeta        `json:",inline"`
	KnativeEventSourceMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	KnativeEventSourceSpec `json:"spec" protobuf:"bytes,2,opt,name=spec"`
}

// KnativeEventSourceList is a list of KnativeEventSourceDefinition objects.
type KnativeEventSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// items list individual KnativeEventSourceDefinition objects
	Items []KnativeEventSourceDefinition `json:"items" protobuf:"bytes,2,rep,name=items"`
}

type CounterWr struct {
	io.Writer
	Count int64
}

func (cw *CounterWr) Write(p []byte) (n int, err error) {
	n, err = cw.Writer.Write(p)
	cw.Count += int64(n)
	return
}

// KnativeEventSourceFilter shall filter partial metadata from knative event sources CRDs before propagating
func KnativeEventSourceFilter(w io.Writer, r io.Reader) (int64, error) {
	var knativeEventSourceList KnativeEventSourceList
	cw := &CounterWr{Writer: w}

	if err := json.NewDecoder(r).Decode(&knativeEventSourceList); err != nil {
		plog.Errorf("Knative Event Source CRD response deserialization failed: %s", err)
		SendResponse(w.(http.ResponseWriter), http.StatusInternalServerError, ApiError{Err: err.Error()})
		return 0, err
	}

	if err := json.NewEncoder(cw).Encode(knativeEventSourceList); err != nil {
		plog.Errorf("Knative Event Source CRD response serialization failed: %s", err)
		SendResponse(w.(http.ResponseWriter), http.StatusInternalServerError, ApiError{Err: err.Error()})
		return 0, err
	}

	plog.Printf("\n\nContent Length = %d\n", cw.Count)
	return cw.Count, nil
}
