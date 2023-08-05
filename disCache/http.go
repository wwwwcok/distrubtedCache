package disCache

import (
	"distrubtedCache/disCache/consistenthash"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
)

type HTTPPool struct {
	self        string
	basePath    string
	mu          sync.Mutex
	peers       *consistenthash.Map
	httpGetters map[string]*httpGetter
}

const (
	defaultBasePath = "/_disCache"
	defaultReplicas = 50
)

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

//设置分布式的节点
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

//根据key选择节点
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true

	}
	return nil, false
}

var _PeerPicker = (*HTTPPool)(nil)

func (p *HTTPPool) Route(Engine *gin.Engine) {
	Engine.GET("/_disCache/:groupname/:key", p.handleDefault)
	Engine.GET("/api", p.handleDefault)

}

func (p *HTTPPool) handleDefault(ctx *gin.Context) {
	groupName := ctx.Param("groupname")
	key := ctx.Param("key")
	group := GetGroup(groupName)
	val, err := group.Get(key)
	if err != nil {
		fmt.Println(err)
	}
	ctx.Writer.Write(val.ByteSlice())
}

type httpGetter struct {
	baseURL string
}

func (h *httpGetter) Get(group string, key string) ([]byte, error) {

	u := fmt.Sprintf(
		"%v%v/%v", h.baseURL, url.QueryEscape(group),
		url.QueryEscape(key),
	)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}
	return bytes, nil
}
