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
var apiv1Version string = "0.1.625"

type User struct {
	LocalUserId string
}

func main() {
	fmt.Println("Starting...")

	ginRoot := gin.Default()
	keeper, keeperInit_err := dao.DefaultRedis("localhost:6379", "", 1)
	if keeperInit_err != nil {
		fmt.Print("Error Accured while init SessionManager,Maybe You need startup Redis First.")
		return
	}

	ginRoot.Use(session.DefaultGinSessionManager(keeper, "localhost"))

	ginRoot.GET("/ping", func(c *gin.Context) {
		c.String(200, "Pong!")
	})
	ginRootApi := ginRoot.Group("/api/v1")
	{
		ginRootApi.GET("/status", func(c *gin.Context) {
			c.String(200, ApiStatus)
		})
		ginRootApi.GET("/version", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"result": 0, "version": apiv1Version})
		})
		ginRootApi.Group("/acfun-helper/options/")
		{
			ginRootApi.POST("/download", func(c *gin.Context) {
				rawmsg := c.PostForm("authCookie")
				msg, err := jsonquery.Parse(strings.NewReader(rawmsg))

				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"result": 400, "info": "Invalid auth data format."})
				}

				acCookies := jsonquery.FindOne(msg, "AcCookies")
				Cookies := acCookies.InnerText()
				acPassToken := jsonquery.FindOne(msg, "AcPassToken")
				PassToken := acPassToken.InnerText()
				authInfo := Cookies + "; acPasstoken=" + PassToken

				userId := jsonquery.FindOne(msg, "LocalUserId").InnerText()

				fmt.Print(authInfo, userId)

				if s, senssoinModStatus := session.GetSession(c); senssoinModStatus {
					if _, ok := s.Get("LocalUserId"); ok {
						var sessionData User
						if err := s.GetStruct("user", &sessionData); err == nil {
							result, status := dataGet(userId)
							fmt.Print(s)
							if status {
								c.String(http.StatusOK, result)
								fmt.Println("use session")
							}
						}
					} else {
						if userAuth(authInfo) {
							result, status := dataGet(userId)
							if status {
								c.String(http.StatusOK, result)
								fmt.Println("after userAuth")
							}
						} else {
							c.JSON(500, gin.H{"result": 500, "info": "You should login to acfun.cn first."})
						}
					}
				} else {
					c.JSON(500, gin.H{"result": 500, "info": "Server Fault."})
				}

			})
			ginRootApi.POST("/upload", func(c *gin.Context) {
				rawmsg := c.PostForm("options_data")

				msg, err := jsonquery.Parse(strings.NewReader(rawmsg))

				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"result": 400, "info": "Invalid auth data format."})
				}

				acCookies := jsonquery.FindOne(msg, "AcCookies")
				Cookies := acCookies.InnerText()
				acPassToken := jsonquery.FindOne(msg, "AcPassToken")
				PassToken := acPassToken.InnerText()
				authInfo := Cookies + "; acPasstoken=" + PassToken

				userId := jsonquery.FindOne(msg, "LocalUserId").InnerText()

				if s, senssoinModStatus := session.GetSession(c); senssoinModStatus {
					//1判断用户是否存于session中
					if _, ok := s.Get("LocalUserId"); ok {
						var sessionData User
						if err := s.GetStruct("user", &sessionData); err == nil {
							status := dataSet(userId, rawmsg)
							if status {
								c.JSON(http.StatusOK, gin.H{"result": 0, "info": "Success Sync and We have your senssoin Info."})
							}
						}
					} else {
						//2用户UID不存在session中则用一般的验证流程并写入session
						if userAuth(authInfo) {
							status := dataSet(userId, rawmsg)
							if status {
								s.Set("LocalUserId", userId)
								_ = s.SetStruct("user", User{LocalUserId: userId})
								c.JSON(http.StatusOK, gin.H{"result": 0, "info": "Success Sync."})
							}
						} else {
							c.JSON(500, gin.H{"result": 500, "info": "You should login to acfun.cn first."})
						}
					}
				} else {
					c.JSON(500, gin.H{"result": 500, "info": "Server Fault."})
				}
			})
		}
	}

	ginRoot.Run("localhost:5000")
}

func userAuth(auhtInfo string) bool {
	var authUrl string = "https://www.acfun.cn/rest/pc-direct/user/personalBasicInfo"
	resultRaw := localGet(authUrl, auhtInfo)
	fmt.Print(resultRaw)
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
