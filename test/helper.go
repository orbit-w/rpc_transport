package test

import (
	"log"
	"net/http"
	_ "net/http/pprof"
)

func StartPProf() {
	go func() {
		log.Println(http.ListenAndServe("127.0.0.1:6060", nil))
	}()
}
