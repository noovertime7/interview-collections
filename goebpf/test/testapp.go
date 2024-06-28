package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

var (
	db  *gorm.DB
	rdb *redis.Client
)

type User struct {
	ID   int    `gorm:"primaryKey"`
	Name string `gorm:"column:name"`
}

func init() {
	var err error
	//初始化数据库连接
	dsn := "root:123456@tcp(125.124.136.63:32727)/test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	//初始化redis连接
	rdb = redis.NewClient(&redis.Options{
		Addr:     "125.124.136.63:32530",
		Password: "Zhangyuge233", // no password set
		DB:       0,              // use default DB
	})
}

func myHandler(r *http.Request) string {
	if r.URL.Query().Get("name") != "" {
		return "Hello " + r.URL.Query().Get("name") + "!"
	} else {
		return "Hello all!"
	}
}

func getUser(id string) string {
	intId, err := strconv.Atoi(id)
	if err != nil {
		return "id must be int"

	}
	user := &User{ID: intId}
	db.First(user)
	return user.Name
}

func main() {
	port := flag.String("port", "8080", "port to serve on")
	flag.Parse()
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(myHandler(request)))
	})
	//根据mysql查询用户
	http.HandleFunc("/user", func(writer http.ResponseWriter, request *http.Request) {
		id := request.URL.Query().Get("id")
		if id != "" {
			writer.Write([]byte(getUser(id)))
		} else {
			writer.Write([]byte("id is required"))
		}
	})
	//redis set
	http.HandleFunc("/set", func(writer http.ResponseWriter, request *http.Request) {
		key := request.URL.Query().Get("key")
		value := request.URL.Query().Get("value")
		if key != "" && value != "" {
			rdb.Set(context.Background(), key, value, 0)
			writer.Write([]byte("set success"))
		} else {
			writer.Write([]byte("key and value are required"))
		}
	})
	//redis get
	http.HandleFunc("/get", func(writer http.ResponseWriter, request *http.Request) {
		key := request.URL.Query().Get("key")
		if key != "" {
			val, err := rdb.Get(context.Background(), key).Result()
			if err != nil {
				writer.Write([]byte("key not found"))
			} else {
				writer.Write([]byte(val))
			}
		} else {
			writer.Write([]byte("key is required"))
		}
	})

	fmt.Println("Serving on port", *port)
	http.ListenAndServe(":"+*port, nil)
}
