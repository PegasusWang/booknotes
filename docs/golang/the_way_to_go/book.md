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

```go
package sort

type Sorter interface {
	Len() int
	Less(i, jint) bool
	Swap(i, j int)
}

fucn Sort(data Sorter) {
	for pass := 1; pass <data.Len(); pass ++ {
		for i :=0; i < data.Len()-pass; i++ {
			if data.Less(i+1,i) {
				data.Swap(i,i+1)
			}
		}
	}
}

func IsSorted(data Sorter) bool {
	n := data.Len()
	for i := n-1; i >0; i-- {
		if data.Less(i, i-1){
			return false
		}
	}
	return true
}

// convenience types for common cases

type IntArray []int
func (p IntArray) Len() int {return len(p)}
func (p IntArray) Less(i,j int) bool {return p[i] < p[j]}
func (p IntArray) Swap(i,j int) {p[i],p[j] = p[j],p[i]}
```


### 11.9.2 contsucting an array of a general type or with variables of different types

```go
type Element interface{}

type Vector struct {
	a []Element
}

func (p *Vector) At(i int) Element {
	return p.a[i]
}

func (p *Vector) Set(i int, Element e) {
	p.a[i] = e
}
```

tree struct: 

```go
package main

import "fmt"

type Node struct {
	le   *Node
	data interface{}
	ri   *Node
}

func NewNode(left, right *Node) *Node {
	return &Node{left, nil, right}
}

func (n *Node) SetData(data interface{}) {
	n.data = data
}

func main() {
	root := NewNode(nil, nil)
	root.SetData("root node")
	a := NewNode(nil, nil)
	a.SetData("left node")
	b := NewNode(nil, nil)
	b.SetData("right node")
	root.le = a
	root.ri = b
	fmt.Printf("%v\n", root)
}
```


## 11.10 The reflect package


