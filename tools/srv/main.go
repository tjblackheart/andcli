package main

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"image/png"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/tyler-sommer/stick"
	"golang.org/x/crypto/bcrypt"
)

const (
	addr          = "127.0.0.1:8080"
	issuer        = "otp.provider.dev"
	otpAlg        = otp.AlgorithmSHA1
	otpLen        = otp.DigitsSix
	otpSecretSize = 32
)

type (
	user struct {
		Username     string
		Password     string
		Secret       string
		OtpConfirmed bool
	}

	flash struct{ Message, Type string }

	m map[string]stick.Value

	db struct {
		*sync.Mutex
		users map[string]*user
	}
)

var (
	env     = stick.New(stick.NewFilesystemLoader("./assets/templates"))
	store   = &db{new(sync.Mutex), make(map[string]*user)}
	session = scs.New()
)

/// app

func main() {
	gob.Register(&user{})
	gob.Register(&flash{})

	s := http.Server{Addr: addr, Handler: session.LoadAndSave(mux())}

	log.Printf("Server started: http://%s\n", addr)
	log.Fatal(s.ListenAndServe())
}

func mux() chi.Router {
	mux := chi.NewMux()
	mux.Use(middleware.Logger, middleware.Recoverer)

	mux.Route("/", func(r chi.Router) {
		r.Get("/", getLogin)
		r.Get("/login", getLogin)
		r.Get("/register", getRegister)
		r.Post("/login", postLogin)
		r.Post("/register", postRegister)
	})

	mux.Route("/otp", func(r chi.Router) {
		r.Use(isLoggedIn)
		r.Get("/", getOTP)
		r.Post("/", postOTP)
		r.Get("/new", createOTP)
	})

	mux.Route("/user", func(r chi.Router) {
		r.Use(isLoggedIn, isOTPValidated)
		r.Get("/", getUser)
		r.Get("/logout", getLogout)
	})

	fs := http.FileServer(http.Dir("./assets/css"))
	mux.Handle("/assets/*", http.StripPrefix("/assets/", fs))

	return mux
}

/// handlers

func getLogin(w http.ResponseWriter, r *http.Request) {
	if u, _ := session.Get(r.Context(), "user").(*user); u != nil {
		if session.GetBool(r.Context(), "otp-validated") {
			http.Redirect(w, r, "/user", http.StatusFound)
			return
		}

		if !u.OtpConfirmed {
			http.Redirect(w, r, "/otp/new", http.StatusFound)
			return
		}
	}

	flash, _ := session.Pop(r.Context(), "flash").(*flash)
	env.Execute("pages/login.html.twig", w, m{"flash": flash})
}

func postLogin(w http.ResponseWriter, r *http.Request) {

	u, err := store.Find(r.PostFormValue("username"))
	if err != nil {
		log.Println(err)
		session.Put(r.Context(), "flash", &flash{"invalid credentials!", "danger"})
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	have, want := []byte(u.Password), []byte(r.PostFormValue("password"))
	if err := bcrypt.CompareHashAndPassword(have, want); err != nil {
		session.Put(r.Context(), "flash", &flash{"invalid credentials!", "danger"})
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	session.Put(r.Context(), "user", u)
	if u.Secret == "" || !u.OtpConfirmed {
		http.Redirect(w, r, "/otp/new", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/otp", http.StatusFound)
}

func getRegister(w http.ResponseWriter, r *http.Request) {
	flash, _ := session.Pop(r.Context(), "flash").(*flash)
	env.Execute("pages/register.html.twig", w, m{"flash": flash})
}

func postRegister(w http.ResponseWriter, r *http.Request) {
	u := strings.TrimSpace(r.PostFormValue("username"))
	p := strings.TrimSpace(r.PostFormValue("password"))

	if u == "" || p == "" {
		session.Put(r.Context(), "flash", &flash{"invalid user data!", "danger"})
		http.Redirect(w, r, "/register", http.StatusFound)
		return
	}

	b, err := bcrypt.GenerateFromPassword([]byte(p), 12)
	if err != nil {
		log.Println(err)
		session.Put(r.Context(), "flash", &flash{"internal error!", "danger"})
		http.Redirect(w, r, "/register", http.StatusFound)
		return
	}

	user := &user{Username: u, Password: string(b)}
	store.Insert(user)

	session.Put(r.Context(), "user", user)
	http.Redirect(w, r, "/otp/new", http.StatusFound)
}

func createOTP(w http.ResponseWriter, r *http.Request) {
	u, _ := session.Get(r.Context(), "user").(*user)
	if u.Secret != "" && u.OtpConfirmed {
		http.Redirect(w, r, "/otp", http.StatusFound)
		return
	}

	if session.GetBool(r.Context(), "otp-validated") {
		http.Redirect(w, r, "/user", http.StatusFound)
		return
	}

	if _, err := store.Find(u.Username); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}

	opts := totp.GenerateOpts{
		Issuer:      issuer,
		Algorithm:   otpAlg,
		AccountName: u.Username,
		SecretSize:  otpSecretSize,
		Digits:      otp.DigitsSix,
	}

	otp, err := totp.Generate(opts)
	if err != nil {
		log.Println("otp:", err)
		session.Put(r.Context(), "flash", &flash{err.Error(), "danger"})
		http.Redirect(w, r, "/otp/new", http.StatusFound)
		return
	}

	u.Secret = otp.Secret()
	if err := store.Update(u); err != nil {
		log.Println(err)
		session.Put(r.Context(), "flash", &flash{err.Error(), "danger"})
		http.Redirect(w, r, "/otp/new", http.StatusFound)
		return
	}

	buf := new(bytes.Buffer)
	qr, err := otp.Image(240, 240)
	if err != nil {
		session.Put(r.Context(), "flash", &flash{err.Error(), "danger"})
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if err := png.Encode(buf, qr); err != nil {
		session.Put(r.Context(), "flash", &flash{err.Error(), "danger"})
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	flash, _ := session.Pop(r.Context(), "flash").(*flash)
	env.Execute("pages/2fa-create.html.twig", w, m{
		"qr":    base64.RawStdEncoding.EncodeToString(buf.Bytes()),
		"flash": flash,
	})
}

func getOTP(w http.ResponseWriter, r *http.Request) {
	if u, _ := session.Get(r.Context(), "user").(*user); !u.OtpConfirmed {
		http.Redirect(w, r, "/otp/new", http.StatusFound)
		return
	}

	flash, _ := session.Pop(r.Context(), "flash").(*flash)
	env.Execute("pages/2fa.html.twig", w, m{"flash": flash})
}

func postOTP(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.PostFormValue("token"))
	u, _ := session.Get(r.Context(), "user").(*user)

	user, err := store.Find(u.Username)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}

	opts := totp.ValidateOpts{Algorithm: otpAlg, Digits: otpLen}
	if valid, err := totp.ValidateCustom(token, user.Secret, time.Now().UTC(), opts); !valid || err != nil {
		if err != nil {
			log.Println("otp: validate:", err)
		}
		session.Put(r.Context(), "flash", &flash{"invalid token, try again.", "danger"})
		http.Redirect(w, r, "/otp/new", http.StatusFound)
		return
	}

	user.OtpConfirmed = true
	if err := store.Update(user); err != nil {
		log.Println(err)
		session.Put(r.Context(), "flash", &flash{err.Error(), "danger"})
		http.Redirect(w, r, "/otp/new", http.StatusFound)
		return
	}

	session.Put(r.Context(), "otp-validated", true)

	http.Redirect(w, r, "/user", http.StatusFound)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	u, _ := session.Get(r.Context(), "user").(*user)
	env.Execute("pages/user.html.twig", w, m{"user": u})
}

func getLogout(w http.ResponseWriter, r *http.Request) {
	session.Destroy(r.Context())
	http.Redirect(w, r, "/", http.StatusFound)
}

/// middleware

func isLoggedIn(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, _ := session.Get(r.Context(), "user").(*user); u == nil {
			session.Put(r.Context(), "flash", &flash{"please login first!", "danger"})
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isOTPValidated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !session.GetBool(r.Context(), "otp-validated") {
			session.Put(r.Context(), "flash", &flash{"invalid otp!", "danger"})
			http.Redirect(w, r, "/otp", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

/// "store"

func (db db) Find(username string) (*user, error) {
	db.Lock()
	defer db.Unlock()

	if u, ok := db.users[username]; ok {
		return u, nil
	}

	return nil, fmt.Errorf("db: find(): user %s not found", username)
}

func (db db) Insert(u *user) {
	db.Lock()
	defer db.Unlock()

	db.users[u.Username] = u
}

func (db db) Update(u *user) error {
	db.Lock()
	defer db.Unlock()

	if _, ok := db.users[u.Username]; !ok {
		return fmt.Errorf("db: update(): user %s not found", u.Username)
	}

	db.users[u.Username].OtpConfirmed = u.OtpConfirmed
	db.users[u.Username].Password = u.Password
	db.users[u.Username].Secret = u.Secret

	return nil
}
