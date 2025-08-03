package main

import (
	"fmt"
	"github.com/antonio-kim-1994/devops-relay/gateway/config"
	"github.com/antonio-kim-1994/devops-relay/gateway/handler"
	"github.com/antonio-kim-1994/devops-relay/gateway/middleware"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func main() {
	// config 설정
	cfg := config.Setting()

	// Gin 모드 설정
	//if os.Getenv("GIN_MODE") != "debug" {
	//	gin.SetMode(gin.ReleaseMode)
	//}

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
	g.GET("/healthz/healthcheck", handler.GatewayHealthCheck)

	v1 := g.Group("/v2")
	{
		github := v1.Group("/github")
		github.Use(middleware.ValidateApiRequest())
		github.POST("/update", handler.GithubRequestHandler)

		slack := v1.Group("/slack")
		slack.Use(middleware.ValidationCheckSlackPayload())
		slack.POST("/deploy", handler.SlackResponseHandler)
	}

	sys := g.Group("/sys")
	{
		sys.Use(middleware.ValidateApiRequest())
		sys.POST("/healthcheck", handler.ServerHealthCheck)
	}
}
