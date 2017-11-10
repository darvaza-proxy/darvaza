# gpp

a sample http proxy handler write in golang.

support http/https proxy and also act as a normal http server.

support http2 if you run https.


##Usage

Use gpp.Handler as a normal http.Handler.

gpp.Handler will automatically detects the local request and proxy request, it handles the proxy request itself and invoke the http.DefaultServeMux to handle local path request.

you can use the `http.Handle` or `http.HandleFunc` to register the local path request handler.

##Example
```go
package main

import (
	. "fmt"
	"github.com/fangdingjun/gpp"
	"log"
	"net/http"
)

func main() {
	port := 8080

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("<h1>welcome!</h1>"))
	})

	log.Print("Listen on: ", Sprintf("0.0.0.0:%d", port))
	err := http.ListenAndServe(Sprintf(":%d", port), &gpp.Handler{EnableProxy:true})
	if err != nil {
		log.Fatal(err)
	}
}

```

run this example

you can use `curl` to test server and proxy server function

local path request
```bash
curl http://127.0.0.1:8080/
```

proxy request
```bash
curl --proxy http://127.0.0.1:8080/ http://httpbin.org/ip
```

see more examples on `samples/` directory.

