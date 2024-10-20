package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"net/http"
	"users_mrc/config"
	db "users_mrc/db/sqlc"
	"users_mrc/infrastructure/jwt_token"
	"users_mrc/infrastructure/worker"
	"users_mrc/interfaces/controllers"
	"users_mrc/interfaces/middleware"
	"users_mrc/interfaces/routers"
	usecase1 "users_mrc/usecase"
)

type Server struct {
	config      config.Config
	store       db.Store
	router      *gin.Engine
	tokenMaker  jwt_token.Maker
	distributor worker.TaskDistributor
}

func NewServer(config config.Config, store db.Store, distributor worker.TaskDistributor) (*Server, error) {
	tokenMaker, err := jwt_token.NewJWTMaker(
		config.TokenSymmetricKey,
		config.AccessTokenExpiresIn,
		config.RefreshTokenExpiresIn,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:      config,
		store:       store,
		tokenMaker:  tokenMaker,
		distributor: distributor,
	}
	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()
	//router.Use(middleware.OpenCORSMiddleware())
	//router.Use(middleware.HandleSessionMiddleware(server.store))
	router.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Route %s not found", ctx.Request.URL)})
	})
	router.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	api := router.Group("api")
	v1 := api.Group("/v1")
	//adminGroup := v1.Group("/admin")
	server.setupAuthRoutes(v1)
	server.setupUsersRoutes(v1)
	server.router = router
}

func (server *Server) setupAuthRoutes(rg *gin.RouterGroup) {
	//jwtDeserializer := middleware.JWTDeserializer(server.tokenMaker)
	usecase := usecase1.NewAuthUsecase(server.store, server.tokenMaker)
	controller := controllers.NewAuthController(usecase, server.distributor)
	route := routers.NewAuthRouter(controller)
	router := rg.Group("/auth")
	public := router.Group("/public")
	private := router.Group("/private")
	//private.Use(jwtDeserializer)
	route.InitAuthRouter(public, private)
}

func (server *Server) setupUsersRoutes(rg *gin.RouterGroup) {
	jwtDeserializer := middleware.JWTDeserializer(server.tokenMaker)
	usecase := usecase1.NewUsersUsecase(server.store)
	controller := controllers.NewUsersController(usecase)
	route := routers.NewUsersRouter(controller)
	router := rg.Group("/users")
	public := router.Group("/public")
	private := router.Group("/private")
	private.Use(jwtDeserializer)
	route.InitAuthRouter(public, private)
}

func (server *Server) Start(address string) error {
	log.Printf("Starting server on %s\n", address)
	return server.router.Run(address)
}

func RunGinServer(config config.Config, store db.Store, distributor worker.TaskDistributor) error {
	server, err := NewServer(config, store, distributor)
	if err != nil {
		return err
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		return err
	}
	return nil
}
