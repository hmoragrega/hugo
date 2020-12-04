// Copyright 2018 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package livereloadinject

import (
	"bytes"
	"fmt"
	"html"
	"net/url"
	"strings"

	"github.com/gohugoio/hugo/helpers"
	"github.com/gohugoio/hugo/transform"
)

type tag struct {
	markup       []byte
	appendScript bool
}

var tags = []tag{
	{markup: []byte("<head>"), appendScript: true},
	{markup: []byte("<HEAD>"), appendScript: true},
	{markup: []byte("</body>")},
	{markup: []byte("</BODY>")},
}

// New creates a function that can be used
// to inject a script tag for the livereload JavaScript in a HTML document.
func New(baseURL url.URL) transform.Transformer {
	return func(ft transform.FromTo) error {
		b := ft.From().Bytes()
		idx := -1
		var match tag
		// We used to insert the livereload script right before the closing body.
		// This does not work when combined with tools such as Turbolinks.
		// So we try to inject the script as early as possible.
		for _, t := range tags {
			idx = bytes.Index(b, t.markup)
			if idx != -1 {
				match = t
				break
			}
		}

		path := strings.TrimSuffix(baseURL.Path, "/")

		src := path + "/livereload.js?mindelay=10&v=2"
		if port := baseURL.Port(); port != "" {
			src += "&port=" + port
		}
		src += "&path=" + strings.TrimPrefix(path+"/livereload", "/")

		c := make([]byte, len(b))
		copy(c, b)

		if idx == -1 {
			_, err := ft.To().Write(c)
			return err
		}

		script := []byte(fmt.Sprintf(`<script src="%s" data-no-instant defer></script>`, html.EscapeString(src)))

		i := idx
		if match.appendScript {
			i += len(match.markup)
		}

		c = append(c[:i], append(script, c[i:]...)...)

		if _, err := ft.To().Write(c); err != nil {
			helpers.DistinctWarnLog.Println("Failed to inject LiveReload script:", err)
		}
		return nil
	}
}
