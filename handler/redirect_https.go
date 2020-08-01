package handler

import (
	"log"
	"net/http"
)

func redirectHttps(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
}

func ServeRedirectHttps(listen string) {
	go func() {
		err := http.ListenAndServe(listen, http.HandlerFunc(redirectHttps))
		if err != nil {
			log.Fatalln(err)
		}
	}()
}
