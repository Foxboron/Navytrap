package main

import (
	loader "github.com/Foxboron/Navytrap/loader"

	net "github.com/Foxboron/Navytrap/net"
	"gopkg.in/gin-gonic/gin.v1"
)

func Run(c *loader.Cmds) error {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
		net.Conn.WriteChannel("#test", "pong")
	})
	r.Run(":8080")
	return nil
}
