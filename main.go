package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/te6lim/chat-web-socket/trace"
	"github.com/te6lim/chat-web-socket/auth"
)

type templateHandler struct {
	once     sync.Once
	fileName string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.fileName)))
	})
	t.templ.Execute(w, nil)
}

func main() {
	var addr = flag.String("addr", ":8080", "The addr of the application.")
	flag.Parse()
	r :=  newRoom()
	r.tracer = trace.New(os.Stdout)
	http.Handle("/chat", auth.MustAuth(&templateHandler{fileName: "chat.html"}))
	http.Handle("/login", &templateHandler{fileName: "login.html"})
	http.HandleFunc("/auth/", auth.LoginHandler)
	http.Handle("/room", r)
	go r.run()
	log.Println("Starting web server on ", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("Listen and serve: ", err)
	}
}
