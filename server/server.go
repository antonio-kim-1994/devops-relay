package main

import (
	"fmt"
	"github.com/antonio-kim-1994/devops-relay/server/config"
	"github.com/antonio-kim-1994/devops-relay/server/handler"
	"github.com/antonio-kim-1994/devops-relay/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {
	if os.Getenv("APP_ENV") == "" {
		log.Fatal().Msg("No APP_ENV environment variable served.")
	}

	// config 설정
	cfg := config.Setting()

	g := gin.Default()

	// Route 등록
	registerMainRoutes(g)

	err := g.Run(fmt.Sprintf(":%s", cfg.ServerPort))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to run DevOps Relay Gateway.")
	}
}

func registerMainRoutes(g *gin.Engine) {
	// 500 error 혹은 panic으로 서버 shutdown 시 재기동
	g.Use(gin.Recovery())
	common := g.Group("/healthz")
	{
		common.GET("/healthcheck", handler.CommonHealthCheck)
	}

	update := g.Group("/update")
	{
		update.Use(middleware.ValidateApiRequest())
		update.POST("/github", handler.HandleGithubRequest)
		update.POST("/slack", handler.HandleSlackResponse)
	}

	sys := g.Group("/sys")
	{
		sys.Use(middleware.ValidateApiRequest())
		sys.POST("/healthcheck", handler.ServerHealthCheck)
	}
}
