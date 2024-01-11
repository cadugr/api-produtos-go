package main

import (
	"log"
	"net/http"

	"github.com/cadugr/api-produtos/configs"
	_ "github.com/cadugr/api-produtos/docs"
	"github.com/cadugr/api-produtos/internal/entity"
	"github.com/cadugr/api-produtos/internal/infra/database"
	"github.com/cadugr/api-produtos/internal/infra/webserver/handlers"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// @title Go Expert API Example
// @version 1.0
// @description Product API with authentication
// @termsOfService http://swagger.io/terms/

// @contact.name Cadu
// @contact.url http://www.fullcycle.com.br
// @license.name Full Cycle License
// @license.url http://www.fullcycle.com.br

// @host localhost:8000
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	configs, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&entity.Product{}, entity.User{})
	productDB := database.NewProduct(db)
	productHandler := handlers.NewProductHandler(productDB)

	userDB := database.NewUser(db)
	userHandler := handlers.NewUserHandler(userDB)

	r := chi.NewRouter()
	r.Use(middleware.Logger)    //Adicionando log para todas as rotas
	r.Use(middleware.Recoverer) //Sempre utilizar para evitar da aplicação cair em caso de algum problema
	r.Use(middleware.WithValue("jwt", configs.TokenAuth))
	r.Use(middleware.WithValue("JwtExpiresIn", configs.JwtExpiresIn))
	//r.Use(LogRequest)

	r.Route("/products", func(r chi.Router) { //Coloca o prefixo (/products) para todas as rotas de products
		r.Use(jwtauth.Verifier(configs.TokenAuth)) //Pega o token enviado na requisição, esteja ele aonde estiver (header, url, etc) e coloca no contexto do chi
		r.Use(jwtauth.Authenticator)               //Valida se o token é válido, ou seja, se não está expirado e se o secret é o mesmo que foi usado para assinar o token
		r.Post("/", productHandler.CreateProduct)
		r.Get("/", productHandler.GetProducts)
		r.Get("/{id}", productHandler.GetProduct)
		r.Put("/{id}", productHandler.UpdateProduct)
		r.Delete("/{id}", productHandler.DeleteProduct)
	})

	r.Post("/users", userHandler.CreateUser)
	r.Post("/users/generate_token", userHandler.GetJWT)

	r.Get("/docs/*", httpSwagger.Handler(httpSwagger.URL("http://localhost:8000/docs/doc.json")))

	http.ListenAndServe(":8000", r)
}

func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r) //Passa o controle adiante
	})
}

//Request -> Middleware(usa os dados, faz alguma coisa e continua)| outro middleware -> Handler -> Response
