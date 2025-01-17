package server

import (
	"bytes"
	"encoding/gob"
	"github.com/jmartin82/mmock/internal/config"
	"github.com/jmartin82/mmock/pkg/mock"
	"log"

	"github.com/jmartin82/mmock/pkg/match"
)

//RequestResolver contains the functions to check the http request and return the matching mock.
type RequestResolver interface {
	Resolve(req *mock.Request) (*mock.Definition, *match.Result)
}

//NewRouter returns a pointer to new Router
func NewRouter(mapping config.Mapping, checker match.Matcher) *Router {
	return &Router{
		Mapping: mapping,
		Checker: checker,
	}
}

//Router checks http requesta and try to figure out what is the best mock for each one.
type Router struct {
	Mapping config.Mapping
	Checker match.Matcher
}

func (rr *Router) copy(src, dst *mock.Definition) {
	var mod bytes.Buffer
	enc := gob.NewEncoder(&mod)
	dec := gob.NewDecoder(&mod)
	err := enc.Encode(src)
	if err != nil {
		log.Fatal("encode error:", err)
	}
	err = dec.Decode(dst)
	if err != nil {
		log.Fatal("decode error:", err)
	}

}

//Route checks the request with all available mock definitions and return the matching mock for it.
func (rr *Router) Resolve(req *mock.Request) (*mock.Definition, *match.Result) {
	mocks := rr.Mapping.List()
	mLog := &match.Result{Found: false}
	mLog.Errors = make([]match.Error, 0, len(mocks))

	for _, m := range mocks {
		r, err := rr.Checker.Match(req, &m, true)
		if r {
			//we return a copy of it, not the config itself because we will working on it.
			md := mock.Definition{}
			rr.copy(&m, &md)
			mLog.Found = true
			mLog.URI = m.URI
			return &md, mLog
		}
		mLog.Errors = append(mLog.Errors, match.Error{URI: m.URI, Reason: err.Error()})
		if err != match.ErrPathNotMatch {
			log.Printf("Discarding mock: %s Reason: %s\n", m.URI, err.Error())
		}
	}
	return getNotFoundResult(), mLog
}

func getNotFoundResult() *mock.Definition {
	return &mock.Definition{Response: mock.Response{StatusCode: 404}}
}
