package main

/*
CREATE TABLE `pages` (
	`id` int(11) unsigned NOT NULL AUTO_INCREMENT,
	`page_guid` varchar(256) NOT NULL DEFAULT '',
	`page_title` varchar(256) DEFAULT NULL,
	`page_content` mediumtext,
	`page_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (`id`),
	UNIQUE KEY `page_guid` (`page_guid`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=latin1;

INSERT INTO `pages` (`id`, `page_guid`, `page_title`, `page_content`, `page_date`) VALUES (NULL, 'hello-world', 'Hello, World', 'I\'m so glad you found this page!  It\'s been sitting patiently on the Internet for some time, just waiting for a visitor.', CURRENT_TIMESTAMP);
INSERT INTO `pages` (`id`, `page_guid`, `page_title`, `page_content`, `page_date`) VALUES (3, 'a-new-blog', 'A New Blog', 'I hope you enjoyed the last blog!  Well brace yourself, because my latest blog is even <i>better</i> than the last!', '2015-04-29 02:16:19');

// create user
CREATE USER 'wnn'@'localhost';
GRANT ALL PRIVILEGES ON test.* To 'wnn'@'localhost' IDENTIFIED BY 'wnnwnn';

*/

import (
	"database/sql"
	"html/template"

	_ "github.com/go-sql-driver/mysql"

	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

var database *sql.DB

// PORT
const (
	PORT    = ":9000"
	DBHost  = "localhost"
	DBPort  = ":3306"
	DBUser  = "wnn"
	DBPass  = "wnnwnn"
	DBDbase = "test"
)

// Page def
type Page struct {
	Title      string
	RawContent string
	Content    template.HTML
	Date       string
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

// ServePage def
func ServePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageID := vars["id"]
	thisPage := Page{}
	fmt.Println(pageID)
	err := database.QueryRow(
		"select page_title,page_content,page_date from pages where id=?", pageID,
	).Scan(&thisPage.Title, &thisPage.RawContent, &thisPage.Date) // note use pointer
	thisPage.Content = template.HTML(thisPage.RawContent)
	if err != nil {
		log.Println("Couldn't get page:" + pageID)
		log.Println(err.Error)
	}
	t, _ := template.ParseFiles("templates/blog.html")
	t.Execute(w, thisPage)
}

// ServePageByGUID def
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
	t, _ := template.ParseFiles("templates/blog.html")
	t.Execute(w, thisPage)
}

// RedirIndex def
func RedirIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/home", 301)
}

// ServeIndex def
func ServeIndex(w http.ResponseWriter, r *http.Request) {
	var Pages = []Page{}
	pages, err := database.Query("select page_title,page_content,page_date from pages order by ? desc", "page_date")
	if err != nil {
		fmt.Fprintln(w, err.Error)
	}
	defer pages.Close()
	for pages.Next() {
		thisPage := Page{}
		pages.Scan(&thisPage.Title, &thisPage.RawContent, &thisPage.Date)
		thisPage.Content = template.HTML(thisPage.RawContent)
		Pages = append(Pages, thisPage)
	}
	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(w, Pages)
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
	rtr.HandleFunc("/pages/{id:[0-9]+}", ServePage)
	// rtr.HandleFunc("/pages/{guid:[a-z0-9A\\-]+}", ServePageByGUID)
	// rtr.HandleFunc("/", RedirIndex)
	// rtr.HandleFunc("/home", ServeIndex)
	http.Handle("/", rtr)
	http.ListenAndServe(PORT, nil)
}
