package runtime

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var Router *gin.Engine

// METHODGET constant
const METHODGET uint8 = 0

// METHODPOST constant
const METHODPOST uint8 = 1

// METHODPUT constant
const METHODPUT uint8 = 3

// METHODDELETE constant
const METHODDELETE uint8 = 4

// METHODPATCH constant
const METHODPATCH uint8 = 5

// Route is the information for every URI.
type Route struct {
	// Name is the name of this Route.
	Name string
	// Method is the string for the HTTP method. ex) GET, POST etc..
	Method uint8
	// Pattern is the pattern of the URI.
	Pattern string
	// HandlerFunc is the handler function of this route.
	HandlerFunc gin.HandlerFunc
}

// Routes type
type Routes []Route

// NewRouter build and return a new Router
func NewRouter(routes []Route, path string, engine *gin.Engine) *gin.RouterGroup {

	group := engine.Group(path)

	for _, route := range routes {
		switch route.Method {
		case METHODGET:
			group.GET(route.Pattern, route.HandlerFunc)
		case METHODPOST:
			group.POST(route.Pattern, route.HandlerFunc)
		case METHODPUT:
			group.PUT(route.Pattern, route.HandlerFunc)
		case METHODPATCH:
			group.PATCH(route.Pattern, route.HandlerFunc)
		case METHODDELETE:
			group.DELETE(route.Pattern, route.HandlerFunc)
		}
	}
	return group

}

// InitRouter initialize all the routers of the application
func InitRouter(secure bool, production bool) error {

	if production {
		gin.SetMode(gin.ReleaseMode)
	}

	if secure {
		// Secure mod
	}

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "OPTIONS", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "User-Agent", "Referrer", "Host", "Token", "X-Requested-With", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowAllOrigins:  true,
		MaxAge:           86400,
	}))

	var dataPlaneRouter = Routes{
		{
			"pdu",
			METHODPOST,
			"/pdu",
			HandlePDU,
		},
	}

	var admissionControlRouter = Routes{
		{
			"adm",
			METHODPOST,
			"/adm",
			HandleAdmissionControl,
		},
	}
	
	var deleteConnection = Routes{
		{
			"deleteConnection",
			METHODDELETE,
			"/cutoff",
			HandleDeleteConnection,
		},
	}

	NewRouter(dataPlaneRouter, "/data-plane", router)
	NewRouter(admissionControlRouter, "/control-plane", router)
	NewRouter(deleteConnection, "/data-plane", router)

	Router = router

	return nil
}
