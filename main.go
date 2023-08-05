package main

import (
	"distrubtedCache/disCache"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	gee := createGroup()
	if api {
		go startAPIServer(apiAddr, gee)
	}
	startCacheServer(addrMap[port], []string(addrs), gee)
}

func createGroup() *disCache.Group {
	return disCache.NewGroup("scores", 2<<10, disCache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, gee *disCache.Group) {

	peers := disCache.NewHTTPPool(addr)
	peers.Set(addrs...)
	gee.RegisterPeers(peers)
	ginSrv := gin.Default()
	ginSrv.GET("/api", func(ctx *gin.Context) {
		key := ctx.Query("key")
		view, err := gee.Get(key)
		if err != nil {
			ctx.Status(404)
			return
		}
		ctx.Writer.Write(view.ByteSlice())
	})
	log.Println("disCache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], ginSrv))
}

func startAPIServer(apiAddr string, gee *disCache.Group) {
	disCache.NewGroup("scores", 2<<10, disCache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
	ginSrv := gin.Default()
	//	addr := "localhost:9999"
	peers := disCache.NewHTTPPool(apiAddr)
	peers.Route(ginSrv)
	log.Println("disCache is running at", apiAddr)
	log.Fatal(ginSrv.Run(apiAddr))
}
