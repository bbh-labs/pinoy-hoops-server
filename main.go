package main

import (
    "database/sql"
    "encoding/json"
    "flag"
    "log"
    "net/http"
    "path"
    "os"
    "os/signal"
    "strconv"
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
    apiRouter.HandleFunc("/story", storyHandler)
    apiRouter.HandleFunc("/stories", storiesHandler)
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
        user.FacebookID = authuser.UserID
    case "instagram":
        user.InstagramID = authuser.UserID
    case "twitter":
        user.TwitterID = authuser.UserID
    default:
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    user.Name = authuser.Name
    user.Description = authuser.Description
    user.Email = authuser.Email
    user.ImageURL = authuser.AvatarURL

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
        ok, _ := loggedIn(w, r, false)
        if ok {
            w.Write([]byte("Logged In"))
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
    case "POST":
        ok, user := loggedIn(w, r, false)
        if !ok {
            w.WriteHeader(http.StatusForbidden)
            return
        }

        imageURL := r.FormValue("image_url")
        if imageURL == "" {
            if _, fileheader, err := r.FormFile("image"); err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            } else {
                if err := os.MkdirAll("content", os.ModeDir | 0775); err != nil {
                    log.Println(err)
                    w.WriteHeader(http.StatusInternalServerError)
                    return
                }

                basename := path.Base(fileheader.Filename)
                destname := "content/" + basename
                if err := os.Rename(fileheader.Filename, destname); err != nil {
                    log.Println(err)
                    w.WriteHeader(http.StatusInternalServerError)
                    return
                }

                imageURL = destname
            }
        }

        latitude, err := strconv.ParseFloat(r.FormValue("latitude"), 64)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }

        longitude, err := strconv.ParseFloat(r.FormValue("longitude"), 64)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }

        name := r.FormValue("name")
        description := r.FormValue("description")

        // Start Transaction
        tx, err := db.Begin()
        if err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }

        var hoopID, storyID int64

        // Insert Hoop
        if result, err := tx.Exec(INSERT_HOOP_SQL, user.ID, name, description, latitude, longitude); err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        } else if id, err := result.LastInsertId(); err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        } else {
            hoopID = id
        }

        // Insert Story
        if result, err := tx.Exec(INSERT_STORY_SQL, user.ID, name, description, imageURL); err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        } else if id, err := result.LastInsertId(); err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        } else {
            storyID = id
        }

        // Insert HoopFeaturedStory
        if _, err := tx.Exec(INSERT_HOOP_FEATURED_STORY_SQL, hoopID, storyID); err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }

        // Insert Activity
        if _, err := tx.Exec(INSERT_POST_HOOP_ACTIVITY_SQL, user.ID, ACTIVITY_POST_HOOP, hoopID); err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }

        // End Transaction
        if err := tx.Commit(); err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

func hoopsHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        var hoops []Hoop

        rows, err := db.Query(GET_HOOPS_SQL)
        if err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }

        for rows.Next() {
            var hoop Hoop

            if err := rows.Scan(
                &hoop.ID,
                &hoop.UserID,
                &hoop.Name,
                &hoop.Description,
                &hoop.Latitude,
                &hoop.Longitude,
                &hoop.CreatedAt,
                &hoop.UpdatedAt,
            ); err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            }

            hoops = append(hoops, hoop)
        }

        data, err := json.Marshal(hoops)
        if err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }

        w.Write(data)
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

func storyHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "POST":
        ok, user := loggedIn(w, r, false)
        if !ok {
            w.WriteHeader(http.StatusForbidden)
            return
        }

        hoopID, err := strconv.ParseInt(r.FormValue("hoop_id"), 10, 64)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }

        imageURL := r.FormValue("image_url")
        if imageURL == "" {
            if _, fileheader, err := r.FormFile("image"); err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            } else {
                if err := os.MkdirAll("content", os.ModeDir | 0775); err != nil {
                    log.Println(err)
                    w.WriteHeader(http.StatusInternalServerError)
                    return
                }

                basename := path.Base(fileheader.Filename)
                destname := "content/" + basename
                if err := os.Rename(fileheader.Filename, destname); err != nil {
                    log.Println(err)
                    w.WriteHeader(http.StatusInternalServerError)
                    return
                }

                imageURL = destname
            }
        }

        name := r.FormValue("name")
        description := r.FormValue("description")

        tx, err := db.Begin()
        if err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }

        // Insert Story
        result, err := tx.Exec(INSERT_STORY_SQL, hoopID, user.ID, name, description, imageURL)
        if err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }

        storyID, err := result.LastInsertId()
        if err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }

        // Insert Activity
        if _, err := tx.Exec(INSERT_POST_STORY_ACTIVITY_SQL, user.ID, ACTIVITY_POST_STORY, storyID); err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }

        if err := tx.Commit(); err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

func storiesHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        hoopID, err := strconv.ParseInt(r.FormValue("hoop_id"), 10, 64)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }

        var stories []Story

        rows, err := db.Query(GET_STORIES_SQL, hoopID)
        if err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }

        for rows.Next() {
            var story Story

            if err := rows.Scan(
                &story.ID,
                &story.UserID,
                &story.Name,
                &story.Description,
                &story.ImageURL,
                &story.CreatedAt,
                &story.UpdatedAt,
            ); err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            }

            stories = append(stories, story)
        }

        data, err := json.Marshal(stories)
        if err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }

        w.Write(data)
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

func activitiesHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        var activities []Activity

        rows, err := db.Query(GET_ACTIVITIES_SQL)
        if err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }

        var hoopID, storyID sql.NullInt64

        for rows.Next() {
            var activity Activity

            if err := rows.Scan(
                &activity.UserID,
                &activity.Type,
                &hoopID,
                &storyID,
                &activity.CreatedAt,
            ); err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            }

            activity.HoopID = fromNullInt64(hoopID)
            activity.StoryID = fromNullInt64(storyID)

            activities = append(activities, activity)
        }

        data, err := json.Marshal(activities)
        if err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }

        w.Write(data)
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}
