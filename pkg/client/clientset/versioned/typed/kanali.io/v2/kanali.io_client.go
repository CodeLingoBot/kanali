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
package v2

import (
	v2 "github.com/northwesternmutual/kanali/pkg/apis/kanali.io/v2"
	"github.com/northwesternmutual/kanali/pkg/client/clientset/versioned/scheme"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	rest "k8s.io/client-go/rest"
)

type KanaliV2Interface interface {
	RESTClient() rest.Interface
	ApiKeysGetter
	ApiKeyBindingsGetter
	ApiProxiesGetter
	MockTargetsGetter
}

// KanaliV2Client is used to interact with features provided by the kanali.io group.
type KanaliV2Client struct {
	restClient rest.Interface
}

func (c *KanaliV2Client) ApiKeys() ApiKeyInterface {
	return newApiKeys(c)
}

func (c *KanaliV2Client) ApiKeyBindings(namespace string) ApiKeyBindingInterface {
	return newApiKeyBindings(c, namespace)
}

func (c *KanaliV2Client) ApiProxies(namespace string) ApiProxyInterface {
	return newApiProxies(c, namespace)
}

func (c *KanaliV2Client) MockTargets(namespace string) MockTargetInterface {
	return newMockTargets(c, namespace)
}

// NewForConfig creates a new KanaliV2Client for the given config.
func NewForConfig(c *rest.Config) (*KanaliV2Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &KanaliV2Client{client}, nil
}

// NewForConfigOrDie creates a new KanaliV2Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *KanaliV2Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new KanaliV2Client for the given RESTClient.
func New(c rest.Interface) *KanaliV2Client {
	return &KanaliV2Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v2.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *KanaliV2Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
