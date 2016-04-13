package main

import (
    "database/sql"
    "flag"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"

    "github.com/codegangsta/negroni"
    "github.com/gorilla/mux"
    "github.com/gorilla/pat"
    "github.com/gorilla/sessions"
    "github.com/markbates/goth"
    "github.com/markbates/goth/gothic"
    "github.com/markbates/goth/providers/facebook"
    "github.com/markbates/goth/providers/instagram"
    "github.com/markbates/goth/providers/twitter"
    "github.com/lib/pq"
)

var db *sql.DB
var ss = sessions.NewCookieStore([]byte("SHuADRV4npfjU4stuN5dvcYaMmblSZlUyZbEl/mKyyw="))

// Command-line flags
var dbhost = flag.String("dbhost", "localhost", "database host")
var dbport = flag.String("dbport", "5432", "database port")
var address = flag.String("address", "http://localhost:8080", "server address")
var port = flag.String("port", "8080", "server port")

func main() {
    var err error

    // Handle OS signals
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM, os.Kill)
    go func() {
        sig := <-c
        if db != nil {
            db.Close()
        }
        log.Println("Received signal:", sig)
        os.Exit(0);
    }()

    // Parse command-line flags
    flag.Parse()

    // Connect to database
    if db, err = sql.Open("postgres", "user=postgres dbname=postgres sslmode=disable host=" + *dbhost + " port=" + *dbport); err != nil {
        log.Fatal(err)
    }

    if err = db.Ping(); err != nil {
        log.Fatal(err)
    }

    // Prepare database
    if _, err := db.Exec(CREATE_USER_TABLE_SQL); err != nil {
        if err := err.(*pq.Error); err.Code != "42P07" {
            log.Fatal(err)
        }
    }
    if _, err := db.Exec(CREATE_STORY_TABLE_SQL); err != nil {
        if err := err.(*pq.Error); err.Code != "42P07" {
            log.Fatal(err)
        }
    }
    if _, err := db.Exec(CREATE_HOOP_TABLE_SQL); err != nil {
        if err := err.(*pq.Error); err.Code != "42P07" {
            log.Fatal(err)
        }
    }
    if _, err := db.Exec(CREATE_ACTIVITY_TABLE_SQL); err != nil {
        if err := err.(*pq.Error); err.Code != "42P07" {
            log.Fatal(err)
        }
    }
    if _, err := db.Exec(CREATE_HOOP_FEATURED_STORY_TABLE_SQL); err != nil {
        if err := err.(*pq.Error); err.Code != "42P07" {
            log.Fatal(err)
        }
    }
    if _, err := db.Exec(CREATE_HOOP_STORY_TABLE_SQL); err != nil {
        if err := err.(*pq.Error); err.Code != "42P07" {
            log.Fatal(err)
        }
    }

    // Setup social logins
    gothic.Store = sessions.NewFilesystemStore(os.TempDir(), []byte("pinoy-hoops"))
    goth.UseProviders(
        facebook.New(os.Getenv("FACEBOOK_KEY"), os.Getenv("FACEBOOK_SECRET"), *address + "/auth/facebook/callback"),
        instagram.New(os.Getenv("INSTAGRAM_KEY"), os.Getenv("INSTAGRAM_SECRET"), *address + "/auth/instagram/callback"),
        twitter.New(os.Getenv("TWITTER_KEY"), os.Getenv("TWITTER_SECRET"), *address + "/auth/twitter/callback"),
    )

    // Prepare web server
    router := mux.NewRouter()
    apiRouter := router.PathPrefix("/api").Subrouter()
    apiRouter.HandleFunc("/login", loginHandler)
    apiRouter.HandleFunc("/logout", logoutHandler)
    apiRouter.HandleFunc("/hoop", hoopHandler)
    apiRouter.HandleFunc("/hoops", hoopsHandler)
    apiRouter.HandleFunc("/activities", activitiesHandler)

    // Prepare social login authenticators
    patHandler := pat.New()
    patHandler.Get("/auth/{provider}/callback", authHandler)
    patHandler.Get("/auth/{provider}", gothic.BeginAuthHandler)
    router.PathPrefix("/auth").Handler(patHandler)

    // Run web server
    n := negroni.Classic()
    n.UseHandler(router)
    n.Run(":" + *port)
}

func authHandler(w http.ResponseWriter, r *http.Request) {
    authuser, err := gothic.CompleteUserAuth(w, r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    user := &User{}

    switch authuser.Provider {
    case "facebook":
        user.FacebookID.String = authuser.UserID
        user.FacebookID.Valid = true
    case "instagram":
        user.InstagramID.String = authuser.UserID
        user.InstagramID.Valid = true
    case "twitter":
        user.TwitterID.String = authuser.UserID
        user.TwitterID.Valid = true
    default:
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    user.Name.String = authuser.Name
    user.Description.String = authuser.Description
    user.Email.String = authuser.Email
    user.ImageURL.String = authuser.AvatarURL
    user.Name.Valid = true
    user.Description.Valid = true
    user.ImageURL.Valid = true
    user.Email.Valid = true

    if err := insertUser(user); err != nil {
        log.Println(err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    if err := logIn(w, r, user); err != nil {
        log.Println(err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        ok := loggedIn(w, r)
        if ok {
            w.WriteHeader(http.StatusOK)
        } else {
            w.WriteHeader(http.StatusForbidden)
        }
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
    if err := logOut(w, r); err != nil {
        log.Println(err)
        w.WriteHeader(http.StatusInternalServerError)
    }
    w.WriteHeader(http.StatusOK)
}

func hoopHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

func hoopsHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

func activitiesHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}
