package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Atomic int64

func (a *Atomic) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher := w.(http.Flusher)
	closed := w.(http.CloseNotifier)
	s := []byte(r.FormValue("sep"))[0]
	d, err := time.ParseDuration(r.FormValue("interval"))
	if err != nil {
		fmt.Println(err)
		d = time.Second
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Method", "*")
	w.Header().Set("Connection", "keep-alive")
	t := time.NewTicker(d)
	defer t.Stop()
	count := 0
	for count < 10 {
		select {
		case <-closed.CloseNotify():
			fmt.Println("client stoped")
			return
		case <-t.C:
			count++

			pld, _ := json.Marshal(struct {
				Time int64  `json:"time"`
				Code string `json:"code"`
			}{
				Time: time.Now().Unix(),
				Code: fmt.Sprintf("%d", count),
			})
			w.Write(append(pld, s, '\r', '\n'))
			flusher.Flush()
		}
	}
	fmt.Printf("--> %s\n", r.RemoteAddr)
}

func main() {
	atomic := Atomic(1)
	http.Handle("/watch", &atomic)

	if err := http.ListenAndServe(":8088", nil); err != nil {
		fmt.Println(err)
		return
	}
}
