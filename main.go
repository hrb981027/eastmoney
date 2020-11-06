package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"
)

const resultRegex = `jsonpgz\((.*?)\);`

var ctx = context.Background()

type Msg struct {
	Fundcode string `json:"fundcode"`
	Name  string `json:"name"`
	Gszzl string `json:"gszzl"`
	GzTime string `json:"gztime"`
}

var rdb *redis.Client

func main() {
	rdb = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB: 0,
	})

	router := gin.Default()

	router.Use(cors.Default())

	r := router.Group("/")

	{
		r.GET("/:id/:abbreviation", func(c *gin.Context) {
			id := c.Param("id")
			abbreviation := c.Param("abbreviation")

			result := getCache(id)

			if result != "" {
				c.String(http.StatusOK, result)

				return
			}

			msg := parse("http://fundgz.1234567.com.cn/js/" + id + ".js")

			if abbreviation != "" {
				msg.Name = abbreviation
			}

			if msg.Fundcode == "" {
				msg.Name = "Not Found"
				msg.Gszzl = "0"
			}

			result = format(msg)

			setCache(id, result)

			c.String(http.StatusOK, result)
		})
	}

	router.Run(":8000")
}

func getCache(id string) string {
	val, err := rdb.Get(ctx, id).Result()

	if err == redis.Nil {
		return ""
	}

	return val
}

func setCache(id string, context string) {
	t6 := (60 - time.Now().Second()) * 1e9

	rdb.Set(ctx, id, context, time.Duration(t6))
}

func parse(bUrl string) Msg {
	result := Msg{}

	client := http.Client{}

	request, err := http.NewRequest("GET", bUrl, nil)
	if err != nil {
		fmt.Println(err)

		return result
	}

	request.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	request.Header.Add("Accept-Charset", "UTF-8,*;q=0.5")
	request.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7")
	request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64; rv:60.0) Gecko/20100101 Firefox/60.0")
	request.Header.Add("referer", "http://fund.eastmoney.com")

	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)

		return result
	}

	defer response.Body.Close()

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)

		return result
	}

	s := reg(resultRegex, respBody)

	err = json.Unmarshal([]byte(s), &result)
	if err != nil {
		fmt.Println("Umarshal failed:", err)

		return result
	}

	return result
}

func format(msg Msg) string {
	return msg.Name +
		"：" + msg.Gszzl +
		"\n"
}

func reg(regexString string, content []byte) string {
	Reg := regexp.MustCompile(regexString)
	match := Reg.FindAllSubmatch(content, -1)

	for _, m := range match {
		return string(m[1])
	}

	return ""
}
