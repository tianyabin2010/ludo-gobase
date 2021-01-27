package util

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"runtime/debug"
)

func BtRecover(fnname string) {
	if err := recover(); err != nil {
		panicDetail := fmt.Sprintf("%v\n\n", err)
		traceLog := fmt.Sprintf("%v%v", panicDetail, string(debug.Stack()))
		log.Error().Str("stacktrace from panic", traceLog).Msgf("%s error", fnname)
	}
}
