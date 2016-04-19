package main

import (
    "database/sql"
    "encoding/json"
    "errors"
    "flag"
    "io"
    "log"
    "mime/multipart"
    "net/http"
    "os"
    "os/exec"
    "os/signal"
    "strconv"
    "strings"
    "syscall"

    "github.com/codegangsta/negroni"
    "github.com/gorilla/mux"
    "github.com/gorilla/pat"
    "github.com/gorilla/sessions"
    "github.com/lib/pq"
    "github.com/markbates/goth"
    "github.com/markbates/goth/gothic"
    "github.com/markbates/goth/providers/facebook"
    "github.com/markbates/goth/providers/instagram"
    "github.com/markbates/goth/providers/twitter"
    "golang.org/x/crypto/bcrypt"
)

var db *sql.DB
var ss = sessions.NewCookieStore([]byte("SHuADRV4npfjU4stuN5dvcYaMmblSZlUyZbEl/mKyyw="))

// Command-line flags
var dbhost = flag.String("dbhost", "localhost", "database host")
var dbport = flag.String("dbport", "5432", "database port")
var address = flag.String("address", "http://localhost:8080", "server address")
var port = flag.String("port", "8080", "server port")

// Errors
var (
    ErrEmailTooShort    = errors.New("Email too short")
    ErrPasswordTooShort = errors.New("Password too short")
    ErrNotLoggedIn      = errors.New("User is not logged in")
    ErrPasswordMismatch = errors.New("Password mismatch")
)

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
        os.Exit(0)
    }()

    // Parse command-line flags
    flag.Parse()

    // Connect to database
    if db, err = sql.Open("postgres", "user=postgres dbname=postgres sslmode=disable host="+*dbhost+" port="+*dbport); err != nil {
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
    if _, err := db.Exec(CREATE_HOOP_TABLE_SQL); err != nil {
        if err := err.(*pq.Error); err.Code != "42P07" {
            log.Fatal(err)
        }
    }
    if _, err := db.Exec(CREATE_STORY_TABLE_SQL); err != nil {
        if err := err.(*pq.Error); err.Code != "42P07" {
            log.Fatal(err)
        }
    }
    if _, err := db.Exec(CREATE_COMMENT_TABLE_SQL); err != nil {
        if err := err.(*pq.Error); err.Code != "42P07" {
            log.Fatal(err)
        }
    }
    if _, err := db.Exec(CREATE_LIKE_TABLE_SQL); err != nil {
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
        facebook.New(os.Getenv("FACEBOOK_KEY"), os.Getenv("FACEBOOK_SECRET"), *address+"/auth/facebook/callback"),
        instagram.New(os.Getenv("INSTAGRAM_KEY"), os.Getenv("INSTAGRAM_SECRET"), *address+"/auth/instagram/callback"),
        twitter.New(os.Getenv("TWITTER_KEY"), os.Getenv("TWITTER_SECRET"), *address+"/auth/twitter/callback"),
    )

    // Prepare web server
    router := mux.NewRouter()
    apiRouter := router.PathPrefix("/api").Subrouter()
    apiRouter.HandleFunc("/login", loginHandler)
    apiRouter.HandleFunc("/signup", signupHandler)
    apiRouter.HandleFunc("/logout", logoutHandler)
    apiRouter.HandleFunc("/user", userHandler)
    apiRouter.HandleFunc("/hoop", hoopHandler)
    apiRouter.HandleFunc("/hoops", hoopsHandler)
    apiRouter.HandleFunc("/story", storyHandler)
    apiRouter.HandleFunc("/stories", storiesHandler)
    apiRouter.HandleFunc("/activities", activitiesHandler)
    apiRouter.HandleFunc("/comment", commentHandler)
    apiRouter.HandleFunc("/comments", commentsHandler)
    apiRouter.HandleFunc("/like", likeHandler)
    apiRouter.HandleFunc("/likes", likesHandler)

    // Prepare extra handlers
    apiRouter.HandleFunc("/hoops/nearby", nearbyHoopsHandler)
    apiRouter.HandleFunc("/hoops/popular", popularHoopsHandler)
    apiRouter.HandleFunc("/hoops/latest", latestHoopsHandler)
    apiRouter.HandleFunc("/stories/mostliked", mostLikedStoriesHandler)
    apiRouter.HandleFunc("/stories/mostviewed", mostViewedStoriesHandler)
    apiRouter.HandleFunc("/stories/latest", latestStoriesHandler)
    apiRouter.HandleFunc("/likes/count", likesCountHandler)

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
        log.Println(err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if loggedIn, user := loggedIn(w, r, true); loggedIn {
        switch authuser.Provider {
        case "facebook":
            if _, err := db.Exec(UPDATE_USER_FACEBOOK_SQL, authuser.UserID, user.ID); err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            }
        case "instagram":
            if _, err := db.Exec(UPDATE_USER_INSTAGRAM_SQL, authuser.UserID, user.ID); err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            }
        case "twitter":
            if _, err := db.Exec(UPDATE_USER_TWITTER_SQL, authuser.UserID, user.ID); err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            }
        default:
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
        return
    }

    user := &User{}

    switch authuser.Provider {
    case "facebook":
        user.FacebookID = authuser.UserID
    case "instagram":
        user.InstagramID = authuser.NickName
    case "twitter":
        user.TwitterID = authuser.UserID
    default:
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    if exists, user := userExists(user, true); exists {
        if err := logIn(w, r, user); err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
        }
        return
    }

    name := strings.Split(authuser.Name, " ")

    if len(name) > 1 {
        user.Firstname = strings.Join(name[:len(name)-1], " ")
        user.Lastname = name[len(name)-1]
    } else {
        user.Firstname = name[0]
    }
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

    http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        if ok, user := loggedIn(w, r, true); !ok {
            w.WriteHeader(http.StatusForbidden)
        } else {
            if data, err := json.Marshal(user); err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
            } else {
                w.Write(data)
            }
        }
    case "POST":
        email := r.FormValue("email")
        if len(email) < 6 {
            w.WriteHeader(http.StatusBadRequest)
            w.Write([]byte("Email is too short"))
            return
        }

        password := r.FormValue("password")
        if len(password) < 8 {
            w.WriteHeader(http.StatusBadRequest)
            w.Write([]byte("Password is too short"))
            return
        }

        user := &User{Email: email}
        if exists, user := userExists(user, true); exists {
            if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
                w.WriteHeader(http.StatusForbidden)
            } else {
                if err := logIn(w, r, user); err != nil {
                    log.Println(err)
                    w.WriteHeader(http.StatusInternalServerError)
                } else {
                    w.WriteHeader(http.StatusOK)
                }
            }
        } else {
            w.WriteHeader(http.StatusForbidden)
        }
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "POST":
        email := r.FormValue("email")
        if len(email) < 6 {
            http.Error(w, ErrEmailTooShort.Error(), http.StatusBadRequest)
            return
        }

        password := r.FormValue("password")
        if len(password) < 8 {
            http.Error(w, ErrPasswordTooShort.Error(), http.StatusBadRequest)
            return
        }

        firstname := r.FormValue("firstname")
        lastname := r.FormValue("lastname")

        imageURL := ""
        if destination, err := copyFile(r, "image", "content", randomFilename()); err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        } else {
            imageURL = destination
        }

        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
        if err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }

        user := &User{
            Firstname: firstname,
            Lastname:  lastname,
            Email:     email,
            Password:  string(hashedPassword),
            ImageURL:  imageURL,
        }

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
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

func userHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "PATCH":
        loggedIn, user := loggedIn(w, r, true)
        if !loggedIn {
            http.Error(w, ErrNotLoggedIn.Error(), http.StatusForbidden)
            return
        }
        
        // Set user firstname and lastname
        user.Firstname = r.FormValue("firstname")
        user.Lastname = r.FormValue("lastname")
        user.Email = r.FormValue("email")

        // Check if user is updating password
        oldPassword := r.FormValue("old-password")
        newPassword := r.FormValue("new-password")
        if user.Password != "" {
            // Process valid input (both passwords are at least the minimum length)
            if len(oldPassword) >= 8 && len(newPassword) >= 8 {
                // Check if old password matches
                if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
                    w.WriteHeader(http.StatusBadRequest)
                    return
                }

                // Create hashed password from new password
                hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
                if err != nil {
                    log.Println(err)
                    w.WriteHeader(http.StatusInternalServerError)
                    return
                }
                user.Password = string(hashedPassword)

            // Invalid input (at least one of the password is less than minimum length
            } else if len(oldPassword) > 0 && len(newPassword) > 0 {
                w.WriteHeader(http.StatusBadRequest)
                return
            
            // Ignore input (at least one of the two passwords is empty)
            } else {
                user.Password = ""
            }
        }

        // Update user avatar if necessary
        if destination, err := copyFile(r, "image", "content", randomFilename()); err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        } else {
            if user.ImageURL != "" {
                if err := os.Remove(user.ImageURL); err != nil {
                    log.Println(err)
                }
            }
            user.ImageURL = destination
        }

        if err := updateUser(user); err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
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
        ok, user := loggedIn(w, r, true)
        if !ok {
            w.WriteHeader(http.StatusForbidden)
            return
        }

        imageURL := r.FormValue("image_url")
        if imageURL == "" {
            if destination, err := copyFile(r, "image", "content", randomFilename()); err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            } else {
                imageURL = destination
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
        if err := tx.QueryRow(INSERT_HOOP_SQL, user.ID, name, description, latitude, longitude).Scan(&hoopID); err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }

        // Insert Story
        if err := tx.QueryRow(INSERT_STORY_SQL, hoopID, user.ID, name, description, imageURL).Scan(&storyID); err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
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
        var rows *sql.Rows
        var err error

        if name := r.FormValue("name"); name != "" {
            rows, err = db.Query(GET_HOOPS_WITH_NAME_SQL, "%"+name+"%")
        } else {
            rows, err = db.Query(GET_HOOPS_SQL)
        }

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
        ok, user := loggedIn(w, r, true)
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
            if destination, err := copyFile(r, "image", "content", randomFilename()); err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            } else {
                imageURL = destination
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

        var storyID int64

        // Insert Story
        if err := tx.QueryRow(INSERT_STORY_SQL, hoopID, user.ID, name, description, imageURL).Scan(&storyID); err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }

        // Insert Activity
        if _, err := tx.Exec(INSERT_POST_STORY_ACTIVITY_SQL, user.ID, ACTIVITY_POST_STORY, storyID); err != nil {
            log.Println("test", err)
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

        stories, err := getStories(GET_STORIES_SQL, hoopID)
        if err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
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
            activity.fetchData()

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

func commentHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "POST":
        ok, user := loggedIn(w, r, false)
        if !ok {
            w.WriteHeader(http.StatusForbidden)
            return
        }

        text := r.FormValue("text")
        if len(text) < 2 {
            w.WriteHeader(http.StatusBadRequest)
            return
        }

        hoopID := r.FormValue("hoopID")
        if hoopID != "" {
            hoopID, err := strconv.ParseInt(hoopID, 10, 64)
            if err != nil {
                w.WriteHeader(http.StatusBadRequest)
                return
            }

            // Start Transaction
            tx, err := db.Begin()
            if err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            }

            // Insert Comment
            if _, err = tx.Exec(INSERT_HOOP_COMMENT_SQL, user.ID, hoopID, text); err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            }

            // Insert Activity
            if _, err = tx.Exec(INSERT_HOOP_COMMENT_ACTIVITY_SQL, user.ID, ACTIVITY_POST_COMMENT_HOOP, hoopID); err != nil {
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
        } else if storyID := r.FormValue("storyID"); storyID != "" {
            storyID, err := strconv.ParseInt(storyID, 10, 64)
            if err != nil {
                w.WriteHeader(http.StatusBadRequest)
                return
            }

            // Start Transaction
            tx, err := db.Begin()
            if err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            }

            // Insert Comment
            if _, err = tx.Exec(INSERT_STORY_COMMENT_SQL, user.ID, hoopID, text); err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            }

            // Insert Activity
            if _, err = db.Exec(INSERT_STORY_COMMENT_ACTIVITY_SQL, user.ID, ACTIVITY_POST_COMMENT_STORY, storyID); err != nil {
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

        } else {
            w.WriteHeader(http.StatusBadRequest)
            return
        }

        w.WriteHeader(http.StatusOK)
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

func likeHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "POST":
        ok, user := loggedIn(w, r, true)
        if !ok {
            w.WriteHeader(http.StatusForbidden)
            return
        }

        if hoopID := r.FormValue("hoop-id"); hoopID != "" {
            hoopID, err := strconv.ParseInt(hoopID, 10, 64)
            if err != nil {
                w.WriteHeader(http.StatusBadRequest)
                return
            }

            if err := toggleLike(user.ID, hoopID, "hoop"); err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            }

        } else if storyID := r.FormValue("story-id"); storyID != "" {
            storyID, err := strconv.ParseInt(storyID, 10, 64)
            if err != nil {
                w.WriteHeader(http.StatusBadRequest)
                return
            }

            if err := toggleLike(user.ID, storyID, "story"); err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            }
        } else {
            w.WriteHeader(http.StatusBadRequest)
            return
        }

        w.WriteHeader(http.StatusOK)
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

func commentsHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        var comments []Comment

        hoopID := r.FormValue("hoop_id")
        if hoopID != "" {
            hoopID, err := strconv.ParseInt(hoopID, 10, 64)
            if err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            }

            rows, err := db.Query(GET_HOOP_COMMENTS_SQL, hoopID)
            if err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            }

            var text sql.NullString

            for rows.Next() {
                var comment Comment

                if err := rows.Scan(
                    &comment.UserID,
                    &text,
                    &comment.CreatedAt,
                    &comment.UpdatedAt,
                ); err != nil {
                    log.Println(err)
                    w.WriteHeader(http.StatusInternalServerError)
                    return
                }
                comment.Text = fromNullString(text)

                comments = append(comments, comment)
            }
        } else if storyID := r.FormValue("story_id"); storyID != "" {
            storyID, err := strconv.ParseInt(storyID, 10, 64)
            if err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            }

            rows, err := db.Query(GET_STORY_COMMENTS_SQL, storyID)
            if err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            }

            var text sql.NullString

            for rows.Next() {
                var comment Comment

                if err := rows.Scan(
                    &comment.UserID,
                    &text,
                    &comment.CreatedAt,
                    &comment.UpdatedAt,
                ); err != nil {
                    log.Println(err)
                    w.WriteHeader(http.StatusInternalServerError)
                    return
                }
                comment.Text = fromNullString(text)

                comments = append(comments, comment)
            }
        } else {
            w.WriteHeader(http.StatusBadRequest)
            return
        }

        data, err := json.Marshal(comments)
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

func likesHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        var likes []Like

        hoopID := r.FormValue("hoop-id")
        if hoopID != "" {
            hoopID, err := strconv.ParseInt(hoopID, 10, 64)
            if err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            }

            rows, err := db.Query(GET_HOOP_LIKES_SQL, hoopID)
            if err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            }

            for rows.Next() {
                var like Like

                if err := rows.Scan(
                    &like.UserID,
                    &like.CreatedAt,
                    &like.UpdatedAt,
                ); err != nil {
                    log.Println(err)
                    w.WriteHeader(http.StatusInternalServerError)
                    return
                }

                likes = append(likes, like)
            }
        } else if storyID := r.FormValue("story-id"); storyID != "" {
            storyID, err := strconv.ParseInt(storyID, 10, 64)
            if err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            }

            rows, err := db.Query(GET_STORY_LIKES_SQL, storyID)
            if err != nil {
                log.Println(err)
                w.WriteHeader(http.StatusInternalServerError)
                return
            }

            for rows.Next() {
                var like Like

                if err := rows.Scan(
                    &like.UserID,
                    &like.CreatedAt,
                    &like.UpdatedAt,
                ); err != nil {
                    log.Println(err)
                    w.WriteHeader(http.StatusInternalServerError)
                    return
                }

                likes = append(likes, like)
            }
        } else {
            w.WriteHeader(http.StatusBadRequest)
            return
        }

        data, err := json.Marshal(likes)
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

func likesCountHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        var count int64

        if hoopID, err := strconv.ParseInt(r.FormValue("hoop-id"), 10, 64); err == nil {
            if err := db.QueryRow(COUNT_HOOP_LIKES_SQL, hoopID).Scan(&count); err == nil {
                w.Write([]byte(strconv.FormatInt(count, 10)))
                return
            }
        } else if storyID, err := strconv.ParseInt(r.FormValue("story-id"), 10, 64); err == nil {
            if err := db.QueryRow(COUNT_STORY_LIKES_SQL, storyID).Scan(&count); err == nil {
                w.Write([]byte(strconv.FormatInt(count, 10)))
                return
            }
        }

        w.WriteHeader(http.StatusBadRequest)
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

func nearbyHoopsHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        var hoops []Hoop

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

        radius, err := strconv.ParseFloat(r.FormValue("radius"), 64)
        if err != nil {
            radius = 100
        }

        rows, err := db.Query(GET_NEARBY_HOOPS_SQL, latitude, latitude, longitude, radius)
        if err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
        defer rows.Close()

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

func popularHoopsHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        var hoops []Hoop

        rows, err := db.Query(GET_POPULAR_HOOPS_SQL)
        if err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
        defer rows.Close()

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

func latestHoopsHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        var hoops []Hoop

        rows, err := db.Query(GET_LATEST_HOOPS_SQL)
        if err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
        defer rows.Close()

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

func mostLikedStoriesHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        hoopID, err := strconv.ParseInt(r.FormValue("hoop_id"), 10, 64)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }

        stories, err := getStories(GET_MOST_LIKED_STORIES_SQL, hoopID)
        if err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
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

func mostViewedStoriesHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        hoopID, err := strconv.ParseInt(r.FormValue("hoop_id"), 10, 64)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }

        stories, err := getStories(GET_MOST_VIEWED_STORIES_SQL, hoopID)
        if err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
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

func latestStoriesHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        hoopID, err := strconv.ParseInt(r.FormValue("hoop_id"), 10, 64)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }

        stories, err := getStories(GET_LATEST_STORIES_SQL, hoopID)
        if err != nil {
            log.Println(err)
            w.WriteHeader(http.StatusInternalServerError)
            return
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

func copyFile(r *http.Request, name string, folder, filename string) (destination string, err error) {
    var fileheader *multipart.FileHeader

    if _, fileheader, err = r.FormFile("image"); err != nil {
        return
    } else {
        var infile multipart.File
        var outfile *os.File

        if err = os.MkdirAll(folder, os.ModeDir|0775); err != nil {
            return
        }

        destination = folder + "/" + filename

        // Open received file
        if infile, err = fileheader.Open(); err != nil {
            return
        }
        defer infile.Close()

        // Create destination file
        if outfile, err = os.OpenFile(destination, os.O_CREATE|os.O_WRONLY, 0664); err != nil {
            return
        }
        defer outfile.Close()

        // Copy file to destination
        if _, err = io.Copy(outfile, infile); err != nil {
            return
        }
    }

    return
}

func randomFilename() string {
    cmd := exec.Command("openssl", "rand", "-base64", "64")

    output, err := cmd.Output()
    if err != nil {
        log.Fatal(err)
    }

    for i := range output {
        if output[i] == '/' {
            output[i] = '-'
        }
    }

    return string(output)
}
