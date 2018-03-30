// Copyright (c) 2017 Northwestern Mutual.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package tracer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"

	"github.com/northwesternmutual/kanali/cmd/kanali/app/options"
	"github.com/northwesternmutual/kanali/pkg/tags"
	"github.com/northwesternmutual/kanali/pkg/utils"
)

// StartSpan will create the and return the top level span
// that Kanali will use. If there exists a span context from
// an upstream app, the created span will be a child of this
// span. Otherwise, a new root span will be created.
func StartSpan(r *http.Request) opentracing.Span {
	tracer := opentracing.GlobalTracer()
	name := fmt.Sprintf("%s %s",
		r.Method,
		r.URL.EscapedPath(),
	)
	ctx, err := tracer.Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header),
	)
	if err != nil {
		return opentracing.StartSpan(name)
	}

	return opentracing.StartSpan(name, opentracing.ChildOf(ctx))
}

func HydrateSpanFromRequest(req *http.Request, span opentracing.Span) {
	if req == nil {
		span.SetTag(tags.HTTPRequest, nil)
		return
	}

	span.SetTag(tags.HTTPRequestMethod, req.Method)
	span.SetTag(tags.HTTPRequestURLPath, req.URL.EscapedPath())
	span.SetTag(tags.HTTPRequestURLHost, req.Host)

	if closerOne, closerTwo, err := dupReader(req.Body); err != nil {
		span.SetTag(tags.HTTPRequestBody, tags.Error)
	} else {
		buf, err := ioutil.ReadAll(closerOne)
		if err != nil {
			span.SetTag(tags.HTTPRequestBody, tags.Error)
		} else {
			span.SetTag(tags.HTTPRequestBody, string(buf))
		}
		req.Body = closerTwo
	}

	jsonHeaders, err := json.Marshal(omitHeaderValues(
		req.Header,
		viper.GetString(options.FlagProxyHeaderMaskValue.GetLong()),
		viper.GetStringSlice(options.FlagProxyMaskHeaderKeys.GetLong())...,
	))
	if err != nil {
		span.SetTag(tags.HTTPRequestHeaders, tags.Error)
	}
	span.SetTag(tags.HTTPRequestHeaders, string(jsonHeaders))

	jsonQuery, err := json.Marshal(req.URL.Query())
	if err != nil {
		span.SetTag(tags.HTTPRequestURLQuery, tags.Error)
	}
	span.SetTag(tags.HTTPRequestURLQuery, string(jsonQuery))
}

func HydrateSpanFromResponse(res *http.Response, span opentracing.Span) {
	if res == nil {
		span.SetTag(tags.HTTPResponse, nil)
		return
	}

	if closerOne, closerTwo, err := dupReader(res.Body); err != nil {
		span.SetTag(tags.HTTPResponseBody, tags.Error)
	} else {
		buf, err := ioutil.ReadAll(closerOne)
		if err != nil {
			span.SetTag(tags.HTTPResponseBody, tags.Error)
		} else {
			span.SetTag(tags.HTTPResponseBody, string(buf))
		}
		res.Body = closerTwo
	}

	jsonHeaders, err := json.Marshal(omitHeaderValues(
		res.Header,
		viper.GetString(options.FlagProxyHeaderMaskValue.GetLong()),
		viper.GetStringSlice(options.FlagProxyMaskHeaderKeys.GetLong())...,
	))
	if err != nil {
		span.SetTag(tags.HTTPResponseHeaders, tags.Error)
	}
	span.SetTag(tags.HTTPResponseHeaders, string(jsonHeaders))
	span.SetTag(tags.HTTPResponseStatusCode, res.StatusCode)
}

func dupReader(closer io.ReadCloser) (io.ReadCloser, io.ReadCloser, error) {

	buf, err := ioutil.ReadAll(closer)
	if err != nil {
		return nil, nil, err
	}

	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))

	return rdr1, rdr2, nil

}

func omitHeaderValues(h http.Header, msg string, keys ...string) http.Header {
	if h == nil {
		return http.Header{}
	}

	clone := utils.CloneHTTPHeader(h)
	for _, key := range keys {
		if clone.Get(key) != "" {
			clone.Set(key, msg)
		}
	}

	return clone
}
