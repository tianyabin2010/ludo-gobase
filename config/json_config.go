package config

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
)

func LoadJsonConfig(filePath string, dst interface{}) error {
	f, err := os.Open(filePath)
	if err != nil {
		log.Error().Err(err).Str("filePath", filePath).
			Msgf("LoadJsonConfig open file error")
		return err
	}
	defer f.Close()
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		log.Error().Err(err).Str("filePath", filePath).
			Msgf("LoadJsonConfig read file error")
		return err
	}
	err = json.Unmarshal(bs, dst)
	if err != nil {
		log.Error().Err(err).Str("filePath", filePath).
			Msgf("LoadJsonConfig json unmarshal error")
		return err
	}
	return nil
}

func GetFileLastModTime(filepath string) int64 {
	fi, err := os.Stat(filepath)
	if err != nil {
		log.Error().Err(err).
			Msgf("GetFileLastModTime stat fileinfo error")
		return 0
	}
	return fi.ModTime().Unix()
}