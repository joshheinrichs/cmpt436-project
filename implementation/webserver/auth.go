/*
 * Uses goth/gothic for authentication, and also makes use of the session that
 * gothic uses (so that there are not two sessions being used.)
 */
package webserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/pat"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/gplus"
	"golang.org/x/oauth2/google"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

// SESSION_NAME is the key used to access the session store.
const (
	SESSION_DURATION_MINUTES         int    = 30
	USER_KEY                         string = "goth_user"
	SESSION_SECRET_CONFIG_FILE_PATH         = "../../.session_secret"
	SESSION_NAME                            = "user_session"
	SESSION_KEY_USERNAME                    = "username"
	GOOGLE_CLIENT_SECRET_FILE_PATH          = "../../.gplus_client_secret.json"
	FACEBOOK_CLIENT_SECRET_FILE_PATH        = "../../.facebook_client_secret.json"
	AUTH_CALLBACK_RELATIVE_PATH             = "/oauth2callback"
)

func marshalUser(user *goth.User) (string, error) {
	b, err := json.Marshal(user)
	return string(b), err
}

func unmarshalUser(data string) (*goth.User, error) {
	user := &goth.User{}
	err := json.Unmarshal([]byte(data), user)
	return user, err
}

func getUserFromSession(s *sessions.Session) (*goth.User, error) {
	val := s.Values[USER_KEY]
	if val == nil {
		return nil, errors.New("user not stored in session")
	}
	userString := val.(string)
	return unmarshalUser(userString)
}

/*
 * Does not save the session.
 */
func putUserInSession(user *goth.User, s *sessions.Session) error {
	userString, err := marshalUser(user)
	if err != nil {
		return err
	}
	s.Values[USER_KEY] = userString
	return nil
}

func initAuth(router *pat.Router) {
	//get all the providers set up.
	googleJsonKey, err := ioutil.ReadFile(GOOGLE_CLIENT_SECRET_FILE_PATH)
	if err != nil {
		log.Fatalln("unable to read file ", GOOGLE_CLIENT_SECRET_FILE_PATH,
			":", err)
	}
	facebookJsonKey, err := ioutil.ReadFile(FACEBOOK_CLIENT_SECRET_FILE_PATH)
	if err != nil {
		log.Fatalln("unable to read file ", FACEBOOK_CLIENT_SECRET_FILE_PATH,
			":", err)
	}

	// do I need more scopes?
	// https://developers.google.com/+/domains/authentication/scopes
	googleConfig, err := google.ConfigFromJSON(googleJsonKey)
	if err != nil {
		log.Fatalln("unable to get google provider config:", err)
	}
	facebookConfig := &genericConfig{}
	err = json.Unmarshal(facebookJsonKey, facebookConfig)
	if err != nil {
		log.Fatalln("unable to get facebook provider config:", err)
	}

	AUTH_CALLBACK_PATH := fmt.Sprint(DOMAIN_NAME, AUTH_CALLBACK_RELATIVE_PATH)
	//I need "profile", "email", scopes. gplus and facebook provide these by
	//default.
	goth.UseProviders(
		gplus.New(googleConfig.ClientID, googleConfig.ClientSecret,
			fmt.Sprint(AUTH_CALLBACK_PATH, "/gplus")),
		facebook.New(facebookConfig.Client_id, facebookConfig.Client_secret,
			fmt.Sprint(AUTH_CALLBACK_PATH, "/facebook")),
	)

	//initialize the gothic store.
	key, err := ioutil.ReadFile(SESSION_SECRET_CONFIG_FILE_PATH)
	if err != nil {
		log.Println("could not load session secret from file ",
			SESSION_SECRET_CONFIG_FILE_PATH)
		log.Fatalln(err)
	}
	gothic.Store = sessions.NewCookieStore([]byte(key))
	gothic.Store.(*sessions.CookieStore).Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * SESSION_DURATION_MINUTES,
		HttpOnly: true,
		Secure:   true,
	}

	log.Println()
	router.Get(fmt.Sprint(AUTH_CALLBACK_RELATIVE_PATH, "/{provider}"),
		authCallbackHandler)
	router.Get("/auth/{provider}", gothic.BeginAuthHandler)
	router.Get("/", authHandler)
}

func validateSessionAndLogInIfNecessary(
	w http.ResponseWriter, r *http.Request) *sessions.Session {
	session, err := validateSession(w, r)
	if session == nil {
		if err != nil {
			log.Println(err.Error())
		}
		serveNewLogin(w, r)
	}

	return session
}

/**
 * return a session pointer. It is nil if the session could not be validated
 * (and thus the session is unauthorized). An error is also returned, if one
 * exists.
 */
func validateSession(
	w http.ResponseWriter, r *http.Request) (*sessions.Session, error) {
	session, err := gothic.Store.Get(r, SESSION_NAME)
	log.Println("validating session...")

	if err != nil {
		log.Println("unable to get session.")
		return nil, err
	}

	if session.IsNew {
		endSession(session, w, r)
		return nil, nil
	}

	_, err = getUserFromSession(session)
	if err != nil {
		log.Println("unable to unmarshal user from session.")
		endSession(session, w, r)
		return nil, err
	}

	return session, nil
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("serving auth request")

	session, err := validateSession(w, r)
	if err != nil {
		log.Println(err)
	} else if session != nil {
		http.Redirect(w, r, "/app", http.StatusMovedPermanently)
		return
	}

	serveNewLogin(w, r)
}

func serveNewLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("serving new login.")
	t, err := template.New("login").Parse(indexTemplate)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	_, err = gothic.Store.Get(r, SESSION_NAME)
	if err != nil {
		http.Error(w, "unable to get session", 500)
		log.Println(err.Error())
	}

	t.Execute(w, nil)
}

func authCallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("serving auth callback")

	//TODO make use of more user attributes, besides name.
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	session, err := gothic.Store.Get(r, SESSION_NAME)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	// log.Printf("number of values already in new session: %d.\n",
	// len(session.Values))

	err = putUserInSession(&user, session)
	if err != nil {
		http.Error(w, "unable to store user in session", 500)
		endSession(session, w, r)
		return
	}

	session.Save(r, w)
	http.Redirect(w, r, "/app", http.StatusMovedPermanently)
}

func endSession(s *sessions.Session, w http.ResponseWriter, r *http.Request) {
	s.Options = &sessions.Options{MaxAge: -1}
	s.Save(r, w)
}