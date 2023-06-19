package network

import (
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"
)

var globalTransport *http.Transport
var timeout time.Duration

var (
	// DefaultLocalAddr is the default local IP address an Attacker uses.
	DefaultLocalAddr = net.IPAddr{IP: net.IPv4zero}
	// DefaultTLSConfig is the default tls.Config an Attacker uses.
	DefaultTLSConfig = &tls.Config{InsecureSkipVerify: true}
	// DefaultConnections is the default amount of max open idle connections per
	// target host.
	DefaultConnections = 10000
	// DefaultMaxConnections is the default amount of connections per target
	// host.
	DefaultMaxConnections = 0

	pool = &sync.Pool{
		New: func() interface{} {
			return &http.Client{
				Timeout: 120 * time.Second,
				Transport: &http.Transport{
					TLSHandshakeTimeout: 120 * time.Second,
					MaxIdleConnsPerHost: DefaultConnections,
				},
			}
		},
	}
)

func init() {
	timeout = 0
	globalTransport = &http.Transport{
		DisableKeepAlives:     true,
		ResponseHeaderTimeout: timeout,
		MaxIdleConnsPerHost:   100000,
		IdleConnTimeout:       timeout,
		TLSHandshakeTimeout:   timeout,
	}
}

func DoGetRequest(url string) (response []byte, err error) {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
func retry(url string, data interface{}, token string, attempts int, sleep time.Duration, fn func(string, interface{}, string) ([]byte, error)) ([]byte, error) {
	b, err := fn(url, data, token)
	if err != nil {
		if attempts--; attempts > 0 {
			time.Sleep(sleep)
			return retry(url, data, token, attempts, 2*sleep, fn)
		}
		return nil, err
	}
	return b, nil
}

// application/json; charset=utf-8
func Post2Api(url string, data interface{}, token string) (content []byte, err error) {
	c, err := postLogic(url, data, token)
	return c, utils.Wrap(err, " post")
	// return retry(url, data, token, 1, 10*time.Second, postLogic)
}

func Post2ApiWithoutAlives(url string, data interface{}, token string) (content []byte, err error) {
	c, err := postLogicWithoutAlives(url, data, token)
	return c, utils.Wrap(err, " post")
	// return retry(url, data, token, 1, 10*time.Second, postLogic)
}

func Post2ApiForRead(url string, data interface{}, token string) (content []byte, err error) {
	return retry(url, data, token, 3, 10*time.Second, postLogic)
}

func postLogic(url string, data interface{}, token string) (content []byte, err error) {
	jsonStr, err := json.Marshal(data)
	if err != nil {
		return nil, utils.Wrap(err, "marshal failed, url")
	}

	timeout := 120 * time.Second
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, utils.Wrap(err, "newRequest failed, url")
	}
	req.Close = true
	req.Header.Add("content-type", "application/json")
	req.Header.Add("token", token)
	tp := &http.Transport{
		DialContext: (&net.Dialer{
			KeepAlive: 10 * time.Minute,
		}).DialContext,
		ResponseHeaderTimeout: timeout,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       timeout,
		TLSHandshakeTimeout:   timeout,
	}
	client := &http.Client{Timeout: timeout, Transport: tp}
	resp, err := client.Do(req)
	if err != nil {
		return nil, utils.Wrap(err, "client.Do failed, url")
	}

	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, utils.Wrap(err, "ioutil.ReadAll failed, url")
	}
	if resp.StatusCode != 200 {
		return result, utils.Wrap(errors.New(resp.Status), "status code failed "+url)
	}
	//	fmt.Println(url, "Marshal data: ", string(jsonStr), string(result))
	return result, nil
}

func postLogicWithoutAlives(url string, data interface{}, token string) (content []byte, err error) {
	jsonStr, err := json.Marshal(data)
	if err != nil {
		return nil, utils.Wrap(err, "marshal failed, url")
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, utils.Wrap(err, "newRequest failed, url")
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("token", token)
	req.Header.Add("Accept-Encoding", "identity")
	req.Close = true
	//client := &http.Client{Timeout: timeout, Transport: globalTransport}

	//resp, err := http.DefaultClient.Do(req)
	//client := http.Client{
	//	Timeout:   timeout,
	//	Transport: globalTransport,
	//}
	client := pool.Get().(*http.Client)
	defer pool.Put(client)
	resp, err := client.Do(req)
	if err != nil {
		return nil, utils.Wrap(err, "client.Do failed, url")
	}
	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, utils.Wrap(err, "ioutil.ReadAll failed, url")
	}
	if resp.StatusCode != 200 {
		return result, utils.Wrap(errors.New(resp.Status), "status code failed "+url)
	}
	//	fmt.Println(url, "Marshal data: ", string(jsonStr), string(result))
	return result, nil
}
