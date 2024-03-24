package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

//**
// Код представляет собой реализацию HTTP-сервера с использования фреймворка gin
// (скрипт можно выпилить, если работать чисто по API?)
//

// Message resp struct
type Message struct {
	Status  int         `json:"status"`
	Payload interface{} `json:"payload"`
}

// Это основная функция, которая настраивает и запускает HTTP-сервер. Она определяет маршруты для различных HTTP-запросов и их обработчики.
func HTTPAPIServer() {
	//Set HTTP API mode
	log.WithFields(logrus.Fields{
		"module": "http_server",
		"func":   "RTSPServer",
		"call":   "Start",
	}).Infoln("Server HTTP start")
	var public *gin.Engine
	if !Storage.ServerHTTPDebug() {
		// по сути, возращает engine-экземпляр http-объекта
		gin.SetMode(gin.ReleaseMode)
		public = gin.New()
	} else {
		gin.SetMode(gin.DebugMode)
		public = gin.Default()
	}

	//Это middleware, который добавляет заголовки CORS ко всем ответам сервера, чтобы разрешить доступ к ресурсам из разных источников.
	public.Use(CrossOrigin())
	//Add private login password protect methods
	privat := public.Group("/")
	if Storage.ServerHTTPLogin() != "" && Storage.ServerHTTPPassword() != "" {
		privat.Use(gin.BasicAuth(gin.Accounts{Storage.ServerHTTPLogin(): Storage.ServerHTTPPassword()}))
	}

	/*
		Static HTML Files Demo Mode
	*/
	if Storage.ServerHTTPDemo() {
		public.LoadHTMLGlob(Storage.ServerHTTPDir() + "/templates/*")
		public.GET("/", HTTPAPIServerIndex)
		public.GET("/pages/player/webrtc/:uuid/:channel", HTTPAPIPlayWebrtc)

		// изначально это было в privat, там просится аутентификация
		public.GET("/streams", func(c *gin.Context) {
			// Эта функция возвращает список потоков данных из Storage
			list, err := Storage.MarshalledStreamsList()
			if err != nil {
				c.IndentedJSON(500, Message{Status: 0, Payload: err.Error()})
				return
			}
			c.IndentedJSON(200, Message{Status: 1, Payload: list})
		}) 

		public.StaticFS("/static", http.Dir(Storage.ServerHTTPDir()+"/static"))
	}

	/*
		Stream Control elements
	*/

	// privat.GET("/streams", HTTPAPIServerStreams)

	err := public.Run(Storage.ServerHTTPPort())
	if err != nil {
		log.WithFields(logrus.Fields{
			"module": "http_router",
			"func":   "HTTPAPIServer",
			"call":   "ServerHTTPPort",
		}).Fatalln(err.Error())
		os.Exit(1)
	}

}

// Обрабатывают запросы и отдают html-страницы
func HTTPAPIServerIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"port":    Storage.ServerHTTPPort(),
		"streams": Storage.Streams,
		"version": time.Now().String(),
		"page":    "index",
	})

}

func HTTPAPIPlayWebrtc(c *gin.Context) {
	c.HTML(http.StatusOK, "play_webrtc.tmpl", gin.H{
		"port":    Storage.ServerHTTPPort(),
		"streams": Storage.Streams,
		"version": time.Now().String(),
		"page":    "play_webrtc",
		"uuid":    c.Param("uuid"),
		"channel": c.Param("channel"),
	})
}

// CrossOrigin Access-Control-Allow-Origin any methods
func CrossOrigin() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
