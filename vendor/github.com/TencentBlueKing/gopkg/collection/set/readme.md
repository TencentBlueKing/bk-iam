# collection/set

`set`底层使用`map`实现, 封装了set类操作

目前支持类型:

- string
- int64

注意, 目前未使用泛型, 所以没有统一的interface


## Usage

### String Set

```go
import "github.com/TencentBlueKing/gopkg/collection/set"

s := set.NewStringSet()
// s := set.NewStringSetWithValues([]string{"hello", "world"})
// s := set.NewFixedLengthStringSet(2)
// s := set.SplitStringToSet("a,b,c", ",")
s.Add("abc")
s.Has("abc")
s.Append([]string{"abc", "def"}...)
s.Size()
sli1 := s.ToSlice()
s1 := s.ToString(",")
```

### Int64 Set

```go
import "github.com/TencentBlueKing/gopkg/collection/set"

s := set.NewInt64Set()
// s := set.NewInt64SetWithValues([]int64{123, 456})
// s :=  set.NewFixedLengthInt64Set(2)
s.Add(123)
s.Has(123)
s.Append([]int64{123, 456}...)
s.Size()
sli1 := s.ToSlice()
```
