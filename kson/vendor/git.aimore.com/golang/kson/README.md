
# 使用示例
```
package main

import (
	"fmt"
	"github.com/werbenhu/go-tools/kson"
)

var data = []byte(`{
	"version":1,
	"person": {
	  "name": {
		"first": "Leonid",
		"last": "Bugaev",
		"fullName": "Leonid Bugaev"
	  },
	  "avatars": [
		{ "url": "xxxx", "type": "thumbnail", "age":1 },
		{ "url": "bbb", "type": "dddd", "age":2 },
		{ "url": "bbb", "type": "dddd", "age":3 }
	  ]
	},
	"company": {
	  "name": "Acme"
	}
  }`)

type Name struct {
	First    string `json:"first"`
	Last     string `json:"last"`
	FullName string `json:"fullName"`
}

func testParser() {
	root, err := kson.ParseBytes(data)
	if err != nil {
		fmt.Printf("err:%s\n", err)
	}

	exist1 := root.Exists("person.avatars")
	fmt.Printf("exist1:%t\n", exist1)
	//exist1:true

	exist2 := root.Exists("aaa.xxx")
	fmt.Printf("exist2:%t\n", exist2)
	//exist2:false

	exist3 := root.Get("aaa.exist3")
	if exist3 == nil {
		fmt.Printf("exist3 not exist\n")
		//exist3 not exist
	}

	url := root.Get("person.avatars.2.url")
	if url != nil {
		fmt.Printf("url:%s\n", url.GetString(""))
		//url:bbb
	}

	company := root.Get("company")
	if url != nil {
		fmt.Printf("company:%s\n", company.String())
		//company:{"name":"Acme"}
	}

	avatars := root.GetArray("person.avatars")
	fmt.Printf("size:%d\n", len(avatars))
	//size:7

	age := root.GetInt("person.avatars.2.age")
	fmt.Printf("age:%d\n", age)
	//age:3

	first := root.GetString("person.name.first")
	fmt.Printf("first:%s\n", first)
	//first:Leonid

	name := root.Get("person.name")
	fmt.Printf("name:%s\n", string(name.Bytes()))
	//name:{"first":"Leonid","last":"Bugaev","fullName":"Leonid Bugaev"}

	var nameObject Name
	kson.Unmarshal(name.Bytes(), &nameObject)
	fmt.Printf("nameObject:%+v\n", nameObject)
	//nameObject:{First:Leonid Last:Bugaev FullName:Leonid Bugaev}

	bs, _ := kson.Marshal(nameObject)
	fmt.Printf("bs:%s\n", string(bs))
	//bs:{"first":"Leonid","last":"Bugaev","fullName":"Leonid Bugaev"}
}

func main() {
	testParser()
}

```