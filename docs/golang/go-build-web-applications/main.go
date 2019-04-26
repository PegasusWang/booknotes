package main

/*
export GO111MODULE=on
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
INSERT INTO `pages` (`id`, `page_guid`, `page_title`, `page_content`, `page_date`) VALUES (4, 'lorem-ipsum', 'Lorem Ipsum', 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas sem tortor, lobortis in posuere sit amet, ornare non eros. Pellentesque vel lorem sed nisl dapibus fringilla. In pretium...', '2015-05-06 04:09:45');

// create user
CREATE USER 'wnn'@'localhost';
GRANT ALL PRIVILEGES ON test.* To 'wnn'@'localhost' IDENTIFIED BY 'wnnwnn';


CREATE TABLE `comments` (
`id` int(11) unsigned NOT NULL AUTO_INCREMENT,
`page_id` int(11) NOT NULL,
`comment_guid` varchar(256) DEFAULT NULL,
`comment_name` varchar(64) DEFAULT NULL,
`comment_email` varchar(128) DEFAULT NULL,
`comment_text` mediumtext,
`comment_date` timestamp NULL DEFAULT NULL,
PRIMARY KEY (`id`),
KEY `page_id` (`page_id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

*/

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"strconv"

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
	ID         int
	Title      string
	RawContent string
	Content    template.HTML
	Date       string
	Comments   []Comment
	GUID       string
	// Session    Session
}

// JSONResponse return type  struct
type JSONResponse struct {
	Fields map[string]string
}

// Comment def
type Comment struct {
	ID          int
	Name        string
	Email       string
	CommentText string
}

// TruncateText def
func (p Page) TruncateText() template.HTML {
	chars := 0
	for i := range p.Content {
		chars++
		if chars > 150 {
			return p.Content[:i] + ` ...`
		}
	}
	return p.Content
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
	err := database.QueryRow(
		"select id,page_title,page_content,page_date from pages where page_guid=?", pageGUID,
	).Scan(&thisPage.ID, &thisPage.Title, &thisPage.Content, &thisPage.Date)
	if err != nil {
		log.Println("Couldn't get page:" + pageGUID)
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		log.Println(err.Error)
		return // 这里应该需要 return，书里没有
	}

	// query comments
	comments, err := database.Query("SELECT id, comment_name as Name, comment_email, comment_text FROM comments WHERE page_id=?", thisPage.ID)
	if err != nil {
		log.Println(err)
	}
	for comments.Next() {
		var comment Comment
		comments.Scan(&comment.ID, &comment.Name, &comment.Email, &comment.CommentText)
		thisPage.Comments = append(thisPage.Comments, comment)
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
	pages, err := database.Query("select page_title,page_content,page_date,page_guid from pages order by ? desc", "page_date")
	if err != nil {
		fmt.Fprintln(w, err.Error)
	}
	defer pages.Close() // need close
	for pages.Next() {
		thisPage := Page{}
		pages.Scan(&thisPage.Title, &thisPage.RawContent, &thisPage.Date, &thisPage.GUID)
		thisPage.Content = template.HTML(thisPage.RawContent)
		Pages = append(Pages, thisPage)
	}
	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(w, Pages)
}

// APIPage def
func APIPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageGUID := vars["guid"]
	thisPage := Page{}
	fmt.Println(pageGUID)

	err := database.QueryRow(
		"SELECT page_title,page_content,page_date FROM pages WHERE page_guid=?", pageGUID,
	).Scan(&thisPage.Title, &thisPage.RawContent, &thisPage.Date)
	thisPage.Content = template.HTML(thisPage.RawContent)
	if err != nil {
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		log.Println(err)
		return
	}
	APIOutput, err := json.Marshal(thisPage)
	fmt.Println(APIOutput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, thisPage)
}

// APICommentPost create comment
func APICommentPost(w http.ResponseWriter, r *http.Request) {
	var commentAdded bool
	err := r.ParseForm()
	if err != nil {
		log.Println(err.Error)
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	comments := r.FormValue("comments")
	pageID := r.FormValue("page_id")
	res, err := database.Exec(
		"insert into comments set page_id=?,comment_name=?,comment_email=?,comment_text=?", pageID, name, email, comments,
	)
	if err != nil {
		log.Println(err.Error)
	}

	id, err := res.LastInsertId()
	if err != nil {
		commentAdded = false
	} else {
		commentAdded = true
	}

	commentAddedBool := strconv.FormatBool(commentAdded) // to string
	var resp JSONResponse
	resp.Fields = make(map[string]string)
	resp.Fields["id"] = strconv.FormatInt(id, 10)
	resp.Fields["added"] = commentAddedBool
	jsonResp, _ := json.Marshal(resp.Fields)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, jsonResp)
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

	routes := mux.NewRouter()
	// rest api
	routes.HandleFunc("/api/pages", APIPage).Methods("GET").Schemes("https")
	routes.HandleFunc("/api/pages/{guid:[0-9a-zA\\-]+}", APIPage).Methods("GET").Schemes("https")
	// comment rest api
	routes.HandleFunc("/api/comments", APICommentPost).Methods("POST")

	// routes.HandleFunc("/pages/{id:[0-9]+}", ServePage)
	routes.HandleFunc("/pages/{guid:[a-z0-9A\\-]+}", ServePageByGUID)
	routes.HandleFunc("/", RedirIndex)
	routes.HandleFunc("/home", ServeIndex)
	http.Handle("/", routes)
	http.ListenAndServe(PORT, nil)
}
