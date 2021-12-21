# cache

cache the data in memory, auto to fetch the data if missing.

- based on via [patrickmn/go-cache](https://github.com/patrickmn/go-cache)
- retrieveFunc will be called if the key not in cache
- TTL required
- cache the missing key for 5s, avoid `cache penetration`
- use [singleflight ](https://godoc.org/golang.org/x/sync/singleflight) to avoid `cache breakdown`
- cache in memory, no need to worry about `cache avalanche`
- support go version 1.11+

## usage

#### use string key

```go
package main

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/cache/memory"
)

// 1. impl the reterive func
func RetrieveOK(k cache.Key) (interface{}, error) {
	arg := k.(memory.StringKey)
	fmt.Println("arg: ", arg)
	// you can use the arg to fetch data from database or http request
	// username, err := GetFromDatabase(arg)
	// if err != nil {
	//     return nil, err
	// }
	return "ok", nil
}

func main() {
	// 2. new a cache
	c := memory.NewCache(
		"example",
		false,
		RetrieveOK,
		5*time.Minute,
		nil)

	// 4. use it
	k := memory.NewStringKey("hello")

	data, err := c.Get(k)
	fmt.Println("err == nil: ", err == nil)
	fmt.Println("data from cache: ", data)
}
```

#### use your own key


```go
package main

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/cache/memory"
)

// 1. impl the key interface, Key() string
type ExampleKey struct {
	Field1 string
	Field2 int64
}

func (k ExampleKey) Key() string {
	return fmt.Sprintf("%s:%d", k.Field1, k.Field2)
}

// 2. impl the reterive func
func RetrieveExample(inKey cache.Key) (interface{}, error) {
	k := inKey.(ExampleKey)
	fmt.Println("ExampleKey Field1 and Field2 value:", k.Field1, k.Field2)
	// data, err := GetFromDatabase(k.Field1, k.Field2)
	// if err != nil {
	//     return nil, err
	// }
	return "world", nil
}

func main() {
	// 3. new a cache
	c := memory.NewCache(
		"example",
		false,
		RetrieveExample,
		5*time.Minute,
		nil)

	// 4. use it
	k := ExampleKey{
		Field1: "hello",
		Field2: 42,
	}

	data, err := c.Get(k)
	fmt.Println("err == nil: ", err == nil)
	fmt.Println("data from cache: ", data)

	dataStr, err := c.GetString(k)
	fmt.Println("err == nil: ", err == nil)
	fmt.Printf("data type is %T, value is %s\n", dataStr, dataStr)
}
```

#### add a random expired duration

打散过期时间, 防止并发较高的场景下, 同时失效发起retrieve导致后端压力过大


```go
func newRandomDuration(seconds int) backend.RandomExpirationDurationFunc {
	return func() time.Duration {
		return time.Duration(rand.Intn(seconds*1000)) * time.Millisecond
	}
}

LocalSubjectCache = memory.NewCache(
    "local_subject",
    disabled,
    retrieveSubject,
    1*time.Minute,
    newRandomDuration(30),
)
```
