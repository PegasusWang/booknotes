# 11 Interfaces and reflection

An Interface define a set of methods. And cannot contain variables.

```go
type Namer interface {
	Method1(param_list) return_type
	Method2(param_list) return_type
}
```

## 11.3 type assertion

```go
if v, ok := varI.(T); ok {
//
}
```

Value-received methods can be called with pointer values because they can be dereference first.
