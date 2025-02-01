package main

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Althaf66/go-backend-social-app/docs"
	"github.com/Althaf66/go-backend-social-app/internal/auth"
	"github.com/Althaf66/go-backend-social-app/internal/mailer"
	"github.com/Althaf66/go-backend-social-app/internal/store"
	"github.com/Althaf66/go-backend-social-app/internal/store/cache"
	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

type application struct {
	config        config
	store         store.Storage
	logger        *zap.SugaredLogger
	mailer        mailer.Client
	authenticator auth.Authenticator
	cacheStorage  cache.Storage
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type config struct {
	addr        string
	db          dbConfig
	env         string
	apiUrl      string
	mail        mailconfig
	frontendURL string
	auth        authConfig
	redisCfg    redisConfig
}

type redisConfig struct {
	addr    string
	pw      string
	db      int
	enabled bool
}

type authConfig struct {
	basic basicConfig
	token tokenConfig
}

type tokenConfig struct {
	secret string
	exp    time.Duration
	iss    string
}

type basicConfig struct {
	username string
	password string
}

type mailconfig struct {
	exp time.Duration
	// sendGrid  sendGridConfig
	mailTrap  mailTrapConfig
	fromEmail string
}

type mailTrapConfig struct {
	apiKey string
}

// type sendGridConfig struct {
// 	apiKey string
// }

func (app *application) mount() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Use(middleware.Timeout(time.Minute))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthcheckHandler)
		r.With(app.BasicAuthMiddleware()).Get("/debug/vars", expvar.Handler().ServeHTTP)

		docsUrl := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL(docsUrl),
		))

		r.Route("/posts", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)
			r.Post("/", app.createPostHandler)

			r.Route("/{postID}", func(r chi.Router) {
				r.Use(app.postsContextMiddleware)
				r.Get("/", app.getPostHandler)

				r.Patch("/", app.checkPostOwnership("moderator", app.updatePostHandler))
				r.Delete("/", app.checkPostOwnership("admin", app.deletePostHandler))
			})
		})
		r.Route("/users", func(r chi.Router) {
			r.Put("/activate/{token}", app.activateUserHandler)

			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)

				r.Get("/", app.getUserHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})
			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Get("/feed", app.getUserFeed)
			})

		})
		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
			r.Post("/token", app.createTokenHandler)
		})
	})

	return r
}

func (app *application) run(mux *chi.Mux) error {

	docs.SwaggerInfo.Version = "0.0.1"
	docs.SwaggerInfo.Host = app.config.apiUrl
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}
	shutdown := make(chan error)

	go func(){
		quit := make(chan os.Signal,1)
		signal.Notify(quit,syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		ctx,cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		app.logger.Infow("signal caught","signal", s.String())
		shutdown <- srv.Shutdown(ctx)
	}()

	app.logger.Info("server has started ", "addr", app.config.addr, " env:", app.config.env)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdown
	if err != nil {
		return err
	}

	app.logger.Infow("server has stopped", "addr", app.config.addr, "env", app.config.env)

	return nil
}
