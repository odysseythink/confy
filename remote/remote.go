package remote

import (
	"bytes"
	"io"
	"os"
	"strings"

	"mlib.com/confy"
)

type remoteConfigProvider struct{}

func (rc remoteConfigProvider) Get(rp confy.RemoteProvider) (io.Reader, error) {
	cm, err := getConfigManager(rp)
	if err != nil {
		return nil, err
	}
	b, err := cm.Get(rp.Path())
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

func (rc remoteConfigProvider) Watch(rp confy.RemoteProvider) (io.Reader, error) {
	cm, err := getConfigManager(rp)
	if err != nil {
		return nil, err
	}
	resp, err := cm.Get(rp.Path())
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(resp), nil
}

func (rc remoteConfigProvider) WatchChannel(rp confy.RemoteProvider) (<-chan *confy.RemoteResponse, chan bool) {
	cm, err := getConfigManager(rp)
	if err != nil {
		return nil, nil
	}
	quit := make(chan bool)
	quitwc := make(chan bool)
	viperResponsCh := make(chan *confy.RemoteResponse)
	cryptoResponseCh := cm.Watch(rp.Path(), quit)
	// need this function to convert the Channel response form Response to confy.Response
	go func(cr <-chan *Response, vr chan<- *confy.RemoteResponse, quitwc <-chan bool, quit chan<- bool) {
		for {
			select {
			case <-quitwc:
				quit <- true
				return
			case resp := <-cr:
				vr <- &confy.RemoteResponse{
					Error: resp.Error,
					Value: resp.Value,
				}
			}
		}
	}(cryptoResponseCh, viperResponsCh, quitwc, quit)

	return viperResponsCh, quitwc
}

func getConfigManager(rp confy.RemoteProvider) (ConfigManager, error) {
	var cm ConfigManager
	var err error

	endpoints := strings.Split(rp.Endpoint(), ";")
	if rp.SecretKeyring() != "" {
		var kr *os.File
		kr, err = os.Open(rp.SecretKeyring())
		if err != nil {
			return nil, err
		}
		defer kr.Close()
		switch rp.Provider() {
		case "etcd3":
			cm, err = NewEtcdV3ConfigManager(endpoints, kr)
		default:
			cm, err = NewEtcdV3ConfigManager(endpoints, kr)
		}
	} else {
		switch rp.Provider() {
		case "etcd3":
			cm, err = NewStandardEtcdV3ConfigManager(endpoints)
		default:
			cm, err = NewStandardEtcdV3ConfigManager(endpoints)
		}
	}
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func init() {
	confy.RemoteConfig = &remoteConfigProvider{}
}
