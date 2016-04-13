package main

import (
    "log"
    "net/http"
)

func loggedIn(w http.ResponseWriter, r *http.Request) bool {
    session, err := ss.Get(r, "session")
    if err != nil {
        log.Println(err)
        return false
    }

    val := session.Values["user"]
    if user, ok := val.(User); !ok {
        return false
    } else if exists, _ := userExists(&user, false); !exists {
        return false
    }

    return true
}

func logIn(w http.ResponseWriter, r *http.Request, user *User) error {
    session, err := ss.Get(r, "session")
    if err != nil {
        return err
    }

    session.Values["user"] = *user
    return nil
}

func logOut(w http.ResponseWriter, r *http.Request) error {
    session, err := ss.Get(r, "session")
    if err != nil {
        return err
    }

    session.Values["user"] = User{}
    return nil
}
