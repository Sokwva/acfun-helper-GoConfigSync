package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/antchfx/jsonquery"
	"github.com/gin-gonic/gin"
	session "github.com/loop-xxx/gin-session"
	"github.com/loop-xxx/gin-session/dao"
)

var ApiStatus string = "ok"
var apiv1Version string = "0.1.415"

func main() {
	fmt.Println("Starting...")

	ginRoot := gin.Default()
	keeper, keeperInit_err := dao.DefaultRedis("localhost:6379", "", 1)
	if keeperInit_err != nil {
		fmt.Print("Error Accured while init GinSessionManager.")
		return
	}

	ginRoot.Use(session.DefaultGinSessionManager(keeper, "localhost"))

	ginRoot.GET("/", func(c *gin.Context) {
		c.String(200, "Hi there!")
	})
	ginRoot.GET("/ping", func(c *gin.Context) {
		c.String(200, "Pong!")
	})
	ginRootApi := ginRoot.Group("/api/v1")
	{
		ginRootApi.GET("/status", func(c *gin.Context) {
			c.String(200, ApiStatus)
		})
		ginRootApi.GET("/versionDetail", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"result": 0, "version": apiv1Version})
		})
		ginRootApi.POST("/acfun-helper/options/upload", func(c *gin.Context) {
			rawmsg := c.PostForm("options_data")
			fmt.Print(rawmsg)

			msg, err := jsonquery.Parse(strings.NewReader(rawmsg))
			fmt.Print(msg)

			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"result": 400, "info": "Invalid auth data format."})
			}

			acCookies := jsonquery.FindOne(msg, "AcCookies")
			Cookies := acCookies.InnerText()
			acPassToken := jsonquery.FindOne(msg, "AcPassToken")
			PassToken := acPassToken.InnerText()
			authInfo := Cookies + "; acPasstoken=" + PassToken
			userId := jsonquery.FindOne(msg, "LocalUserId").InnerText()

			if _, err := session.GetSession(c); err {
				status := dataSet(userId, rawmsg)
				if status {
					c.JSON(http.StatusOK, gin.H{"result": 0, "info": "Success Sync and We have your senssoin Info."})
				}
			} else {
				if userAuth(authInfo) {
					status := dataSet(userId, rawmsg)
					if status {
						c.JSON(http.StatusOK, gin.H{"result": 0, "info": "Success Sync."})
					}
				} else {
					c.JSON(500, gin.H{"result": 500, "info": "You should login to acfun.cn first."})
				}
			}

		})
	}

	ginRoot.Run()
}

func userAuth(auhtInfo string) bool {
	var authUrl string = "https://api-new.app.acfun.cn/rest/app/user/hasSignedIn"
	resultRaw := localGet(authUrl, auhtInfo)
	result, err := jsonquery.Parse(strings.NewReader(resultRaw))
	if err != nil {
		return false
	}
	status := jsonquery.FindOne(result, "result").InnerText()
	if status == "0" {
		return true
	} else {
		return false
	}
}

func localGet(url string, cookies string) string {
	client := &http.Client{Timeout: 3 * time.Second}
	resRaw, _ := http.NewRequest("GET", url, nil)
	resRaw.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.85 Safari/537.36 Edg/90.0.818.46")
	resRaw.Header.Add("Cookie", cookies)
	// res, err := client.Get(url)
	res, err := client.Do(resRaw)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	var buffer [512]byte
	result := bytes.NewBuffer(nil)
	for {
		n, err := res.Body.Read(buffer[0:])
		result.Write(buffer[0:n])
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
	}
	return result.String()
}
