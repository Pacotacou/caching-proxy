package proxy

import (
	"bytes"
	"caching-proxy/cache"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ProxyServer struct {
	Port       int
	OriginURL  string
	Cache      *cache.Cache
	httpClient *http.Client
}

func NewProxyServer(port int, originURL string) *ProxyServer {
	return &ProxyServer{
		Port:      port,
		OriginURL: originURL,
		Cache:     cache.NewCache(),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *ProxyServer) Start() error {
	log.Printf("Starting the caching proxy server on port %d, forwarding to %s\n", p.Port, p.OriginURL)
	log.Printf("Cache is currently empty\n")

	http.HandleFunc("/admin/cache", p.handleCacheAdmin)

	http.HandleFunc("/", p.handleRequest)

	return http.ListenAndServe(fmt.Sprintf(":%d", p.Port), nil)
}

func (p *ProxyServer) ClearCache() {
	p.Cache.Clear()
	log.Println("Cache cleared")
}

func (p *ProxyServer) handleCacheAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		p.Cache.Clear()
		cacheSize := p.Cache.Size()
		log.Printf("Cache cleared via admin endpoint")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Cache cleared. Current size: %d items", cacheSize)
		return
	}
}

func (p *ProxyServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	originURL, err := url.Parse(p.OriginURL)
	if err != nil {
		http.Error(w, "Invalid origin URL", http.StatusInternalServerError)
		return
	}

	cacheKey := cache.CreateCacheKey(r)

	if cachedResp, found := p.Cache.Get(cacheKey); found {
		log.Printf("Cache HIT for %s %s", r.Method, r.URL.Path)
		cachedResp.WriteToResponse(w)
		return
	}

	log.Printf("Cache MISS for %s %s, forwarding to origin", r.Method, r.URL.Path)
	targetURL := &url.URL{
		Scheme:   originURL.Scheme,
		Host:     originURL.Host,
		Path:     r.URL.Path,
		RawQuery: r.URL.RawQuery,
	}

	proxyReq, err := http.NewRequest(r.Method, targetURL.String(), nil)
	if err != nil {
		http.Error(w, "Error creating a proxy request", http.StatusInternalServerError)
		return
	}

	if r.Body != nil {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading the request body", http.StatusInternalServerError)
			return
		}
		proxyReq.Body = io.NopCloser(bytes.NewReader(body))
		r.Body.Close()
	}

	for header, values := range r.Header {
		if strings.ToLower(header) != "connection" {
			for _, value := range values {
				proxyReq.Header.Add(header, value)
			}
		}
	}

	if clientIP, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		if prior, ok := proxyReq.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		proxyReq.Header.Set("X-Forwarded-For", clientIP)
	}

	resp, err := p.httpClient.Do(proxyReq)
	if err != nil {
		http.Error(w, "Error forwarding request to origin: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	cachedResp, err := cache.CopyResponse(resp)
	if err != nil {
		http.Error(w, "Error caching response", http.StatusInternalServerError)
		return
	}

	p.Cache.Set(cacheKey, cachedResp.StatusCode, cachedResp.Headers, cachedResp.Body)
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.Header().Set("X-Cache", "MISS")

	w.WriteHeader(resp.StatusCode)

	io.Copy(w, resp.Body)
}
