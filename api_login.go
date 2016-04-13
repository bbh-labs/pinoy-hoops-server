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

    val := session.Values["userID"]
    if userID, ok := val.(int64); !ok {
        log.Println("Failed to get user from session")
        return false
    } else if exists, _ := userExists(&User{ID: userID}, false); !exists {
        log.Println("User doesn't exist")
        return false
    }

    return true
}

func logIn(w http.ResponseWriter, r *http.Request, user *User) error {
    session, err := ss.Get(r, "session")
    if err != nil {
        return err
    }

    session.Values["userID"] = user.ID
    session.Save(r, w)
    return nil
}

func logOut(w http.ResponseWriter, r *http.Request) error {
    session, err := ss.Get(r, "session")
    if err != nil {
        return err
    }

    session.Values["userID"] = 0
    session.Save(r, w)
    return nil
}
