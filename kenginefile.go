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

package webdav

import (
	"github.com/khulnasoft/kengine/v2/kengineconfig/kenginefile"
	"github.com/khulnasoft/kengine/v2/kengineconfig/httpkenginefile"
	"github.com/khulnasoft/kengine/v2/modules/kenginehttp"
)

func init() {
	httpkenginefile.RegisterHandlerDirective("webdav", parseWebdav)
}

// parseWebdav parses the Kenginefile tokens for the webdav directive.
func parseWebdav(h httpkenginefile.Helper) (kenginehttp.MiddlewareHandler, error) {
	wd := new(WebDAV)
	err := wd.UnmarshalKenginefile(h.Dispenser)
	if err != nil {
		return nil, err
	}
	return wd, nil
}

// UnmarshalKenginefile sets up the handler from Kenginefile tokens.
//
//	webdav [<matcher>] {
//	    root <path>
//	}
func (wd *WebDAV) UnmarshalKenginefile(d *kenginefile.Dispenser) error {
	for d.Next() {
		if d.NextArg() {
			return d.ArgErr()
		}
		for d.NextBlock(0) {
			switch d.Val() {
			case "root":
				if wd.Root != "" {
					return d.Err("root path already specified")
				}
				if !d.NextArg() {
					return d.ArgErr()
				}
				wd.Root = d.Val()
			case "prefix":
				if wd.Prefix != "" {
					return d.Err("prefix already specified")
				}
				if !d.NextArg() {
					return d.ArgErr()
				}

				wd.Prefix = d.Val()
			default:
				return d.Errf("unrecognized subdirective: %s", d.Val())
			}
		}
	}
	return nil
}
