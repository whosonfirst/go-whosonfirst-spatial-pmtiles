# go-brooklynintegers-api

Go package for the Brooklyn Integers API.

## Documentation

Documentation is incomplete at this time.

## Usage

## Simple

```
package main

import (
	"fmt"
	
	"github.com/aaronland/go-brooklynintegers-api"
)

func main() {

	client := api.NewAPIClient()
	i, _ := client.CreateInteger()

	fmt.Println(i)
}
```

## Less simple

```
import (
       "fmt"
       
       "github.com/aaronland/go-brooklynintegers-api"
)

func main() {

	client := api.NewAPIClient()

	params := url.Values{}
	method := "brooklyn.integers.create"

	rsp, _ := client.ExecuteMethod(method, &params)
	i, _ := rsp.Int()

	fmt.Println(i)
}
```

## Tools

### brooklynt

Mint one or more Brooklyn Integers.

```
$> ./bin/brooklynt -h
Usage of ./bin/int:
  -count int
    	The number of Brooklyn Integers to mint (default 1)
```

## See also

* http://brooklynintegers.com/
* http://brooklynintegers.com/api
* https://github.com/aaronland/go-artisanal-integers
