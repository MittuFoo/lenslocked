package main

import (
	"fmt"
	"net/http"
	"github.com/joho/godotenv"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"lenslockd/controllers"
	"lenslockd/templates"
	"lenslockd/views"
	"lenslockd/models"
	"lenslockd/migrations"

	"github.com/gorilla/csrf"
)

type config struct {
	PSQL models.PostgresConfig
	SMTP models.SMTPConfig
	CSRF struct {
		Key    string
		Secure bool
	}
	Server struct {
		Address string
	}
}


func loadEnvConfig() (config, error) {
	var cfg config
	err := godotenv.Load()
	if err != nil {
		return cfg, err
	}
	// TODO: Read the PSQL values from an ENV variable
	cfg.PSQL = models.PostgresConfig{
		Host:     os.Getenv("PSQL_HOST"),
		Port:     os.Getenv("PSQL_PORT"),
		User:     os.Getenv("PSQL_USER"),
		Password: os.Getenv("PSQL_PASSWORD"),
		Database: os.Getenv("PSQL_DATABASE"),
		SSLMode:  os.Getenv("PSQL_SSLMODE"),
	}
	if cfg.PSQL.Host == "" && cfg.PSQL.Port == "" {
		return cfg, fmt.Errorf("No PSQL Config provided.")
	}
	// TODO: Setup SMTP config
	cfg.SMTP.Host = os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	cfg.SMTP.Port, err = strconv.Atoi(portStr)
	if err != nil {
		return cfg, err
	}
	cfg.SMTP.Username = os.Getenv("SMTP_USERNAME")
	cfg.SMTP.Password = os.Getenv("SMTP_PASSWORD")
	
	cfg.CSRF.Key = os.Getenv("CSRF_KEY")
	cfg.CSRF.Secure = os.Getenv("CSRF_SECURE") == "true"

	cfg.Server.Address = os.Getenv("SERVER_ADDRESS")
	return cfg, nil
}

func main() {
	cfg, err := loadEnvConfig()
  		if err != nil {
    	panic(err)
  	}
	// Set up a database connection
	db, err := models.Open(cfg.PSQL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = models.MigrateFS(db, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	// Set up services
	userService := models.UserService{
  		DB: db,
	}
	sessionService := models.SessionService{
  		DB: db,
	}
	pwResetService := models.PasswordResetService{
  		DB: db,
	}
	emailService := models.NewEmailService(cfg.SMTP)
	galleryService := &models.GalleryService{
    DB: db,
  	}

	// Set up middleware
	umw := controllers.UserMiddleware{
  		SessionService: &sessionService,
	}

	csrfMw := csrf.Protect(
  		[]byte(cfg.CSRF.Key),
  		csrf.Secure(cfg.CSRF.Secure),
		csrf.Path("/"),
	)

	// Set up controllers
	usersC := controllers.Users{
  		UserService:          &userService,
  		SessionService:       &sessionService,
  		PasswordResetService: &pwResetService,
  		EmailService:         emailService,
	}
	usersC.Templates.New = views.Must(views.ParseFS(
  		templates.FS,
		"signup.gohtml", "tailwind.gohtml"))
	usersC.Templates.SignIn = views.Must(views.ParseFS(
  		templates.FS,
		"signin.gohtml", "tailwind.gohtml"))
	usersC.Templates.ForgotPassword = views.Must(views.ParseFS(
  		templates.FS,
		"forgot-pw.gohtml", "tailwind.gohtml"))
	usersC.Templates.CheckYourEmail = views.Must(views.ParseFS(
  		templates.FS,
		"check-your-email.gohtml", "tailwind.gohtml",))
	usersC.Templates.ResetPassword = views.Must(views.ParseFS(
  		templates.FS,
  		"reset-pw.gohtml", "tailwind.gohtml",))
	galleriesC := controllers.Galleries{
		GalleryService: galleryService,
	}
	galleriesC.Templates.New = views.Must(views.ParseFS(
		templates.FS,
		"galleries/new.gohtml", "tailwind.gohtml",
	))
	galleriesC.Templates.Edit = views.Must(views.ParseFS(
    	templates.FS,
    	"galleries/edit.gohtml", "tailwind.gohtml",
  	))
	galleriesC.Templates.Index = views.Must(views.ParseFS(
  		templates.FS,
  		"galleries/index.gohtml", "tailwind.gohtml",
	))
	galleriesC.Templates.Show = views.Must(views.ParseFS(
  		templates.FS,
  		"galleries/show.gohtml", "tailwind.gohtml",
	))

	// Set up router and routes
	r := chi.NewRouter()
	// These middleware are used everywhere.
	r.Use(csrfMw)
	r.Use(umw.SetUser)
	// Now we set up routes.
	r.Get("/", controllers.StaticHandler(views.Must(views.ParseFS(
  		templates.FS,
  		"home.gohtml", "tailwind.gohtml",
	))))
	// ... our routes remain unchanged

	r.Get("/", controllers.StaticHandler(views.Must(views.ParseFS(
  		templates.FS,
  		"home.gohtml", "tailwind.gohtml",
	))))
	r.Get("/contact", controllers.StaticHandler(views.Must(views.ParseFS(
  		templates.FS,
  		"contact.gohtml", "tailwind.gohtml",
	))))
	r.Get("/faq", controllers.FAQ(views.Must(views.ParseFS(
  		templates.FS,
  		"faq.gohtml", "tailwind.gohtml",
	))))
	r.Get("/signup", usersC.New)
	r.Post("/signup", usersC.Create)
	r.Get("/signin", usersC.SignIn)
	r.Post("/signin", usersC.ProcessSignIn)
	r.Post("/signout", usersC.ProcessSignOut)
	r.Get("/forgot-pw", usersC.ForgotPassword)
	r.Post("/forgot-pw", usersC.ProcessForgotPassword)
	r.Get("/reset-pw", usersC.ResetPassword)
	r.Post("/reset-pw", usersC.ProcessResetPassword)
	r.Route("/galleries", func(r chi.Router) {
		r.Get("/{id}", galleriesC.Show)
		r.Get("/{id}/images/{filename}", galleriesC.Image)
		r.Group(func(r chi.Router) {
			r.Use(umw.RequireUser)
			r.Get("/", galleriesC.Index)
			r.Get("/new", galleriesC.New)
			r.Post("/", galleriesC.Create)
			r.Get("/{id}/edit", galleriesC.Edit)
			r.Post("/{id}", galleriesC.Update)
			r.Post("/{id}/delete", galleriesC.Delete)
			r.Post("/{id}/images/{filename}/delete", galleriesC.DeleteImage)
			r.Post("/{id}/images", galleriesC.UploadImage)
		})
	})
	r.Route("/users/me", func(r chi.Router) {
	  	r.Use(umw.RequireUser)
  		r.Get("/", usersC.CurrentUser)
	})
	assetsHandler := http.FileServer(http.Dir("assets"))
r.Get("/assets/*", http.StripPrefix("/assets", assetsHandler).ServeHTTP)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
  		http.Error(w, "Page not found", http.StatusNotFound)
	})

	// Start the server
	fmt.Printf("Starting the server on %s...\n", cfg.Server.Address)
	err = http.ListenAndServe(cfg.Server.Address, r)
	if err != nil {
  		panic(err)
	}
}