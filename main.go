package main

import (
  "fmt"
  "log"
  "net/http"
  "github.com/PuerkitoBio/goquery"
  "encoding/json"
  "time"
  "math/rand"
  "github.com/aofei/air"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
  "strings"
  "github.com/aofei/mimesniffer"
  "encoding/base64"
  "github.com/aofei/sandid"
  "github.com/aliyun/aliyun-oss-go-sdk/oss"
  "os"
  "bytes"
  "strconv"
  "math"
  "net/url"
)

var (
  a = air.Default
  errorhtml string
)

type NewsMessage struct {
  PostTime string `json:"get_time"`
  NewsList []NewsLists `json:"news"`
}

type NewsLists struct {
  NewsName string `json:"newsname"`
  NewsUrl   string  `json:"newsurl"`
  Newstime string `json:time`
  Author string `json:author`
  Body string `json:content`
  Imageurl string `json:imageurl`
}

type waterflow_info_struct struct {
  PostTime string `json:"get_time"`
  Waterflow_List []Flows `json:"waterflow_info"`
}

type Flows struct {
  Uid int `json:"uid"`
  Name   string  `json:"name"`
  Order   string  `json:"order"`
  Family   string  `json:"family"`
  Genus   string  `json:"genus"`
}

type Message_board_struct struct {
  PostTime string `json:"get_time"`
  Board []Message `json:"Message"`
}

type Message struct {
  Uid int `json:"uid"`
  Name string `json:"name"`
  Comment string `json:"comment"`
  Mail string `json:"address"`
  SendTime string `json:"sendtime"`
}

type waterflow_detail_struct struct {
  PostTime string `json:"get_time"`
  Waterflow_Detail []Flows_detail `json:"waterflow_detail"`
}

type Flows_detail struct {
  Uid int `json:"uid"`
  Name   string  `json:"name"`
  Latin_name   string  `json:"latin_name"`
  Introduce   string  `json:"introduce"`
  Imgurl   string  `json:"imgurl"`
}

type Test_result struct {
  Ie float64 `json:"内倾/外倾"`
  Sn float64 `json:"感觉/直觉"`
  Tf float64 `json:"思考/情感"`
  Pj float64 `json:"知觉/判断"`
  Kind string `json:"类型"`
  Introduce string `json:"描述"`
  Similar_name string `json:"相似水鸟"`
  Similar string `json:"相似度"`
  Similar_imageurl string `json:"该水鸟图片地址"`
}

type Map map[string]interface{}

func main() {
  a.DebugMode = true
  a.Address = ":8080"
  a.FILE("/error", "templates/error.html") //var errorhtml string
  a.FILE("/building", "templates/build.html") //var errorhtml string
  a.FILE("/test", "templates/test.html")
  a.ErrorHandler = func(err error, req *air.Request, res *air.Response) {
    if res.ContentLength > 0 {
      return
    }
    res.Redirect("/error")
  }

  a.BATCH(
    []string{http.MethodGet, http.MethodHead},
    "/",
    func(req *air.Request, res *air.Response) error {
      return res.Redirect("/building")
    },
  )
  
  a.GET("/getnews", Scrape)
  a.GET("/info", get_waterflow_info)
  a.GET("/detail", get_waterflow_detail)
  a.GET("/getcomments", get_comments)
  a.GET("/s",search)
  a.POST("/api/v1/comments",commentsHandler)
  a.GET("/api/v1/dailypush",dailypushHandler)
  a.POST("/api/v1/test",testHandler)
  a.Serve()
}

func Scrape(req *air.Request, respon *air.Response) error {
  /*
    功能：获取新闻
    URL：/getnews
    方式：GET
    参数：无
    返回格式：json
  */
  //获取当前时间戳，声明变量
  now_time := time.Now().Unix()
  s1 := NewsMessage{}
  var json_file NewsMessage

  db, err := sql.Open("mysql", "root:chapus1215@tcp(cdb-6qucl950.bj.tencentcdb.com:10102)/waterflow_alpha?charset=utf8")
  checkErr(err)

  var get_time int64
  err = db.QueryRow("SELECT gettime FROM waterflow_news where id=(select max(id) from waterflow_news)").Scan(&get_time)
  if err == sql.ErrNoRows || now_time-get_time >= 604800{ //如果返回为空或者时间差大于7天，执行爬取函数
    
    var page_num int //爬取网站页码
    var news_num int //新闻个数
    page_num = 1
    news_num = 0
    
    json_file.PostTime = time.Now().Format("2006-01-02")
    for ;news_num <= 19;{
      page_number:= strconv.Itoa(page_num)
      // Request the HTML page.
      res, err := http.Get("http://search.shidi.org/default.aspx?keyword=水鸟&&page=" + page_number)
      defer res.Body.Close()
      if res.StatusCode != 200 {
        log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
      }

      // Load the HTML document
      doc, err := goquery.NewDocumentFromReader(res.Body)
      if err != nil {
        log.Fatal(err)
      }

      // Find the review items and Restore with Json

      doc.Find("ul").Find("li").EachWithBreak(func(i int, s *goquery.Selection) bool {
        // For each item found, get the band and title
        band := s.Find("a").Text()
        href := s.Find(".siteurl").Text()

        href = strings.Replace(href, " ", "", -1)
        href = strings.Replace(href, "\n", "", -1)
        href = strings.Replace(href, "\r", "", -1)
        href = strings.Replace(href, "\t", "", -1)

        //
        inhref, err := http.Get(href)
        if err != nil {
          log.Fatal(err)
        }

        defer inhref.Body.Close()
        if inhref.StatusCode != 200 {
          log.Fatalf("status code error: %d %s", inhref.StatusCode, inhref.Status)
        }

        // Load the HTML document
        indoc, err := goquery.NewDocumentFromReader(inhref.Body)
        if err != nil {
          log.Fatal(err)
        }

        author := indoc.Find(".arcTitle").Find("strong").Text()
        newstime := indoc.Find(".arcTitle").Find(".arcTime").Text()
        //newstime = strings.Replace(newstime, " ", "", -1)
        newstime = strings.Replace(newstime, "\n", "", -1)
        newstime = strings.TrimSpace(newstime)
        //imageurl := indoc.Find(".arcTitle").Find(".arcTime").Text()
        var str string
        var 属性 string
        var imageurl string
        var imagesrc string

        indoc.Find("#endText").Find("p").Each(func(i_tmp int, s_tmp *goquery.Selection) {
            str = str + s_tmp.Text() + "\\n"
            str = strings.Replace(str, "\\n\\n", "\\n", -1)
        })
        imagesrc, _ = indoc.Find("#endText").Find("img").Attr("src")
        imagesrc = strings.TrimSpace(imagesrc)
        name := sandid.New().String()

        if strings.HasPrefix(imagesrc, "data:image/"){

          imagesrc = strings.TrimPrefix(imagesrc,"data:image/jpeg;base64,")
          
          b, err := base64.StdEncoding.DecodeString(imagesrc)
          if err != nil {
            panic(err)
          }
          属性 = mimesniffer.Sniff(b)
          fmt.Println(属性)
          
          // 打开云存储
          client, err := oss.New("oss-cn-beijing.aliyuncs.com", "LTAIt4GfhTk7x4r5", "PI7lKD9hkAjK42c68tzPIZatUZ5Zc8")
          if err != nil {
            fmt.Println("Error:", err)
            os.Exit(-1)
          }

          // 获取存储空间。
          bucket, err := client.Bucket("waterflow-scrapy")
          if err != nil {
            fmt.Println("Error:", err)
            os.Exit(-1)
          }

          // 上传本地文件。
          err = bucket.PutObject(name, bytes.NewReader(b), oss.ContentType(属性))
          if err != nil {
            fmt.Println("Error:", err)
            os.Exit(-1)
          }
          
          imageurl = "https://waterflow-scrapy.oss-cn-beijing.aliyuncs.com/" + name
          
        } else if(imagesrc!=""){
          
          //如果是url，则直接下载保存到云端
          res, err := http.Get(imagesrc)
          if err != nil{
            fmt.Println(imagesrc)
            panic(err)
          }
          defer res.Body.Close()

          client, err := oss.New("oss-cn-beijing.aliyuncs.com", "LTAIt4GfhTk7x4r5", "PI7lKD9hkAjK42c68tzPIZatUZ5Zc8")
          if err != nil {
            fmt.Println("Error:", err)
            os.Exit(-1)
          }

          // 获取存储空间。
          bucket, err := client.Bucket("waterflow-scrapy")
          if err != nil {
            fmt.Println("Error:", err)
            os.Exit(-1)
          }

          // 上传本地文件。
          err = bucket.PutObject(name, res.Body, oss.ContentType(属性))
          if err != nil {
            fmt.Println("Error:", err)
            os.Exit(-1)
          }
          
          imageurl = "https://waterflow-scrapy.oss-cn-beijing.aliyuncs.com/" + name
        }

        if imageurl != "" {
          stmt, err := db.Prepare("INSERT waterflow_news SET newsname=?,newsurl=?,newstime=?,author=?,body=?,imageurl=?,gettime=?")
            checkErr(err)

          _, err = stmt.Exec(band, href, newstime, author, str, imageurl, now_time)
            checkErr(err)

          json_file.NewsList = append(json_file.NewsList, NewsLists{NewsName: band,NewsUrl: href,Newstime: newstime, Author:author, Body:str, Imageurl:imageurl})
          news_num++
        }

        if news_num >= 20 {
          return false
        }
        return true
      })
      page_num++
    }
  } else {
    //直接查询数据库
    json_file.PostTime = time.Now().Format("2006-01-02")
    rows, err := db.Query("SELECT * FROM waterflow_news where gettime=?",get_time)
    checkErr(err)

    for rows.Next() {
      var uid int
      var name string
      var newsurl string
      var newstime string
      var author string
      var body string
      var imageurl string
      var none_time int
      err = rows.Scan(&uid, &name, &newsurl, &newstime, &author, &body, &imageurl, &none_time)
      checkErr(err)
      json_file.NewsList = append(json_file.NewsList, NewsLists{NewsName: name,NewsUrl: newsurl,Newstime: newstime, Author:author, Body:body, Imageurl:imageurl})
    }
  }

  db.Close()
  json_fin, err := json.Marshal(json_file)
  if err != nil {
    fmt.Println("json err:", err)
  }

  json.Unmarshal(json_fin, &s1)
  respon.Header.Set("Content-Type", "application/json; charset=utf-8")
  return respon.WriteJSON(s1)
}

func commentsHandler(req *air.Request, res *air.Response) error {
  /*
    评论功能
    URL：/api/v1/comments
    方式：post
    参数：
      name（非空，string，用户名字）
      content（非空，string，评论内容）
      mail（非空，string，联系方式）
    返回：json
  */

  //获取参数，转换为string
  user_name_A := req.Param("name").Value()
  content_A := req.Param("content").Value()
  mail_address_A := req.Param("mail").Value()

  name := user_name_A.String()
  content := content_A.String()
  mail_address := mail_address_A.String()
  send_time := time.Now().Format("2006-01-02 15:04:05") 

  db, err := sql.Open("mysql", "root:chapus1215@tcp(cdb-6qucl950.bj.tencentcdb.com:10102)/waterflow_alpha?charset=utf8")
  checkErr(err)

  //存入数据库

  stmt, err := db.Prepare("INSERT message_board SET user_name=?,comment=?,address=?,time=?")
    checkErr(err)

  re, err := stmt.Exec(name, content, mail_address, send_time)
    checkErr(err)

  id, err := re.LastInsertId()
    checkErr(err)

  fmt.Println(id)

  db.Close()

  return Success(res, "")
}

func dailypushHandler(req *air.Request, res *air.Response) error {
  /*
    随机推送功能
    URL：/api/v1/dailypush
    方式：post
    参数：无
    返回：json
  */
  rand.Seed(time.Now().UnixNano())//添加种子
  i := rand.Intn(39)
  j := rand.Intn(39)

  for ;j==i; {
    j = rand.Intn(39)
  }

  s := waterflow_detail_struct{}
  var json_file waterflow_detail_struct
  json_file.PostTime = time.Now().Format("2006-01-02")

  db, err := sql.Open("mysql", "root:chapus1215@tcp(cdb-6qucl950.bj.tencentcdb.com:10102)/waterflow_alpha?charset=utf8")
  checkErr(err)
  //查询数据
  rows, err := db.Query("SELECT * FROM waterflow_detail where id=? or id=?", i+1, j+1)
  checkErr(err)

  for rows.Next() {
    var uid int
    var name string
    var latin_name string
    var introduce string
    var imgurl string
    err = rows.Scan(&uid, &name, &latin_name, &introduce, &imgurl)
    checkErr(err)
    json_file.Waterflow_Detail = append(json_file.Waterflow_Detail, Flows_detail{Uid: uid,Name: name,Latin_name: latin_name,Introduce: introduce, Imgurl: imgurl})
  }
  db.Close()
  json_fin, err := json.Marshal(json_file)
  if err != nil {
    fmt.Println("json err:", err)
  }
  json.Unmarshal(json_fin, &s)
  res.Header.Set("Content-Type", "application/json; charset=utf-8")
  return res.WriteJSON(s)
}

func testHandler(req *air.Request, res *air.Response) error {
  /*
    URL：/api/v1/test
    方式：post
    返回：json
    详情进入页面查看（/test）
  */
  var ans [31]string
  ans[1] = req.Param("Quest1").Value().String()
  ans[2] = req.Param("Quest2").Value().String()
  ans[3] = req.Param("Quest3").Value().String()
  ans[4] = req.Param("Quest4").Value().String()
  ans[5] = req.Param("Quest5").Value().String()
  ans[6] = req.Param("Quest6").Value().String()
  ans[7] = req.Param("Quest7").Value().String()
  ans[8] = req.Param("Quest8").Value().String()
  ans[9] = req.Param("Quest9").Value().String()
  ans[10] = req.Param("Quest10").Value().String()
  ans[11] = req.Param("Quest11").Value().String()
  ans[12] = req.Param("Quest12").Value().String()
  ans[13] = req.Param("Quest13").Value().String()
  ans[14] = req.Param("Quest14").Value().String()
  ans[15] = req.Param("Quest15").Value().String()
  ans[16] = req.Param("Quest16").Value().String()
  ans[17] = req.Param("Quest17").Value().String()
  ans[18] = req.Param("Quest18").Value().String()
  ans[19] = req.Param("Quest19").Value().String()
  ans[20] = req.Param("Quest20").Value().String()
  ans[21] = req.Param("Quest21").Value().String()
  ans[22] = req.Param("Quest22").Value().String()
  ans[23] = req.Param("Quest23").Value().String()
  ans[24] = req.Param("Quest24").Value().String()
  ans[25] = req.Param("Quest25").Value().String()
  ans[26] = req.Param("Quest26").Value().String()
  ans[27] = req.Param("Quest27").Value().String()
  ans[28] = req.Param("Quest28").Value().String()
  ans[29] = req.Param("Quest29").Value().String()
  ans[30] = req.Param("Quest30").Value().String()

  /*
  “外倾/内倾”=（内倾-外倾）/21*10 （正分为内倾I， 负分为外倾E）
  “感觉/直觉”=（感觉-直觉）/26*10（正分为感觉S，负分为直觉N）
  “思考/情感”=（思考-情感）/24*10（正分为思考T，负分为情感F）
  “知觉/判断”=（知觉-判断）/22*10（正分为感性P，负分为判断J）
  */
  var I, E float64 //内倾，外倾
  var S, N float64 //感觉，直觉
  var T, F float64 //思考，情感
  var P, J float64 //感性，判断
  I = 0
  E = 0
  S = 0
  N = 0
  T = 0
  F = 0
  P = 0
  J = 0
  var i int
  //fmt.Println(ans1,ans2,ans3,ans4,ans5,ans6,ans7,ans8,ans9,ans10)
  for i=1; i<=30; i++ {
    if(i == 1||i == 4||i==13||i==27){
      if ans[i] == "A" {
        t := 1.0
        t = RandonFloat(t)
        J += t
      } else if ans[i]  == "B"{
        t := 1.0
        t = RandonFloat(t)
        P += t
      }
    }

    if(i == 10||i == 16||i==23||i==25){
      if ans[i] == "A" {
        t := 1.0
        t = RandonFloat(t)
        P += t
      } else if ans[i] == "B"{
        t := 1.0
        t = RandonFloat(t)
        J += t
      }
    }

    if(i == 20||i == 28){
      if ans[i] == "A" {
        t := 1.0
        t = RandonFloat(t)
        T += t
      } else if ans[i] == "B"{
        t := 1.0
        t = RandonFloat(t)
        F += t
      }
    }

    if(i == 9||i == 29){
      if ans[i] == "A" {
        t := 1.0
        t = RandonFloat(t)
        F += t
      } else if ans[i] == "B"{
        t := 1.0
        t = RandonFloat(t)
        T += t
      }
    }

    if(i == 2||i == 17||i == 30){
      if ans[i] == "A" {
        t := 1.0
        t = RandonFloat(t)
        S += t
      } else if ans[i] == "B"{
        t := 1.0
        t = RandonFloat(t)
        N += t
      }
    }

    if(i == 5||i == 8||i == 11||i == 14|| i == 24){
      if ans[i] == "A" {
        t := 1.0
        t = RandonFloat(t)
        N += t
      } else if ans[i] == "B"{
        t := 1.0
        t = RandonFloat(t)
        S += t
      }
    }

    if(i == 3||i == 6||i == 7||i == 12|| i == 15|| i == 19 || i == 22){
      if ans[i] == "A" {
        t := 1.0
        t = RandonFloat(t)
        E += t
      } else if ans[i] == "B"{
        t := 1.0
        t = RandonFloat(t)
        I += t
      }
    }

    if(i == 18||i == 21||i == 26){
      if ans[i] == "A" {
        t := 1.0
        t = RandonFloat(t)
        I += t
      } else if ans[i] == "B"{
        t := 1.0
        t = RandonFloat(t)
        E += t
      }
    }
  }

  IE := Round2((I-E)/10*10)
  SN := Round2((S-N)/8*10)
  TF := Round2((T-F)/4*10)
  PJ := Round2((P-J)/8*10)

  var kind string
  var introduce string
  if(IE >= 0 && SN >= 0 && TF >= 0 && PJ >= 0){
    kind = "ISTP（探险水鸟）"
    introduce = "你是一只灵活、忍耐力强，是个安静的观察者水鸟咕咕，一旦有问题发生，就会马上行动，找到实用的解决方法。分析事物运作的原理，能从大量的信息中很快的找到关键的症结所在。对于原因和结果感兴趣，用逻辑的方式处理问题，重视效率。"
  }
  if(IE >= 0 && SN >= 0 && TF >= 0 && PJ < 0){
    kind = "ISTJ（检查员水鸟）"
    introduce = "你是一只安静、严肃，通过全面性和可靠性获得成功的水鸟咕咕。实际，有责任感。决定有逻辑性，并一步步地朝着目标前进，不易分心。喜欢将工作、家庭和生活都安排得井井有条。重视传统和忠诚。"
  }
  if(IE >= 0 && SN >= 0 && TF < 0 && PJ >= 0){
    kind = "ISFP（艺术水鸟）"
    introduce = "你是一只安静、友好、敏感、和善的水鸟咕咕。享受当前，喜欢有自己的空间，喜欢能按照自己的时间表工作。对于自己的价值观和自己觉得重要的水鸟非常忠诚，有责任心。不喜欢争论和冲突。不会将自己的观念和价值观强加到别的水鸟身上。"
  }
  if(IE >= 0 && SN >= 0 && TF < 0 && PJ < 0){
    kind = "ISFJ（保姆水鸟）"
    introduce = "你是一只安静、友好、有责任感和良知的水鸟咕咕。坚定地致力于完成他们的义务。全面、勤勉、精确，忠诚、体贴，留心和记得他们重视的水鸟的小细节，关心他们的感受。努力把工作和家庭环境营造得有序而温馨。 "
  }
  if(IE >= 0 && SN < 0 && TF >= 0 && PJ >= 0){
    kind = "INTP（学术水鸟）"
    introduce = "你是一只对于自己感兴趣的任何事物都寻求找到合理的解释的水鸟咕咕。喜欢理论性的和抽象的事物，热衷于思考而非社交活动。安静、内向、灵活、适应力强。对于自己感兴趣的领域有超凡的集中精力深度解决问题的能力。多疑，有时会有点挑剔，喜欢分析。"
  }
  if(IE >= 0 && SN < 0 && TF >= 0 && PJ < 0){
    kind = "INTJ（专家水鸟）"
    introduce = "你是一只在实现自己的想法和达成自己的目标时有创新的想法和非凡的动力的水鸟咕咕。能很快洞察到外界事物间的规律并形成长期的远景计划。一旦决定做一件事就会开始规划并直到完成为止。多疑、独立，对于自己和其他水鸟能力和表现的要求都非常高。"
  }
  if(IE >= 0 && SN < 0 && TF < 0 && PJ >= 0){
    kind = "INFP（哲学水鸟）"
    introduce = "你是一只理想主义，对于自己的价值观和自己觉得重要的水鸟非常忠诚的水鸟咕咕。希望外部的生活和自己内心的价值观是统一的。好奇心重，很快能看到事情的可能性，能成为实现想法的催化剂。寻求理解别的水鸟和帮助他们实现潜能。适应力强，灵活，善于接受，除非是有悖于自己的价值观的。"
  }
  if(IE >= 0 && SN < 0 && TF < 0 && PJ < 0){
    kind = "INFJ（博爱水鸟）"
    introduce = "你是一只寻求思想、关系、物质等之间的意义和联系的水鸟咕咕。希望了解什么能够激励水鸟，对水鸟有很强的洞察力。有责任心，坚持自己的价值观。对于怎样更好的服务大众有清晰的远景。在对于目标的实现过程中有计划而且果断坚定。"
  }
  if(IE < 0 && SN >= 0 && TF >= 0 && PJ > 0){
    kind = "ESTP（挑战者水鸟）"
    introduce = "你是一只灵活、忍耐力强，实际，注重结果的水鸟咕咕。觉得理论和抽象的解释非常无趣。喜欢积极地采取行动解决问题。注重当前，自然不做作，享受和其他水鸟在一起的时刻。喜欢物质享受和时尚。学习新事物最有效的方式是通过亲身感受和练习。"
  }
  if(IE < 0 && SN >= 0 && TF >= 0 && PJ < 0){
    kind = "ESTJ（管家水鸟）"
    introduce = "你是一只实际、现实主义的水鸟咕咕。果断，一旦下决心就会马上行动。善于将项目和水鸟组织起来将事情完成，并尽可能用最有效率的方法得到结果。注重日常的细节。有一套非常清晰的逻辑标准，有系统性地遵循，并希望其他水鸟也同样遵循。在实施计划时强而有力。"
  }
  if(IE < 0 && SN >= 0 && TF < 0 && PJ >= 0){
    kind = "ESFP（表演者水鸟）"
    introduce = "你是一只外向、友好、接受力强的水鸟咕咕。热爱生活、水鸟和物质上的享受。喜欢和其他水鸟一起将事情做成功。在工作中讲究常识和实用性，并使工作显得有趣。灵活、自然不做作，对于新的任何事物都能很快地适应。学习新事物最有效的方式是和其他水鸟一起尝试。"
  }
  if(IE < 0 && SN >= 0 && TF < 0 && PJ < 0){
    kind = "ESFJ（主人水鸟）"
    introduce = "你是一只热心肠、有责任心、合作的水鸟咕咕。希望周边的环境温馨而和谐，并为此果断地执行。喜欢和其他水鸟一起精确并及时地完成任务。事无巨细都会保持忠诚。能体察到其他水鸟在日常生活中的所需并竭尽全力帮助。希望自己和自己的所为能受到其他水鸟的认可和赏识。"
  }
  if(IE < 0 && SN < 0 && TF >= 0 && PJ >= 0){
    kind = "ENTP（智多星水鸟）"
    introduce = "你是一只反应快、睿智，有激励其他水鸟的能力，警觉性强、直言不讳的水鸟咕咕。在解决新的、具有挑战性的问题时机智而有策略。善于找出理论上的可能性，然后再用战略的眼光分析。善于理解别的水鸟。不喜欢例行公事，很少会用相同的方法做相同的事情，倾向于一个接一个的发展新的爱好。"
  }
  if(IE < 0 && SN < 0 && TF >= 0 && PJ < 0){
    kind = "ENTJ（统帅水鸟）"
    introduce = "你是一只坦诚、果断，有天生的领导能力的水鸟咕咕。能很快看到公司/组织程序和政策中的不合理性和低效能性，发展并实施有效和全面的系统来解决问题。善于做长期的计划和目标的设定。通常见多识广，博览群书，喜欢拓广自己的知识面并将此分享给其他水鸟。在陈述自己的想法时非常强而有力。"
  }
  if(IE < 0 && SN < 0 && TF < 0 && PJ > 0){
    kind = "ENFP（公关水鸟）"
    introduce = "你是一只热情洋溢、富有想象力的水鸟咕咕。认为鸟生有很多的可能性。能很快地将事情和信息联系起来，然后很自信地根据自己的判断解决问题。总是需要得到其他水鸟的认可，也总是准备着给与其他水鸟赏识和帮助。灵活、自然不做作，有很强的即兴发挥的能力，言语流畅。"
  }
  if(IE < 0 && SN < 0 && TF < 0 && PJ < 0){
    kind = "ENFJ（教导水鸟）"
    introduce = "你是一只热情、为他鸟着想、易感应、有责任心的水鸟咕咕。非常注重其他水鸟的感情、需求和动机。善于发现其他水鸟的潜能，并希望能帮助他们实现。能成为水鸟或水鸟群体成长和进步的催化剂。忠诚，对于赞扬和批评都会积极地回应。友善、好社交。在团体中能很好地帮助水鸟，并有鼓舞水鸟的领导能力。"
  }

  db, err := sql.Open("mysql", "root:chapus1215@tcp(cdb-6qucl950.bj.tencentcdb.com:10102)/waterflow_alpha?charset=utf8")
  checkErr(err)
  //查询数据
  rows, err := db.Query("SELECT * FROM waterflow_test")
  checkErr(err)

  var max_similar float64
  var similar_num float64
  max_similar = 9999.0
  var similar_name string
  var similar_imageurl string

  for rows.Next() {
    var uid int
    var name string
    var y1 float64
    var y2 float64
    var y3 float64
    var y4 float64
    var euclidean_num float64

    err = rows.Scan(&uid, &name, &y1, &y2, &y3, &y4)
    euclidean_num = Euclidean(IE,SN,TF,PJ,y1,y2,y3,y4)
    if(max_similar>euclidean_num){
      max_similar = euclidean_num
      similar_name = name
      similar_num = Pearson(IE,SN,TF,PJ,y1,y2,y3,y4)
    }
    checkErr(err)
  }
  db.Close()

  similar_num = Round2(similar_num)
  similar_imageurl = "https://waterflow-image.oss-cn-beijing.aliyuncs.com/" + url.QueryEscape(similar_name) + ".jpg"
  var similar string
  similar = strconv.FormatFloat((similar_num+1)*50,'f',2,64)
  similar += "%"
  var json_file Test_result

  json_file.Ie = IE
  json_file.Sn = SN
  json_file.Tf = TF
  json_file.Pj = PJ
  json_file.Kind = kind
  json_file.Introduce = introduce
  json_file.Similar_name = similar_name
  json_file.Similar_imageurl = similar_imageurl
  json_file.Similar = similar
  res.Header.Set("Content-Type", "application/json; charset=utf-8")
    return res.WriteJSON(json_file)
}

func get_comments(req *air.Request, res *air.Response) error {
  /*
    URL：/getcomments
    方式：GET
    参数：page（非空，int，页码）
    返回格式：json
  */

  pNUM := req.Param("page")
  if pNUM == nil {
    return a.NotFoundHandler(req, res)
  }
  p,_ := pNUM.Value().Int()

  s := Message_board_struct{}
  var json_file Message_board_struct
  json_file.PostTime = time.Now().Format("2006-01-02 15:04:05")

  i := 5*(p-1) + 1//每页显示五条评论

  db, err := sql.Open("mysql", "root:chapus1215@tcp(172.21.0.11:3306)/waterflow_alpha?charset=utf8")
  checkErr(err)
  //查询数据

  for ; i <= 5*p; i++{
    var uid int
    var user_name string
    var comment string
    var mail string
    var sendtime string
    err := db.QueryRow("SELECT * FROM message_board WHERE id=?", i).Scan(&uid, &user_name, &comment, &mail, &sendtime)
    if err == sql.ErrNoRows{
      break;
    } 
    json_file.Board = append(json_file.Board, Message{Uid: uid,Name: user_name,Comment: comment,Mail: mail,SendTime: sendtime})
  }
  db.Close()
  json_fin, err := json.Marshal(json_file)
  if err != nil {
    fmt.Println("json err:", err)
  }
  json.Unmarshal(json_fin, &s)
  res.Header.Set("Content-Type", "application/json; charset=utf-8")
  return res.WriteJSON(s)
}

func get_waterflow_info(req *air.Request, res *air.Response) error {
  /*
    获取所有水鸟简略信息
    URL：/info
    方式：GET
    参数：page（可空，int，页码，default：1）
    返回格式：json
  */
  s := waterflow_info_struct{}
  var json_file waterflow_info_struct
  json_file.PostTime = time.Now().Format("2006-01-02")

  db, err := sql.Open("mysql", "root:chapus1215@tcp(cdb-6qucl950.bj.tencentcdb.com:10102)/waterflow_alpha?charset=utf8")
  checkErr(err)

  var uid int
  var name string
  var Order string
  var Family string
  var Genus string
  var p int 

  pNUM := req.Param("page")
  if pNUM == nil {  
    p = 1
  } else {
    p,_ = pNUM.Value().Int()
  }

  i := 10*(p-1)+1
  for ; i <= 10*p; i++{
    err := db.QueryRow("SELECT * FROM waterflow_info WHERE id=?", i).Scan(&uid, &name, &Order, &Family, &Genus)
    if err == sql.ErrNoRows{
      break;
    } 
    json_file.Waterflow_List = append(json_file.Waterflow_List, Flows{Uid: uid,Name: name,Order: Order,Family: Family, Genus: Genus}) 
  }
  
  db.Close()
  json_fin, err := json.Marshal(json_file)
  if err != nil {
    fmt.Println("json err:", err)
  }
  json.Unmarshal(json_fin, &s)
  res.Header.Set("Content-Type", "application/json; charset=utf-8")
  return res.WriteJSON(s)
}

func get_waterflow_detail(req *air.Request, res *air.Response) error {
  /*
    获取单个水鸟详细信息
    URL：/detail
    方式：GET
    参数：name（非空，string，水鸟名字）
    返回格式：json
  */
  p_NAME := req.Param("name")
  if p_NAME == nil {
    return a.NotFoundHandler(req, res)
  }

  pNAME := req.Param("name").Value().String()
  s := waterflow_detail_struct{}
  var json_file waterflow_detail_struct
  json_file.PostTime = time.Now().Format("2006-01-02")
  
  db, err := sql.Open("mysql", "root:chapus1215@tcp(cdb-6qucl950.bj.tencentcdb.com:10102)/waterflow_alpha?charset=utf8")
  checkErr(err)
  //查询数据
  //rows, err := db.Query("SELECT * FROM waterflow_detail")
  //checkErr(err)

  //for rows.Next() {
    var uid int
    var name string
    var latin_name string
    var introduce string
    var imgurl string
  //  err = rows.Scan(&uid, &name, &latin_name, &introduce, &imgurl)
    err = db.QueryRow("SELECT * FROM waterflow_detail where name=?",pNAME).Scan(&uid, &name, &latin_name, &introduce, &imgurl)
    if err != sql.ErrNoRows{
      json_file.Waterflow_Detail = append(json_file.Waterflow_Detail, Flows_detail{Uid: uid,Name: name,Latin_name: latin_name,Introduce: introduce, Imgurl: imgurl})
    }
    
  //} 
  

  db.Close()
  json_fin, err := json.Marshal(json_file)
  if err != nil {
    fmt.Println("json err:", err)
  }
  json.Unmarshal(json_fin, &s)
  res.Header.Set("Content-Type", "application/json; charset=utf-8")
  return res.WriteJSON(s)

}

func search(req *air.Request, res *air.Response) error {

  /*
    搜索功能
    URL：/s
    方式：GET
    参数：
      name（非空，string，水鸟名字，若为“all”则表示不限制水鸟）
      area（非空，string，地区代码，若为“all”则表示不限制地区（详情见附录1））
      page（可空，int，页码，default：1）
  */
  pNAME := req.Param("name")
  pAREA := req.Param("area")
  pPAGE := req.Param("page")
  p_name := pNAME.Value().String()
  p_area := pAREA.Value().String()
  var page_num int
  if pPAGE == nil {
        page_num = 1
  } else {
    page_num,_ = pPAGE.Value().Int()
  }
  count := 1
  if pAREA == nil || pNAME == nil{//错误
    return a.NotFoundHandler(req, res)
  } else if p_area == "all" && p_name != "all"{//仅搜索名字
      s := waterflow_info_struct{}
      var json_file waterflow_info_struct
      json_file.PostTime = time.Now().Format("2006-01-02")

      db, err := sql.Open("mysql", "root:chapus1215@tcp(cdb-6qucl950.bj.tencentcdb.com:10102)/waterflow_alpha?charset=utf8")
      checkErr(err)
      //查询数据
      rows, err := db.Query("SELECT * FROM waterflow_info where name like ? ORDER BY id asc;","%" + p_name + "%")
      checkErr(err)

      for rows.Next() {
        var uid int
        var name string
        var Order string
        var Family string
        var Genus string
        err = rows.Scan(&uid, &name, &Order, &Family, &Genus)
        checkErr(err)
        if count >= 10*(page_num-1)+1 && count <= 10*page_num {
          json_file.Waterflow_List = append(json_file.Waterflow_List, Flows{Uid: count,Name: name,Order: Order,Family: Family, Genus: Genus})
        }
        count++
      }
      db.Close()
      json_fin, err := json.Marshal(json_file)
      if err != nil {
        fmt.Println("json err:", err)
      }
      json.Unmarshal(json_fin, &s)
      res.Header.Set("Content-Type", "application/json; charset=utf-8")
      return res.WriteJSON(s)
  } else if p_area == "all" && p_name == "all"{//检索所有
      s := waterflow_info_struct{}
      var json_file waterflow_info_struct
      json_file.PostTime = time.Now().Format("2006-01-02")

      db, err := sql.Open("mysql", "root:chapus1215@tcp(cdb-6qucl950.bj.tencentcdb.com:10102)/waterflow_alpha?charset=utf8")
      checkErr(err)
      //查询数据
      rows, err := db.Query("SELECT * FROM waterflow_info ORDER BY id asc;")
      checkErr(err)

      for rows.Next() {
        var uid int
        var name string
        var Order string
        var Family string
        var Genus string
        err = rows.Scan(&uid, &name, &Order, &Family, &Genus)
        checkErr(err)
        if count >= 10*(page_num-1)+1 && count <= 10*page_num {
          json_file.Waterflow_List = append(json_file.Waterflow_List, Flows{Uid: count,Name: name,Order: Order,Family: Family, Genus: Genus})
        }
        count++
      }
      db.Close()
      json_fin, err := json.Marshal(json_file)
      if err != nil {
        fmt.Println("json err:", err)
      }
      json.Unmarshal(json_fin, &s)
      res.Header.Set("Content-Type", "application/json; charset=utf-8")
      return res.WriteJSON(s)
  } else if p_area != "all" && p_name == "all"{//检索地区
      s := waterflow_info_struct{}
      var json_file waterflow_info_struct
      json_file.PostTime = time.Now().Format("2006-01-02")

      db, err := sql.Open("mysql", "root:chapus1215@tcp(cdb-6qucl950.bj.tencentcdb.com:10102)/waterflow_alpha?charset=utf8")
      checkErr(err)
      //查询数据
      rows, err := db.Query("select waterflow_info.id, waterflow_info.name, waterflow_info.Order, waterflow_info.Family, waterflow_info.Genus from waterflow_info, area, waterflow_ref_area where waterflow_info.id = waterflow_ref_area.waterflow_id and (area.area_id=? or area.area_id='OK')and area.area_id = waterflow_ref_area.area_id ORDER BY waterflow_info.id asc;",p_area)
      checkErr(err)

      for rows.Next() {
        var uid int
        var name string
        var Order string
        var Family string
        var Genus string
        err = rows.Scan(&uid, &name, &Order, &Family, &Genus)
        checkErr(err)
        if count >= 10*(page_num-1)+1 && count <= 10*page_num {
          json_file.Waterflow_List = append(json_file.Waterflow_List, Flows{Uid: count,Name: name,Order: Order,Family: Family, Genus: Genus})
        }
        count++
      }
      db.Close()
      json_fin, err := json.Marshal(json_file)
      if err != nil {
        fmt.Println("json err:", err)
      }
      json.Unmarshal(json_fin, &s)
      res.Header.Set("Content-Type", "application/json; charset=utf-8")
      return res.WriteJSON(s)
  } else {//地区＋名字联合检索
      s := waterflow_info_struct{}
      var json_file waterflow_info_struct
      json_file.PostTime = time.Now().Format("2006-01-02")

      db, err := sql.Open("mysql", "root:chapus1215@tcp(cdb-6qucl950.bj.tencentcdb.com:10102)/waterflow_alpha?charset=utf8")
      checkErr(err)
      //查询数据
      rows, err := db.Query("select waterflow_info.id, waterflow_info.name, waterflow_info.Order, waterflow_info.Family, waterflow_info.Genus from waterflow_info, area, waterflow_ref_area where waterflow_info.id = waterflow_ref_area.waterflow_id and (area.area_id=? or area.area_id='OK') and area.area_id = waterflow_ref_area.area_id and name like ? ORDER BY waterflow_info.id asc;",p_area, "%" + p_name + "%")
      checkErr(err)

      for rows.Next() {
        var uid int
        var name string
        var Order string
        var Family string
        var Genus string
        err = rows.Scan(&uid, &name, &Order, &Family, &Genus)
        checkErr(err)
        if count >= 10*(page_num-1)+1 && count <= 10*page_num {
          json_file.Waterflow_List = append(json_file.Waterflow_List, Flows{Uid: count,Name: name,Order: Order,Family: Family, Genus: Genus})
        }
        count++
      }
      db.Close()
      json_fin, err := json.Marshal(json_file)
      if err != nil {
        fmt.Println("json err:", err)
      }
      json.Unmarshal(json_fin, &s)
      res.Header.Set("Content-Type", "application/json; charset=utf-8")
      return res.WriteJSON(s)
  }
}

func checkErr(err error) {
  //检查错误
  if err != nil {
    panic(err)
  }
}

func Success(res *air.Response, data interface{}) error {
  //返回200状态码的json
  res.Status = 200
  if data == nil {
    data = ""
  }
  return res.WriteJSON(Map{
    "code":  0,
    "error": "",
    "data":  data,
  })
}

func RandonFloat(data float64) float64 {
  //随机浮点数
  i := rand.Intn(100)
  j := float64(i-50)/1000
  var ans float64 = data*(1.0+j)
  return ans
}

func Euclidean(x1 float64,x2 float64,x3 float64,x4 float64,y1 float64,y2 float64,y3 float64,y4 float64) float64 {
  //欧式距离
  sum := math.Pow(x1-y1,2) + math.Pow(x2-y2,2) + math.Pow(x3-y3,2) + math.Pow(x4-y4,2)
  return math.Sqrt(sum)
}

func Pearson(x1 float64,x2 float64,x3 float64,x4 float64,y1 float64,y2 float64,y3 float64,y4 float64) float64 {
  //皮尔森相关系数
  var avr_x, avr_y float64
  avr_x = (x1 + x2 + x3 + x4)/4
  avr_y = (y1 + y2 + y3 + y4)/4

  var num1 = (x1-avr_x)*(y1-avr_y) + (x2-avr_x)*(y2-avr_y) + (x3-avr_x)*(y3-avr_y) + (x4-avr_x)*(y4-avr_y)
  var num2 = math.Sqrt(math.Pow((x1-avr_x),2) + math.Pow((x2-avr_x),2) + math.Pow((x3-avr_x),2) + math.Pow((x4-avr_x),2))
  var num3 = math.Sqrt(math.Pow((y1-avr_y),2) + math.Pow((y2-avr_y),2) + math.Pow((y3-avr_y),2) + math.Pow((y4-avr_y),2))

  return num1/(num2*num3)
}

func Round2(f float64) float64 {
  //保留两位小数
  n10 := math.Pow10(4)
  return math.Trunc((f+0.5/n10)*n10) / n10
}
