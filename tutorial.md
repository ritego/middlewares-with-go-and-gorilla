# RiteGo - Middlewares with Go and Gorilla

The full source code for this tutorial is available at [Middleware With Go and Gorilla](https://github.com/ritego/middlewares-with-go-and-gorilla).

## Introduction
In this tutorial, we would learn how to:
- Build HTTP middleware that intercepts requests
- Build HTTP middleware that intercepts response
- Integrate middleware with Gorilla Mux

## Middlewares
Middlewares are logic that run between a client request and a server controller. It usually has access to both the request and response object. With middleware, we can intercept and modify request/response to/from a server application.
 
HTTP servers are meant to receive and process request from clients - web apps, mobile apps and other servers. In most cases, there are specific activities common to every route or a group of routes. Middleware help us to implement a set of logic common to one or many of these routes in a single place.

Using middleware improves our code base by providing separation of concern and it's also a nice way to avoid repeating ourselves.

## Middlewares in Go
In Go, a middleware is a function that accepts and return a [http.Handler](https://pkg.go.dev/net/http#Handler).
A [http.Handler](https://pkg.go.dev/net/http#Handler) is an interface that defines a single method `ServeHTTP`. This method is called and expected to provide the appropriate response to a given request.

```go
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}
```

## Request middlewares
We would start by creating a simple interface that prints every request to the console.

First, we would create a `struct` type that would extend and implements the above handler interface. Our custom type `logger` inherits/extends [http.Handler](https://pkg.go.dev/net/http#Handler) by composition. This is to enable our implementation have access to the parent `ServeHTTP` method.
```go
type logger struct {
	h http.Handler
}
```

Then we write our implementation of the `ServeHTTP` method of the handler interface. Essential, what we are doing at this point is to override the parent method with our own custom logic. Since we just want to log the request URL, we print it as shown `fmt.Println(r.URL.String())`
```go
func (l logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Do something to the request,
	// Something as simple as logging it
	fmt.Println(r.URL.String())

	l.h.ServeHTTP(w, r)

	// Do something after the response
	// We would come back to this later
}
```

Next, we craft our middleware function. If you remember, a middleware function accepts and returns a [http.Handler](https://pkg.go.dev/net/http#Handler).
```go
func LogRequest (h http.Handler) http.Handler {
	return logger{h}
}
```

Just as required, our middle accepts and returns a [http.Handler](https://pkg.go.dev/net/http#Handler). Recall that `logger` satisfies the  [http.Handler](https://pkg.go.dev/net/http#Handler) interface by implementing the `ServeHTTP` method.

Finally, to consume our middleware, we can wire it into a router like [Gorilla Mux](https://github.com/gorilla/mux) in this way:
```go
router = mux.NewRouter()
router.Use(
	middlewares.LogRequest,
)
```

## A less powerful, simple approach
There is an easier approach to implementing middleware for our application, all the above can be achieved in fewer lines of code. But the long journey is necessary to understand whats happening in between the lines.
```go
func LogRequest (next http.Handler) http.Handler { 
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.String())
		next.ServeHTTP(w, r)
	})
}
```

The (HandlerFunc)[https://pkg.go.dev/net/http#HandlerFunc] (Not to be confused with [HandleFunc](https://pkg.go.dev/net/http#HandleFunc)), does must of the job for us. This is the recommended way for lightweight middlewares. Just like our `logger` middleware, it satisfies the requirement of [http.Handler](https://pkg.go.dev/net/http#Handler) interface by implementing the `ServeHTTP` method.
```go
type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
	f(w, r)
}
```

This default implementation just propagates the request/response to the provided callback `f`. The only difference between this and our logger is the `type`. This is of type `func` while our logger is a struct.


### A more powerful, functional approach 
Let's take a step further, by creating a middleware that takes in parameters. What we want to achieve is something like this:
```go
router = mux.NewRouter()
router.Use(
	middlewares.LogRequest(os.Stdout), // notice the param: os.Stdout
)
```

We do this by turning our middleware into a closure. The parent function returns a function, that function  in turn returns the actual middleware. 
```go
func LogRequest(l io.Writer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l.Write(r.URL.String())
			next.ServeHTTP(w, r)
		})
	}
}
```

## Complete example of our request middleware
Additionally, this snippet write the log to the provided writer. It also captures more than just the URL.
```go
func LogRequest(l io.Writer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := ioutil.ReadAll(r.Body)
			request, _ := json.Marshal(struct {
				Host   string
				URL    string
				Method string
				Header http.Header
				Status int
				Body   []byte
				Type   string
			}{
				Host:   r.Host,
				URL:    r.URL.String(),
				Method: r.Method,
				Header: r.Header,
				Body:   body,
				Type:   ResponseType,
			})
			l.Write(request)
			next.ServeHTTP(w, r)
		})
	}
}
```

## Response middlewares
We would turn our attention now to a middleware that logs HTTP response before sending it to the client.

First, we would extend `http.ResponseWriter` by composition. The additional `Status` and `Body` parameters are for temporary holding of the response data so we can consume it sometime in the future.

```go
type LogResponseWriter struct {
	http.ResponseWriter
	Status int
	Body   []byte
}
```

Then we override the `Write` and `WriteHeader` methods of `http.ResponseWriter`
```go
func (l *LogResponseWriter) Write(b []byte) (int, error) {
	l.Body = b
	return l.ResponseWriter.Write(b)
}

func (l *LogResponseWriter) WriteHeader(status int) {
	l.Status = status
	l.ResponseWriter.WriteHeader(status)
}
```

Next, we craft our middleware.
```go
func LogResponse(l io.Writer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := &LogResponseWriter{
				ResponseWriter: w,
			}

			next.ServeHTTP(w, r)

			response, _ := json.Marshal(struct {
				Host   string
				URL    string
				Method string
				Header http.Header
				Status int
				Body   []byte
				Type   string
			}{
				Host:   r.Host,
				URL:    r.URL.String(),
				Method: r.Method,
				Header: l.Header(),
				Status: l.Status,
				Body:   l.Body,
				Type:   ResponseType,
			})

			l.Write(response)
		})
	}
}
```

Finally, To consume our middleware, we can wire it into a router like [Gorilla Mux](https://github.com/gorilla/mux) in this way:
```go
router = mux.NewRouter()
router.Use(
	middlewares.LogResponse(),
)
```

## Conclusion
The full source code for a running program is available at [Middleware With Go and Gorilla](https://github.com/ritego/middlewares-with-go-and-gorilla).
