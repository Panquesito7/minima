package minima

import (
	"context"
	"log"
	"net/http"
	"time"
)

/**
 * @info The framework structure
 * @property {*http.Server} [server] The net/http stock server
 * @property {bool} [started] Whether the server has started or not
 * @property {*time.Duration} [Timeout] The router's breathing time
 * @property {*Router} [router] The core router instance running with the server
 * @property {[]Handler} [minmiddleware] The standard Minima handler stack
 * @property {[]http.HandlerFunc} [rawmiddleware] The raw net/http Minima handler stack
 * @property {map[string]interface{}} [properties] The properties for the server instance
 * @property {*Config} [Config] The core config file for middlewares and router instances
 * @property {*time.Duration} [drain] The router's drain time
 */
type Minima struct {
	server        *http.Server
	started       bool
	Timeout       time.Duration
	router        *Router
	minmiddleware []Handler
	rawmiddleware []http.HandlerFunc
	properties    map[string]interface{}
	Config        *Config
	drain         time.Duration
}

/**
 * @info Make a new default Minima instance
 * @example `
func main() {
	app := Minima.New()

	app.Get("/", func(res *Minima.Response, req *Minima.Request) {
		res.Status(200).Send("Hello World")
	})

	app.Listen(":3000")
}
`
 * @returns {Minima}
*/
func New() *Minima {
	return &Minima{
		router: NewRouter(),
		Config: NewConfig(),
		drain:  0,
	}
}

/**
 * @info Starts the actual http server
 * @param {string} [addr] The port for the server instance to run on
 * @returns {error}
 */
func (m *Minima) Listen(addr string) error {
	if m.started {
		log.Panicf("Minimia's instance is already running at %s.", m.server.Addr)
	}
	m.server = &http.Server{Addr: addr, Handler: m}
	m.started = true

	return m.server.ListenAndServe()
}

/**
 * @info Injects the actual Minima server logic to http
 * @param {http.ResponseWriter} [w] The net/http response instance
 * @param {http.Request} [r] The net/http request instance
 * @returns {}
 */
func (m *Minima) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, params, match := m.router.routes[r.Method].Get(r.URL.Path)

	if match {
		if err := r.ParseForm(); err != nil {
			log.Printf("Error parsing form: %s", err)
			return
		}

		res := response(w, r, &m.properties)
		req := request(r)
		req.Params = params

		m.ServeMiddleware(res, req)
		f(res, req)
	} else {
		res := response(w, r, &m.properties)
		req := request(r)
		if m.router.notfound != nil {
			m.router.notfound(res, req)
		} else {
			w.Write([]byte("No matching route found"))
		}
	}
}

/**
 * @info Adds route with Get method
 * @param {string} [path] The route path
 * @param {...Handler} [handler] The handler for the given route
 * @returns {*Minima}
 */
func (m *Minima) Get(path string, handler Handler) *Minima {
	m.router.Get(path, handler)
	return m
}

/**
 * @info Adds route with Put method
 * @param {string} [path] The route path
 * @param {...Handler} [handler] The handler for the given route
 * @returns {*Minima}
 */
func (m *Minima) Put(path string, handler Handler) *Minima {
	m.router.Put(path, handler)
	return m
}

/**
 * @info Adds route with Options method
 * @param {string} [path] The route path
 * @param {...Handler} [handler] The handler for the given route
 * @returns {*Minima}
 */
func (m *Minima) Options(path string, handler Handler) *Minima {
	m.router.Options(path, handler)
	return m
}

/**
 * @info Adds route with Head method
 * @param {string} [path] The route path
 * @param {...Handler} [handler] The handler for the given route
 * @returns {*Minima}
 */
func (m *Minima) Head(path string, handler Handler) *Minima {
	m.router.Head(path, handler)
	return m
}

/**
 * @info Adds route with Delete method
 * @param {string} [path] The route path
 * @param {...Handler} [handler] The handler for the given route
 * @returns {*Minima}
 */
func (m *Minima) Delete(path string, handler Handler) *Minima {
	m.router.Delete(path, handler)
	return m
}

/**
 * @info Adds route with Patch method
 * @param {string} [path] The route path
 * @param {...Handler} [handler] The handler for the given route
 * @returns {*Minima}
 */
func (m *Minima) Patch(path string, handler Handler) *Minima {
	m.router.Patch(path, handler)
	return m
}

/**
 * @info Adds route with Post method
 * @param {string} [path] The route path
 * @param {...Handler} [handler] The handler for the given route
 * @returns {*Minima}
 */
func (m *Minima) Post(path string, handler Handler) *Minima {
	m.router.Post(path, handler)
	return m
}

/**
 * @info Injects the NotFound handler to the Minima instance
 * @param {Handler} [handler] Minima handler instance
 * @returns {*Minima}
 */
func (m *Minima) NotFound(handler Handler) *Minima {
	m.router.NotFound(handler)
	return m
}

/**
 * @info Injects the routes from the router to core stack
 * @param {*Router} [router] Minima router instance
 * @returns {*Minima}
 */
func (m *Minima) UseRouter(router *Router) *Minima {
	m.router.UseRouter(router)
	return m
}

/**
 * @info Mounts router to a specific path
 * @param {string} [path] The route path
 * @param {*Router} [router] Minima router instance
 * @returns {*Minima}
 */
func (m *Minima) Mount(path string, router *Router) *Minima {
	m.router.Mount(path, router)
	return m
}

/**
 * @info Injects middlewares and routers directly to core instance
 * @param {*Config} [config] The config instance
 * @returns {*Minima}
 */
func (m *Minima) UseConfig(config *Config) *Minima {
	m.minmiddleware = append(m.minmiddleware, config.Middleware...)
	m.rawmiddleware = append(m.rawmiddleware, config.HttpHandler...)
	for _, router := range config.Router {
		m.UseRouter(router)
	}
	return m
}

/**
 * @info The drain timeout for the core instance
 * @param {time.Duration} [time] The time period for drain
 * @returns {*Minima}
 */
func (m *Minima) ShutdownTimeout(t time.Duration) *Minima {
	m.drain = t
	return m
}

/**
 * @info Shutdowns the core instance
 * @param {context.Context} [ctx] The context for shutdown
 * @returns {error}
 */
func (m *Minima) Shutdown(ctx context.Context) error {
	log.Println("Stopping the server")
	return m.server.Shutdown(ctx)
}

/**
 * @info Declares prop for core properties
 * @param {string} [key] Key for the prop
 * @param {interface{}} [value] Value of the prop
 * @returns {*Minima}
 */
func (m *Minima) SetProp(key string, value interface{}) *Minima {
	m.properties[key] = value
	return m
}

/**
 * @info Gets prop from core properties
 * @param {string} [key] Key for the prop
 * @returns {interface{}}
 */
func (m *Minima) GetProp(key string) interface{} {
	return m.properties[key]
}

/**
 * @info Injects Minima middleware to the stack
 * @param {...Handler} [handler] The handler stack to append
 * @returns {}
 */
func (m *Minima) Use(handler ...Handler) {
	m.minmiddleware = append(m.minmiddleware, handler...)
}

/**
 * @info Injects net/http middleware to the stack
 * @param {...http.HandlerFunc} [handler] The handler stack to append
 * @returns {}
 */
func (m *Minima) UseRaw(handler ...http.HandlerFunc) {
	m.rawmiddleware = append(m.rawmiddleware, handler...)
}

/**
 * @info Serves and injects the middlewares to Minima logic
 * @param {Response} [res] The Minima response instance
 * @param {Request} [req] The Minima req instance
 * @returns {}
 */
func (m *Minima) ServeMiddleware(res *Response, req *Request) {
	if len(m.rawmiddleware) == 0 {
		return
	}
	for _, raw := range m.rawmiddleware {
		raw(res.ref, req.ref)
	}
	if len(m.minmiddleware) == 0 {
		return
	}
	for _, min := range m.minmiddleware {
		min(res, req)
	}
}
