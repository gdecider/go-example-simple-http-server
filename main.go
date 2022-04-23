package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"regexp"

	_ "github.com/go-sql-driver/mysql"
)

type Article struct {
	Id                   uint
	Name, Anons, Content string
}

type TmplParams interface{}

func main() {
	handleFunc()
}

func getDbConnection() *sql.DB {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/gosite")
	if err != nil {
		panic(err)
	}

	return db
}

func handleFunc() {
	address := "localhost"
	port := "8080"

	fmt.Printf("Start server at port %s:%s\n", address, port)

	fileServer := http.FileServer(http.Dir("."))
	http.Handle("/static/", fileServer)

	http.HandleFunc("/", pageIndex)
	http.HandleFunc("/create", pageCreate)
	http.HandleFunc("/save_article", pageSaveCreate)
	http.HandleFunc("/article/", pageArticle)

	http.ListenAndServe(":"+port, nil)
}

func getTmplPahts(tmplCode string) []string {
	tmplPath := "templates/" + tmplCode + ".html"
	return []string{"templates/header.html", "templates/footer.html", tmplPath}
}

func execTmpl(out http.ResponseWriter, tmplCode string, tmplParams map[string]TmplParams) {
	tmplPaths := getTmplPahts(tmplCode)

	tmpl, err := template.ParseFiles(tmplPaths...)
	if err != nil {
		fmt.Fprintf(out, err.Error())
	}

	tmpl.ExecuteTemplate(out, tmplCode, tmplParams)
}

func pageIndex(w http.ResponseWriter, r *http.Request) {
	db := getDbConnection()
	defer db.Close()

	rows, err := db.Query("select * from `articles` order by name")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var articles = []Article{}

	for rows.Next() {
		var article Article

		err = rows.Scan(&article.Id, &article.Name, &article.Anons, &article.Content)
		if err != nil {
			panic(err)
		}

		articles = append(articles, article)
	}

	params := map[string]TmplParams{"articles": articles}
	execTmpl(w, "index", params)
}

func pageArticle(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	re, _ := regexp.Compile("[0-9]+")
	id := re.FindString(url)

	if len(id) < 1 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	db := getDbConnection()
	defer db.Close()

	row, err := db.Query(fmt.Sprintf("select * from `articles` where id = '%s'", id))
	if err != nil {
		panic(err)
	}

	defer row.Close()

	article := Article{}

	for row.Next() {
		err = row.Scan(&article.Id, &article.Name, &article.Anons, &article.Content)
		if err != nil {
			panic(err)
		}
	}

	execTmpl(w, "article", map[string]TmplParams{"article": article})
}

func pageCreate(w http.ResponseWriter, r *http.Request) {
	execTmpl(w, "create", nil)
}

func pageSaveCreate(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("name")
	anons := r.FormValue("anons")
	content := r.FormValue("text")

	db := getDbConnection()
	defer db.Close()

	insert, err := db.Query(
		fmt.Sprintf("insert into `articles` (`name`, `anons`, `content`) values ('%s','%s','%s')",
			title, anons, content))
	if err != nil {
		panic(err)
	}

	defer insert.Close()

	http.Redirect(w, r, "/create", http.StatusSeeOther)
}
