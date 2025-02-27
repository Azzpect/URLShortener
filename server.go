package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"sync"

	"github.com/gorilla/mux"
)

func handler(w http.ResponseWriter, r *http.Request) {
  http.ServeFile(w, r, "./index.html")
}

var redirectedURL map[string]string = map[string]string{}


func generateId() string {
  letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

  id := make([]byte, 8)

  for i := range 8 {
    id[i] = letters[rand.Intn(len(letters))] 
  }

  return string(id)
}

func shortenerHandler(w http.ResponseWriter, r *http.Request) {

  type Data struct {
    Url string `json:"url"`
  }
  var data Data

  err := json.NewDecoder(r.Body).Decode(&data)

  if err != nil {
    http.Error(w, "Invalid json", http.StatusBadRequest)
    return
  }

  pattern := `^https?:\/\/\S+$`

  match, _ := regexp.Match(pattern, []byte(data.Url))

  if !match {
    data.Url = fmt.Sprintf("https://%s", data.Url)
  }

  exists, id := redirectionExists(data.Url)

  if !exists {
    id = generateId()
    redirectedURL[id] = data.Url
  }

  data.Url = fmt.Sprintf("http://localhost:8080/redirect/%s", id)

  redirectUrl, err := json.Marshal(data)

  if err != nil {
    http.Error(w, "Invalid json", http.StatusInternalServerError)
    return
  }

  fmt.Fprint(w, string(redirectUrl))
}

func redirectionExists(url string) (bool, string) {
  for key, value := range redirectedURL {
    if value == url {
      return true, key
    }
  }
  return false, ""
}


func handleRedirect(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  id := vars["id"]

  url, exists := redirectedURL[id]
  if !exists {
    http.Error(w, "Invalid json", http.StatusNotFound)
    return
  }
  http.Redirect(w, r, url, http.StatusSeeOther)
}

func runServer(wg *sync.WaitGroup) {
  defer wg.Done()

  router := mux.NewRouter()

  router.HandleFunc("/", handler).Methods("get")
  router.HandleFunc("/submit", shortenerHandler).Methods("post")
  router.HandleFunc("/redirect/{id}", handleRedirect).Methods("get")
  router.HandleFunc("/copy.svg", func(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "./copy-svgrepo-com.svg")
  }).Methods("get")


  http.ListenAndServe(":8080", router)

  
  fmt.Println("Server started at port 8080")

}

func main() {
  var wg sync.WaitGroup

  wg.Add(1)

  go runServer(&wg)


  fmt.Println("Server started at port 8080")
  wg.Wait()
}
