package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type App struct {
	Router *mux.Router
}

func (a *App) Initialize() {
	path := getContents()
	pathJoined := fmt.Sprintf("/%s", path)
	router := mux.NewRouter()
	router.Handle(pathJoined, ExecuteFunction(path)).Methods("GET")
	fmt.Printf("Available routes:\n")
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		t, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		m, err := route.GetMethods()
		if err != nil {
			return err
		}
		fmt.Printf("%s %s\n", m, t)
		return nil
	})
	a.Router = router
}

func (a *App) Run(addr string) {
	handler := cors.Default().Handler(a.Router)
	log.Printf("Server is listening on %v\n", addr)
	http.ListenAndServe(addr, handler)
}

func ExecuteFunction(path string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		pythonize := fmt.Sprintf("%s.py", path)
		g := execute(pythonize)
		JSONResponse(w, 200, map[string]string{"Genesis": g})
	})
}

func JSONResponse(w http.ResponseWriter, code int, output interface{}) {
	response, _ := json.Marshal(output)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func execute(pathToFile string) string {
	f, err := ioutil.ReadFile(pathToFile)
	if err != nil {
		log.Fatal("Unable to read file: ", err)
		return ""
	}
	code := string(f)

	cmd := exec.Command("/usr/local/bin/python3", "-c", code)

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	return string(out)
}

func getContents() string {
	var files []string

	root := "/Users/arturkondas/Desktop/work/vio-server"
	err := filepath.Walk(root, visit(&files))
	if err != nil {
		log.Fatal(err)
	}

	pyIndex := Find(files, "py")
	s := files[pyIndex]
	slash := strings.LastIndex(s, "/")
	pathTrimmed := s[slash+1:]
	dot := strings.LastIndex(pathTrimmed, ".")
	dotTrimmed := pathTrimmed[:dot]

	return dotTrimmed
}

func visit(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		*files = append(*files, path)
		return nil
	}
}

func Find(a []string, x string) int {
	for i, n := range a {
		if strings.Contains(n, x) {
			return i
		}
	}
	return len(a)
}
