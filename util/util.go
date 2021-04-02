package util

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"runtime/debug"
)

func BtRecover(fnname string) {
	if err := recover(); err != nil {
		panicDetail := fmt.Sprintf("%v\n\n", err)
		traceLog := fmt.Sprintf("%v%v", panicDetail, string(debug.Stack()))
		log.Error().Str("stacktrace from panic", traceLog).Msgf("%s error", fnname)
	}
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