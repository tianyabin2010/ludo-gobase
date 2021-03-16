package cos

import (
	"errors"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	"net/url"
	"github.com/tencentyun/cos-go-sdk-v5"
	"context"
)

var (
	DefaultCosClient *cos.Client
)

func Init(addr, secretId, secretKey string) {
	DefaultCosClient = NewCosClient(addr, secretId, secretKey)
}

func NewCosClient(cosurl, secretId, secretKey string) *cos.Client {
	u, _ := url.Parse(cosurl)
	b := &cos.BaseURL{BucketURL: u}
	ret := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretId,
			SecretKey: secretKey,
		},
	})
	return ret
}

func GetDataFromCos(filepath string) []byte {
	if nil != DefaultCosClient {
		resp, err := DefaultCosClient.Object.Get(context.Background(), filepath, nil)
		if err != nil {
			log.Error().Err(err).Str("filepath", filepath).
				Msgf("GetDataFromCos cos object get oper error")
			return nil
		}
		if nil != resp && nil != resp.Body {
			defer resp.Body.Close()
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error().Err(err).Str("filepath", filepath).
				Msgf("GetDataFromCos cos resp readall error")
			return nil
		}
		return body
	}
	return nil
}

func PutDataToCos(filepath, cosfilepath string) error {
	if nil != DefaultCosClient {
		_, _, err := DefaultCosClient.Object.Upload(context.Background(), cosfilepath, filepath, nil)
		if err != nil {
			return err
		}
	}
	return errors.New("DefaultCosClient is nil")
}