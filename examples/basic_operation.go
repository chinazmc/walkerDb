package main

import (
	"fmt"
	"walkerDb/database"
)

func main() {
	opts := database.DefaultOptions
	opts.DirPath = "/tmp/bitcask-go"
	db, err := database.Open(opts)
	if err != nil {
		panic(err)
	}

	err = db.Put([]byte("name"), []byte("bitcask"))
	if err != nil {
		panic(err)
	}
	val, err := db.Get([]byte("name"))
	if err != nil {
		panic(err)
	}
	fmt.Println("val = ", string(val))

	err = db.Delete([]byte("name"))
	if err != nil {
		panic(err)
	}
}
