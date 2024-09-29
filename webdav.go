// Copyright 2015 NxPKG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package webdav implements a WebDAV server handler module for Kengine.
//
// Derived from work by Henrique Dias: https://github.com/hacdias/kengine-webdav
package webdav

import (
	"context"
	"errors"
	"io/fs"
	"net/http"

	"github.com/khulnasoft/kengine/v2"
	"github.com/khulnasoft/kengine/v2/kengineconfig/kenginefile"
	"github.com/khulnasoft/kengine/v2/modules/kenginehttp"
	"go.uber.org/zap"
	"golang.org/x/net/webdav"
)

func init() {
	kengine.RegisterModule(WebDAV{})
}

// WebDAV implements an HTTP handler for responding to WebDAV clients.
type WebDAV struct {
	// The root directory out of which to serve files. If
	// not specified, `{http.vars.root}` will be used if
	// set; otherwise, the current directory is assumed.
	// Accepts placeholders.
	Root string `json:"root,omitempty"`

	// The base path prefix used to access the WebDAV share.
	// Should be used if one more more matchers are used with the
	// webdav directive and it's needed to let the webdav share know
	// what the request base path will be.
	// For example:
	// webdav /some/path/match/* {
	//   root /path
	//   prefix /some/path/match
	// }
	// Accepts placeholders.
	Prefix string `json:"prefix,omitempty"`

	lockSystem webdav.LockSystem
	logger     *zap.Logger
}

// KengineModule returns the Kengine module information.
func (WebDAV) KengineModule() kengine.ModuleInfo {
	return kengine.ModuleInfo{
		ID:  "http.handlers.webdav",
		New: func() kengine.Module { return new(WebDAV) },
	}
}

// Provision sets up the module.
func (wd *WebDAV) Provision(ctx kengine.Context) error {
	wd.logger = ctx.Logger(wd)

	wd.lockSystem = webdav.NewMemLS()
	if wd.Root == "" {
		wd.Root = "{http.vars.root}"
	}

	return nil
}

func (wd WebDAV) ServeHTTP(w http.ResponseWriter, r *http.Request, next kenginehttp.Handler) error {
	// TODO: integrate with kengine 2's existing auth features to enforce read-only?
	// read methods: GET, HEAD, OPTIONS
	// write methods: POST, PUT, PATCH, DELETE, COPY, MKCOL, MOVE, PROPPATCH

	repl := r.Context().Value(kengine.ReplacerCtxKey).(*kengine.Replacer)
	root := repl.ReplaceAll(wd.Root, ".")
	prefix := repl.ReplaceAll(wd.Prefix, "")

	wdHandler := webdav.Handler{
		Prefix:     prefix,
		FileSystem: webdav.Dir(root),
		LockSystem: wd.lockSystem,
		Logger: func(req *http.Request, err error) {
			if err == nil {
				return
			}
			// ignore errors about non-exsiting files
			if errors.Is(err, fs.ErrNotExist) {
				return
			}

			wd.logger.Error("internal handler error",
				zap.Error(err),
				zap.Object("request", kenginehttp.LoggableHTTPRequest{Request: req}),
			)
		},
	}

	// Excerpt from RFC4918, section 9.4:
	//
	//     GET, when applied to a collection, may return the contents of an
	//     "index.html" resource, a human-readable view of the contents of
	//     the collection, or something else altogether.
	//
	// GET and HEAD, when applied to a collection, will behave the same as PROPFIND method.
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		info, err := wdHandler.FileSystem.Stat(context.TODO(), r.URL.Path)
		if err == nil && info.IsDir() {
			r.Method = "PROPFIND"
			if r.Header.Get("Depth") == "" {
				r.Header.Add("Depth", "1")
			}
		}
	}

	if r.Method == http.MethodHead {
		w = emptyBodyResponseWriter{w}
	}

	wdHandler.ServeHTTP(w, r)

	return nil
}

// emptyBodyResponseWriter is a response writer that does not write a body.
type emptyBodyResponseWriter struct{ http.ResponseWriter }

func (w emptyBodyResponseWriter) Write(data []byte) (int, error) { return 0, nil }

// Interface guards
var (
	_ kenginehttp.MiddlewareHandler = (*WebDAV)(nil)
	_ kenginefile.Unmarshaler       = (*WebDAV)(nil)
)
