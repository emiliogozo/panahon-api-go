package api

import (
	"fmt"

	db "github.com/emiliogozo/panahon-api-go/db/sqlc"
	"github.com/emiliogozo/panahon-api-go/token"
	"github.com/emiliogozo/panahon-api-go/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	config     util.Config
	router     *gin.Engine
	store      db.Store
	tokenMaker token.Maker
}

// NewServer creates a new HTTP server and setup routing
func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("mobile_number", validMobileNumber)
	}

	server.setupRouter()

	return server, nil
}

func (s *Server) setupRouter() {
	r := gin.Default()

	api := r.Group(s.config.APIBasePath)

	users := api.Group("/users")
	{
		users.POST("/login", s.LoginUser)
		users.POST("/register", s.RegisterUser)

		usersAuth := addMiddleware(users,
			authMiddleware(s.tokenMaker),
			roleMiddleware("SUPERADMIN"))
		usersAuth.GET("", s.ListUsers)
		usersAuth.GET(":id", s.GetUser)
		usersAuth.POST("", s.CreateUser)
		usersAuth.PUT(":id", s.UpdateUser)
		usersAuth.DELETE(":id", s.DeleteUser)
	}

	api.POST("/tokens/renew", s.RenewAccessToken)

	roles := api.Group("/roles")
	{
		rolesAuth := addMiddleware(roles,
			authMiddleware(s.tokenMaker),
			roleMiddleware("SUPERADMIN"))
		rolesAuth.GET("", s.ListRoles)
		rolesAuth.GET(":id", s.GetRole)
		rolesAuth.POST("", s.CreateRole)
		rolesAuth.PUT(":id", s.UpdateRole)
		rolesAuth.DELETE(":id", s.DeleteRole)
	}

	stations := api.Group("/stations")
	{
		stations.GET("", s.ListStations)
		stations.GET(":station_id", s.GetStation)

		stnObservations := stations.Group(":station_id/observations")
		{
			stnObservations.GET("", s.ListStationObservations)
			stnObservations.GET(":id", s.GetStationObservation)
		}

		stationsAuth := addMiddleware(stations,
			authMiddleware(s.tokenMaker),
			roleMiddleware("ADMIN"))
		stationsAuth.POST("", s.CreateStation)
		stationsAuth.PUT(":station_id", s.UpdateStation)
		stationsAuth.DELETE(":station_id", s.DeleteStation)

		stnObservationsAuth := addMiddleware(stnObservations,
			authMiddleware(s.tokenMaker),
			roleMiddleware("ADMIN"))
		{
			stnObservationsAuth.POST("", s.CreateStationObservation)
			stnObservationsAuth.PUT(":id", s.UpdateStationObservation)
			stnObservationsAuth.DELETE(":id", s.DeleteStationObservation)
		}

	}

	glabs := api.Group("/glabs")
	{
		glabs.GET("/optin", s.GLabsOptIn)
		glabs.POST("/load", s.CreateGLabsLoad)
	}

	sm := api.Group("/sm")
	{
		sm.POST("", s.CreateLufftObservationHealth)
	}

	lufft := api.Group("/lufft")
	{
		lufft.GET(":station_id/logs", s.LufftMsgLog)
	}

	api.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	s.router = r
}

func (s *Server) Start(address string) error {
	return s.router.Run(address)
}

func addMiddleware(r *gin.RouterGroup, m ...gin.HandlerFunc) gin.IRoutes {
	if gin.Mode() != gin.TestMode {
		return r.Use(m...)
	}

	return r
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
