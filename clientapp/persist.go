package clientapp

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var pt Persister

func init() {
	pt = new(persist)
}

// Persister .....
type Persister interface {
	JSONPut(fn string, js []byte) error
	B642jsonPut(fname, inStr string) error
	If2B64(interface{}) (string, error)
	B642json(b64 string) ([]byte, error)
	FJsonUnMarshl(fname string, ift interface{}) error
}

type persist struct {
	sync.RWMutex
}

func (p *persist) JSONPut(fn string, js []byte) error {
	p.Lock()
	defer p.Unlock()
	var pj []byte
	if f, err := os.OpenFile(fn, os.O_WRONLY|os.O_TRUNC, 0666); err == nil {
		if sjb, err := simplejson.NewJson(js); err == nil {
			if pj, err = sjb.EncodePretty(); err == nil {
				f.Write(pj)
			} else {
				log.Error(err)
			}
		} else {
			log.Error(err)
		}
		f.Sync()
		f.Close()
	} else {
		log.Error(err)
	}
	if len(pj) == 0 {
		return errors.New("JsonPut failed")
	}
	return nil
}

func (p *persist) B642jsonPut(fn, inStr string) error {
	p.Lock()
	defer p.Unlock()
	var pj []byte
	if f, err := os.OpenFile(fn, os.O_WRONLY|os.O_TRUNC, 0666); err == nil {
		decoder := base64.StdEncoding.WithPadding('?')
		if ts, err := decoder.DecodeString(inStr); err == nil {
			if sjb, err := simplejson.NewJson(ts); err == nil {
				if pj, err = sjb.EncodePretty(); err == nil {
					f.Write(pj)
				}
			}
		}
		f.Sync()
		f.Close()
	}
	if len(pj) == 0 {
		return errors.New("B642jsonPut failed")
	}
	return nil
}

func (p *persist) B642json(b64 string) (js []byte, err error) {
	encoder := base64.StdEncoding.WithPadding('?')
	js, err = encoder.DecodeString(b64)
	return
}

func (p *persist) If2B64(ifs interface{}) (string, error) {
	b, err := json.Marshal(ifs)
	if err == nil {
		decoder := base64.StdEncoding.WithPadding('?')
		return decoder.EncodeToString(b), nil
	}
	return "", err
}

//FJson2B64Get ...
func (p *persist) FJsonUnMarshl(fn string, ift interface{}) error {
	var bs []byte
	p.RLock()
	defer p.RUnlock()
	f, err := os.OpenFile(fn, os.O_RDONLY, 0444)
	defer f.Close()
	if err == nil {
		if bs, err = ioutil.ReadAll(f); err == nil {
			return json.Unmarshal(bs, ift)
		}
	}
	if len(bs) == 0 {
		return errors.New("FJsonUnMarshl failed")
	}
	return nil
}

// CreateFile creat file
func CreateFile(fname string, fmode os.FileMode) (err error) {
	i := strings.LastIndex(fname, "/")
	basepath := string(fname[0 : i+1])
	_, existerr := os.Stat(basepath)
	if os.IsNotExist(existerr) {
		if err = os.MkdirAll(basepath, os.ModePerm); err != nil {
			//os.MkdirAll(basepath, os.ModePerm)
			err = errors.Wrap(err, "in CreateOrCoverFile func")
			return
		}
	}
	_, fserr := os.Stat(fname)
	if os.IsNotExist(fserr) {
		f, cferr := os.Create(fname)
		os.Chmod(fname, fmode)
		f.Sync()
		f.Close()
		err = errors.Wrap(cferr, "in CreateOrCoverFile func")
		if err != nil {
			return
		}
	}
	f, _ := os.OpenFile(fname, os.O_WRONLY, 0666)
	f.Sync()
	f.Close()
	return
}
