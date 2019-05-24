package registry

import (
	"net/http"
	"encoding/json"
	"fmt"
	"errors"
	"io"
	"io/ioutil"
)

type Registry struct {
	url string
	client *http.Client
	auth struct {
		is_set bool
		username string
		password string
	}
}

type RegistryError struct {
	StatusCode int `json:-`
	Errors []struct {
		Code string `json:"code"`
		Message string `json:"message"`
		detail interface{} `json:"detail"`
	} `json:"errors"`
}

func (e *RegistryError) Error() string {
	switch len(e.Errors) {
	case 0:
		return "unexpected error, no extra information provided"
	case 1:
		return fmt.Sprintf("server returned error : %v", e.Errors[0])
	default:
		return fmt.Sprintf("multiple errors %v", e.Errors)
	}
}

func defaultErrorHandle(resp *http.Response) error {
	ret := RegistryError{}
	ret.StatusCode = resp.StatusCode
	if (ret.StatusCode >= 400) && (ret.StatusCode < 500) {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		_ = json.Unmarshal(body, &ret)
	}
	return &ret
}

func New(url string) *Registry {
	ret := Registry{}
	ret.url = url
	ret.client = &http.Client{}
	return &ret
}

func (self *Registry) SetBasicAuth(username string, password string) {
	self.auth.is_set = true
	self.auth.username = username
	self.auth.password = password
}

func (self *Registry) DoRequest(method string, url string, body io.Reader, extra_headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	if self.auth.is_set {
		req.SetBasicAuth(self.auth.username, self.auth.password)
	}
	if extra_headers != nil {
		for k, v := range extra_headers {
			req.Header.Add(k, v)
		}
	}
	resp, err := self.client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (self *Registry) CheckAPIVersionV2() error {
	resp, err := self.DoRequest("GET", self.url + "/v2/", nil, nil)
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case 200:
		return nil
	case 404:
		return errors.New("404 : server down or api unimplemented")
	default:
		return defaultErrorHandle(resp)
	}
}

func (self *Registry) CheckImageManifest(name string, reference string) (digest string, err error) {
	resp, err := self.DoRequest("HEAD", self.url + "/v2/" + name + "/manifests/" + reference, nil, map[string]string{"Accept": "application/vnd.docker.distribution.manifest.v2+json"})
	if err != nil {
		return "", err
	}
	switch resp.StatusCode {
	case 200:
		return resp.Header.Get("Docker-Content-Digest"), nil
	case 404:
		return "", errors.New("image not found")
	default:
		return "", defaultErrorHandle(resp)
	}
}

func (self *Registry) PullImageManifest(name string, reference string) (digest string, manifest interface{}, err error) {
	resp, err := self.DoRequest("GET", self.url + "/v2/" + name + "/manifests/" + reference, nil, map[string]string{"Accept": "application/vnd.docker.distribution.manifest.v2+json"})
	if err != nil {
		return "", nil, err
	}
	switch resp.StatusCode {
	case 200:
		defer resp.Body.Close()
		content_json, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", nil, err
		}
		var content interface{}
		err = json.Unmarshal(content_json, &content)
		if err != nil {
			return "", nil, err
		}
		return resp.Header.Get("Docker-Content-Digest"), content, nil
	default:
		return "", nil, defaultErrorHandle(resp)
	}
}

func (self *Registry) ListImageTags(name string) ([]string, error) {
	resp, err := self.DoRequest("GET", self.url + "/v2/" + name + "/tags/list", nil, nil)
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case 200:
		defer resp.Body.Close()
		content_json, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var content struct {
			Tags []string `json:"tags"`
		}
		err = json.Unmarshal(content_json, &content)
		if err != nil {
			return nil, err
		}
		return content.Tags, nil
	default:
		return nil, defaultErrorHandle(resp)
	}
}

func (self *Registry) DeleteImage(name string, digest string) error {
	resp, err := self.DoRequest("DELETE", self.url + "/v2/" + name + "/manifests/" + digest, nil, nil)
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case 202:
		return nil
	case 404:
		return nil // Already deleted
	default:
		return defaultErrorHandle(resp)
	}
}

