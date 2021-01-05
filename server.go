package main

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	store *sessions.CookieStore = sessions.NewCookieStore([]byte("12345678901234567890"))
)

const (
	SessionName       = "session-name"
	ContextSessionKey = "session"
)

type Task struct {
	Id        int
	UserId    string
	Title     string
	Details   string
	CreatedAt time.Time
}

type User struct {
	Id       int    `json:"id"`
	UserId   string `json:"user_id"`
	Password string `json:"password"`
}

func ChangeMethodsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rg := regexp.MustCompile(`/delete`)

		log.Printf("Method: %s, URI: %s\n", r.Method, r.RequestURI)
		if (r.Method == "POST") && (r.RequestURI == "/update") {
			log.Printf("Method is changed to PUT")
			r.Method = "PUT"
		} else if (r.Method == "POST") && (rg.MatchString(r.RequestURI)) {
			log.Printf("Method is changed to DELETE")
			r.Method = "DELETE"
		}
		next.ServeHTTP(w, r)
	})
}

func sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, SessionName)
		if err != nil {
			panic(err.Error())
		}
		context.Set(r, ContextSessionKey, session)
		next.ServeHTTP(w, r)
	})

}

var Db *sql.DB

func init() {
	var err error
	Db, err = sql.Open("mysql", "dummy:@tcp(127.0.0.1:3306)/todo?parseTime=true")
	if err != nil {
		panic(err.Error())
	}

	file, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", err)
	}
	logger = log.New(file, "INFO ", log.Ldate|log.Ltime|log.Lshortfile)
}

func CreateUser(user_id string, password string) {
	stmt, err := Db.Prepare("INSERT INTO users (user_id, password) VALUES (?, ?);")
	if err != nil {
		panic(err.Error())
	}
	stmt.Exec(user_id, password)
}

func GetUser(user_id string, password string) (user User, err error) {
	user = User{}
	err = Db.QueryRow("SELECT * FROM users WHERE user_id = ? AND password = ?;", user_id, password).Scan(&user.Id, &user.UserId, &user.Password)
	return
}

func GetAllTasks(user_id string) (tasks []Task, err error) {
	rows, err := Db.Query("SELECT * FROM tasks WHERE user_id = ?;", user_id)
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		task := Task{}
		err := rows.Scan(&task.Id, &task.UserId, &task.Title, &task.Details, &task.CreatedAt)
		if err != nil {
			panic(err.Error())
		}
		tasks = append(tasks, task)
	}
	return
}

func GetTask(id string) (task Task, err error) {
	task = Task{}
	err = Db.QueryRow("SELECT * FROM tasks WHERE id = ?;", id).Scan(&task.Id, &task.UserId, &task.Title, &task.Details, &task.CreatedAt)
	return
}

func CreatTask(user_id string, title string, details string) {
	stmt, err := Db.Prepare("INSERT INTO tasks (user_id, title, details) VALUES (?, ?, ?);")
	if err != nil {
		panic(err.Error())
	}
	stmt.Exec(user_id, title, details)
}

func DeleteTask(id string) {
	stmt, err := Db.Prepare("DELETE FROM tasks WHERE id = ?;")
	if err != nil {
		panic(err.Error())
	}
	stmt.Exec(id)
}

func UpdateTask(id string, title string, details string) {
	stmt, err := Db.Prepare("UPDATE tasks SET title = ?, details = ? WHERE id = ?;")
	if err != nil {
		panic(err.Error())
	}
	stmt.Exec(title, details, id)
}

func getTitle(task Task) string {
	return task.Title
}

func getId(task Task) int {
	return task.Id
}

func getDetails(task Task) string {
	return task.Details
}

func getSession(r *http.Request) (*sessions.Session, error) {
	info("Get session")
	if v := context.Get(r, ContextSessionKey); v != nil {
		return v.(*sessions.Session), nil
	}
	return nil, errors.New("failed to get session")
}

func getUserIdFromSession(r *http.Request) (string, error) {
	session, err := getSession(r)
	if err != nil {
		panic(err.Error())
	}
	if v, ok := session.Values["user_id"]; ok {
		return v.(string), nil
	} else {
		return "", errors.New("user_id not found")
	}
}

func handleSignUpGET(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("signup_tmpl.html")
	t.Execute(w, nil)
}

func handleSignUpPOST(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	user_id := r.PostForm["user_id"][0]
	password := r.PostForm["password"][0]
	if user_id == "" || password == "" {
		http.Redirect(w, r, "/signup", http.StatusFound)
		return
	}
	CreateUser(user_id, password)

	session, err := getSession(r)
	if err != nil {
		panic(err.Error())
	}
	session.Values["user_id"] = user_id
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleLogInGET(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("login.html")
	t.Execute(w, nil)
}

func handleLogInPOST(w http.ResponseWriter, r *http.Request) {
	info("handleLogInPOST method is called")
	r.ParseForm()
	user_id := r.PostForm["user_id"][0]
	password := r.PostForm["password"][0]
	if user_id == "" || password == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	user, err := GetUser(user_id, password)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
	session, err := getSession(r)
	if err != nil {
		panic(err.Error())
	}
	session.Values["user_id"] = user.UserId
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleLogOutGET(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("logout.html")
	t.Execute(w, nil)
}

func handleLogOutPOST(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		panic(err.Error())
	}
	delete(session.Values, "user_id")
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	type page struct {
		UserId string
	}
	p := &page{}
	user_id, err := getUserIdFromSession(r)
	log.Println(user_id)
	if err == nil {
		p.UserId = user_id
	}
	log.Println(p)
	t, _ := template.ParseFiles("index_tmpl.html")
	t.Execute(w, p)
}

func handleGetAllTasks(w http.ResponseWriter, r *http.Request) {
	info("GetAllTasks")
	user_id, _ := getUserIdFromSession(r)
	tasks, err := GetAllTasks(user_id)
	if err != nil {
		panic(err.Error())
	}
	funcMap := template.FuncMap{"getTitle": getTitle, "getId": getId}
	t := template.New("list_tmpl.html").Funcs(funcMap)
	t, _ = t.ParseFiles("list_tmpl.html")
	t.Execute(w, tasks)
}

func handleGetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	info("Get a task for id =", id)
	task, err := GetTask(id)
	if err != nil {
		panic(err.Error())
	}

	funcMap := template.FuncMap{"getTitle": getTitle, "getId": getId, "getDetails": getDetails}
	t := template.New("details_tmpl.html").Funcs(funcMap)
	t, _ = t.ParseFiles("details_tmpl.html")
	t.Execute(w, task)
}

func handleCreateTask(w http.ResponseWriter, r *http.Request) {
	// this uri is static ?
	t, _ := template.ParseFiles("create_tmpl.html")
	t.Execute(w, nil)
}

func handleRegisterTask(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	user_id, err := getUserIdFromSession(r)
	if err != nil {
		panic(err.Error())
	}
	title := r.PostForm["title"][0]
	details := r.PostForm["details"][0]
	CreatTask(user_id, title, details)
	http.Redirect(w, r, "/list", 301)
}

func handleEditTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	info("Edit task for", id)
	task, err := GetTask(id)
	if err != nil {
		panic(err.Error())
	}

	funcMap := template.FuncMap{"getTitle": getTitle, "getId": getId, "getDetails": getDetails}
	t := template.New("edit_tmpl.html").Funcs(funcMap)
	t, _ = t.ParseFiles("edit_tmpl.html")
	t.Execute(w, task)
}

func handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id := r.PostForm["id"][0]
	title := r.PostForm["title"][0]
	details := r.PostForm["details"][0]
	info("Update task of id =", id)
	UpdateTask(id, title, details)
	http.Redirect(w, r, "/list", 301)
}

func handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	info("DELETE method")
	uri := r.RequestURI
	uri_coms := strings.Split(uri, "/")
	id := uri_coms[2]
	info("Delete task of id =", id)
	DeleteTask(id)
	http.Redirect(w, r, "/list", 301)
}

var logger *log.Logger

func info(args ...interface{}) {
	logger.SetPrefix("INFO ")
	logger.Println(args...)
}

func danger(args ...interface{}) {
	logger.SetPrefix("ERROR ")
	logger.Println(args...)
}

func warning(args ...interface{}) {
	logger.SetPrefix("WARNING ")
	logger.Println(args...)
}

func main() {
	r := mux.NewRouter()
	r.Use(ChangeMethodsMiddleware)
	r.Use(sessionMiddleware)
	r.HandleFunc("/", handleRoot).Methods("GET")
	r.HandleFunc("/signup", handleSignUpGET).Methods("GET")
	r.HandleFunc("/signup", handleSignUpPOST).Methods("POST")
	r.HandleFunc("/login", handleLogInGET).Methods("GET")
	r.HandleFunc("/login", handleLogInPOST).Methods("POST")
	r.HandleFunc("/logout", handleLogOutGET).Methods("GET")
	r.HandleFunc("/logout", handleLogOutPOST).Methods("POST")
	r.HandleFunc("/list", handleGetAllTasks).Methods("GET")
	r.HandleFunc("/list/{id:[0-9]+}", handleGetTask).Methods("GET")
	r.HandleFunc("/create", handleCreateTask).Methods("GET")
	r.HandleFunc("/register", handleRegisterTask).Methods("POST")
	r.HandleFunc("/edit/{id:[0-9]+}", handleEditTask).Methods("GET")
	r.HandleFunc("/update", handleUpdateTask)
	r.HandleFunc("/delete/{id:[0-9]+}", handleDeleteTask)

	log.Fatal(http.ListenAndServe(":12345", r))
}
