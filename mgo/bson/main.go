package main

import (
	"fmt"

	"gopkg.in/mgo.v2/bson"
)

type (
	Foo struct {
		F string      `bson:"f"`
		B interface{} `bson:"b"`
	}

	Bar struct {
		B string `bson:"b"`
	}

	Buzz struct {
		B string `bson:"b"`
	}
)

var _ bson.Setter

func main() {
	f := Foo{F: "f", B: Bar{"bar"}}

	bs, err := bson.Marshal(f)
	if err != nil {
		panic(err)
	}
	var ff Foo
	if err := bson.Unmarshal(bs, &ff); err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", ff)
}
