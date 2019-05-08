# 构建 web


	package main

	import (
		"database/sql"

		_ "github.com/go-sql-driver/mysql"

		"fmt"
		"log"
		"net/http"
		"os"

		"github.com/gorilla/mux"
	)

	var database *sql.DB

	const (
		PORT    = ":9000"
		DBHost  = "localhost"
		DBPort  = ":3306"
		DBUser  = "wnn"
		DBPass  = "wnnwnn"
		DBDbase = "test"
	)

	type Page struct {
		Title   string
		Content string
		Date    string
	}

	func pageHandler(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r) //解析查询参数为 map
		fmt.Println("%v", vars)
		pageID := vars["id"]
		fileName := "files/" + pageID + ".html"
		_, err := os.Stat(fileName)
		if err != nil {
			fileName = "files/404.html"
		}
		http.ServeFile(w, r, fileName)
	}

	func ServePage(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		pageID := vars["id"]
		thisPage := Page{}
		fmt.Println(pageID)
		err := database.QueryRow("select page_title,page_content,page_date from pages where id=?", pageID).Scan(&thisPage.Title, &thisPage.Content, &thisPage.Date)
		if err != nil {
			log.Println("Couldn't get page:" + pageID)
			log.Println(err.Error)
		}
		html := `<html><head><title>` + thisPage.Title +
			`</title></head><body><h1>` + thisPage.Title + `</h1><div>` +
			thisPage.Content + `</div></body></html>`
		fmt.Fprintln(w, html)
	}

	func ServePageByGUID(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		pageGUID := vars["guid"]
		thisPage := Page{}
		fmt.Println(pageGUID)
		err := database.QueryRow("select page_title,page_content,page_date from pages where page_guid=?", pageGUID).Scan(&thisPage.Title, &thisPage.Content, &thisPage.Date)
		if err != nil {
			log.Println("Couldn't get page:" + pageGUID)
			http.Error(w, http.StatusText(404), http.StatusNotFound)
			log.Println(err.Error)
			return // 这里应该需要 return，书里没有
		}
		html := `<html><head><title>` + thisPage.Title +
			`</title></head><body><h1>` + thisPage.Title + `</h1><div>` +
			thisPage.Content + `</div></body></html>`
		fmt.Fprintln(w, html)
	}

	func main() {
		dbConn := fmt.Sprintf("%s:%s@/%s", DBUser, DBPass, DBDbase)
		fmt.Println(dbConn)
		db, err := sql.Open("mysql", dbConn)
		if err != nil {
			log.Println("Couldn't connect to: " + DBDbase)
			log.Println(err.Error)
		}
		database = db

		rtr := mux.NewRouter()
		// rtr.HandleFunc("/pages/{id:[0-9]+}", ServePage)
		rtr.HandleFunc("/pages/{guid:[a-z0-9A\\-]+}", ServePageByGUID)
		http.Handle("/", rtr)
		http.ListenAndServe(PORT, nil)
	}


# 并发：
注意引用传递在 defer 中的坑。下边的例子输出0而不是100

```go
package main

import "fmt"

func main() {
	a := new(int)

	defer fmt.Println(*a)

	for i := 0; i < 100; i++ {
		*a++
	}
}
```
