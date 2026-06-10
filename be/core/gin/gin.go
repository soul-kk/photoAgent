package gin

import (
	"log"

	"go-service-starter/api/middleware"
	"go-service-starter/api/router"
	"go-service-starter/config"
	"go-service-starter/core/imageprep"

	"github.com/gin-gonic/gin"
)

func GinInit() *gin.Engine {
	r := gin.Default()
	cfg := config.MustLoad()
	edge, qual := cfg.Image.MaxLongEdge, cfg.Image.JPEGQuality
	if edge <= 0 {
		edge = imageprep.DefaultMaxLongEdge
	}
	if qual <= 0 {
		qual = imageprep.DefaultJPEGQuality
	}
	imageprep.ApplyConfig(edge, qual)
	log.Printf("imageprep: max_long_edge=%d jpeg_quality=%d", edge, qual)
	router.GenerateRouter(r)
	middleware.InitSecret(cfg.JWT.Secret)
	return r
}
