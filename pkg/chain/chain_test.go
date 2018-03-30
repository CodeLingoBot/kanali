// Copyright (c) 2018 Northwestern Mutual.
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

package chain

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	assert.Equal(t, &Chain{}, New())
}

func TestAdd(t *testing.T) {
	c := New()
	fakeFunc := func(next http.Handler) http.Handler { return next }
	assert.Equal(t, 0, len(*c.Add()))
	assert.Equal(t, 0, len(*c.Add(nil)))
	assert.Equal(t, 1, len(*c.Add(fakeFunc)))
	assert.Equal(t, 3, len(*c.Add(fakeFunc, fakeFunc)))
}

func TestLink(t *testing.T) {
	getCurrentCalculation := func(t *testing.T, ctx context.Context) int {
		v, ok := ctx.Value("calculation").(int)
		if !ok {
			assert.Fail(t, "")
		}
		return v
	}

	outer, middle, inner := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(context.Background(), "calculation", 25)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "calculation", getCurrentCalculation(t, r.Context())/5)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, getCurrentCalculation(t, r.Context()), 5)
	})

	req, _ := http.NewRequest("GET", "/foo", nil)
	handler, _ := New().Add(outer, middle).Link(inner).(http.HandlerFunc)
	handler.ServeHTTP(nil, req)

	var c *Chain
	assert.Nil(t, c.Link(nil))
	assert.Nil(t, New().Link(inner))
}
