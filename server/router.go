package server

import (
	"AynaAPI/config"
	_ "AynaAPI/docs"
	"AynaAPI/server/api/v1/anime"
	"AynaAPI/server/api/v1/auth"
	"AynaAPI/server/api/v1/general"
	"AynaAPI/server/api/v1/novel"
	"AynaAPI/server/api/v1/upload"
	"AynaAPI/server/fs"
	"AynaAPI/server/middleware/jwt"
	"AynaAPI/server/middleware/perm"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(config.ServerConfig.GinMode)
}

// InitRouter initialize routing information
func InitRouter() *gin.Engine {
	var engine *gin.Engine = gin.New()
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	corsconfig := cors.DefaultConfig()
	corsconfig.AllowOrigins = []string{"*"}

	engine.Use(cors.New(corsconfig))

	staticDirFs := engine.Group("/_static")
	{
		staticDirFs.Use(jwt.AuthUser(), perm.CheckPermission(128))
		staticDirFs.StaticFS("/", gin.Dir(config.ServerConfig.FileRoot, true))
	}

	staticFs := engine.Group("/static")
	{
		staticFs.Static("/file", config.ServerConfig.GetFilePath("file"))
		staticFs.Static(fs.GetUploadUrl(), fs.GetUploadPath())
	}

	apiV1 := engine.Group("/api/v1")
	{
		authApi := apiV1.Group("/auth")
		{
			authApi.GET("/login", auth.Login)
			authApi.GET("/info", jwt.GetUserOnly(), auth.GetInfo)
		}
		uploadApi := apiV1.Group("/upload")
		{
			uploadApi.Use(jwt.AuthUser())
			uploadApi.POST("/bilipic", upload.UploadBiliPic)
			uploadApi.POST("/image", upload.UploadImage)
		}
		generalApi := apiV1.Group("/general")
		{
			generalApi.GET("/bypasscors", general.BypassCors)
			generalApi.GET("/teamsplit", general.GetRandomSeparation)
		}
		animeApi := apiV1.Group("/anime")
		{
			animeApi.GET("/plist", anime.GetProviderList)
			animeApi.GET("/providerlist", anime.GetProviderList)

			animeApi.GET("/search/:provider", anime.Search)
			animeApi.GET("/playurl/:provider", anime.GetPlayUrl)
			animeApi.GET("/info/:provider", anime.GetInfo)
			animeApi.GET("/resolve/:provider", anime.Resolve)

			animeApi.GET("/search", anime.SearchAll)
			animeApi.GET("/playurl", anime.GetPlayUrlAll)
			animeApi.GET("/info", anime.InfoAll)
			animeApi.GET("/resolve", anime.ResolveAll)
		}
		novelApi := apiV1.Group("/novel")
		{
			novelApi.GET("/plist", novel.GetProviderList)
			novelApi.GET("/providerlist", novel.GetProviderList)
			novelApi.GET("/rlist", novel.GetProviderRules)
			novelApi.GET("/rulelist", novel.GetProviderRules)

			novelApi.GET("/info", novel.GetInfo)
			novelApi.GET("/content", novel.GetContent)

			novelApi.GET("/search/:provider", novel.Search)
			novelApi.GET("/search", novel.SearchAll)
		}
	}
	return engine
}
