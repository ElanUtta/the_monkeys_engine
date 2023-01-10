package main

import (
	"github.com/89minutes/the_new_project/api_gateway/config"
	"github.com/89minutes/the_new_project/api_gateway/pkg/auth"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatalf("failed to load the config: %v", err)
	}

	router := gin.Default()

	auth.RegisterRouter(router, &cfg)
	logrus.Info("starting the gateway server at port: ", cfg.Port)
	if err = router.Run(cfg.Port); err != nil {
		logrus.Fatalf("failed to start the gateway: %v", err)
	}
}