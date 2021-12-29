### Forward Your Request With One Line Of Code

```go
package main

import (
	"github.com/nfangxu/proxy"
	"log"
	"net/http"
)

func main() {
	log.Fatal(http.ListenAndServe(":7788", http.HandlerFunc(proxy.NewSimpleProxy("https://foo.com").Proxy)))
}
```
