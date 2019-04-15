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

  a.GET("/getnews", jsontest)
  a.GET("/getwaterflowinfo", get_waterflow_info)
  a.GET("/getwaterflowdetail", get_waterflow_detail)
  a.GET("/getcomments", get_comments)
  a.GET("/s",search)
  a.POST("/api/v1/comments",commentsHandler)
  a.POST("/api/v1/dailypush",dailypushHandler)
  a.POST("/api/v1/test",testHandler)
  a.Serve()
}

func Scrape(req *air.Request, respon *air.Response) error {
  s1 := NewsMessage{}
  // Request the HTML page.
  res, err := http.Get("http://search.shidi.org/default.aspx?keyword=水鸟")
  if err != nil {
    log.Fatal(err)
  }

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

  var json_file NewsMessage
  json_file.PostTime = time.Now().Format("2006-01-02")

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
      json_file.NewsList = append(json_file.NewsList, NewsLists{NewsName: band,NewsUrl: href,Newstime: newstime, Author:author, Body:str, Imageurl:imageurl})
      if i==4{
        return false
      }
    }

    return true
  })

  json_fin, err := json.Marshal(json_file)
  if err != nil {
    fmt.Println("json err:", err)
  }

  json.Unmarshal(json_fin, &s1)
  respon.Header.Set("Content-Type", "application/json; charset=utf-8")
  return respon.WriteJSON(s1)
}

func commentsHandler(req *air.Request, res *air.Response) error {
  user_name_A := req.Param("name").Value()
  content_A := req.Param("content").Value()
  mail_address_A := req.Param("mail").Value()

  name := user_name_A.String()
  content := content_A.String()
  mail_address := mail_address_A.String()
  send_time := time.Now().Format("2006-01-02 15:04:05") 

  db, err := sql.Open("mysql", "root:123456@/test?charset=utf8")
  checkErr(err)

 

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
  i := rand.Intn(4)
  j := rand.Intn(4)

  for ; i == 0; {
    i = rand.Intn(4)
  }
  for ; j == 0 || j==i; {
    j = rand.Intn(4)
  }

  s := waterflow_detail_struct{}
  var json_file waterflow_detail_struct
  json_file.PostTime = time.Now().Format("2006-01-02")

  db, err := sql.Open("mysql", "root:123456@/waterflow_alpha?charset=utf8")
  checkErr(err)
  //查询数据
  rows, err := db.Query("SELECT * FROM waterflow_detail where id=? or id=?", i, j)
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

  db, err := sql.Open("mysql", "root:123456@/waterflow_alpha?charset=utf8")
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

func jsontest(req *air.Request, res *air.Response) error {
  //r.ParseForm()
  s := NewsMessage{}
  str := []byte(`
  {
  "get_time": "2019-04-07",
  "news": [
    {
      "newsname": "美丽的漳江河畔成为鸥类水鸟栖息觅食的天堂",
      "newsurl": "http://www.shidi.org/sf_2D9645FDF1564F6A86E27E7038764723_151_weishanlin.html",
      "Newstime": "2019/4/1 11:03:25",
      "Author": "媒体：原创  作者：福建漳江口",
      "Body": "福建漳江口红树林国家级自然保护区位于云霄县漳江入海口，是以红树植物、湿地水鸟、珍稀水产种质资源为主要对象的湿地类型自然保护区，近年来，保护区始终坚持“共建与管理先行，保护与发展并重”，致力于“人与自然和谐发展”，各项工作均取得长足发展，区内红树林面积不断扩大，生态环境日益改善，湿地水鸟显著增多，保护区的东升埭海域滩涂及养殖池塘已经成为了鸥类水鸟活动的场所，大量的红嘴鸥、红嘴巨鸥、黑嘴鸥等都在这里栖息觅食，形成一幅和谐的美丽画卷。\\n",
      "Imageurl": "https://waterflow-scrapy.oss-cn-beijing.aliyuncs.com/HpWPWBM1elCQSSxNVCVMUA"
    },
    {
      "newsname": "东洞庭湖保护区客人到访升金湖观测小白额雁等水鸟越冬情况",
      "newsurl": "http://www.shidi.org/sf_7A1A784D66AA47188CA258B0B370C93A_151_8129996.html",
      "Newstime": "2019/3/15 10:14:42",
      "Author": "媒体：原创  作者：升金湖",
      "Body": "3月9日-10日，湖南东洞庭湖国家级管理局姚毅总工、张鸿副局长等一行4人到安徽升金湖国家级自然保护区就小白额雁等越冬水鸟情况开展调查。升金湖管理局科研救护中心工作人员进行了陪同。东洞庭湖管理局一行先后到升金湖保护区雁类栖息地、觅食地等雁类集中地进行观测，共观测到小白额雁2300余只，小天鹅370只，白头鹤240只，鸿雁6500只，一行人员对升金湖优质的越冬环境表示了高度认可。之后双方就小白额雁的数量、越冬迁飞时间、种群分布做了深入的讨论交流，并达成2019年年底候鸟越冬期同步监测合作意向，同步监测数据将为长江中下游候鸟越冬迁飞变化分析提供有力支撑。\\n注：小白额雁（学名：Anser erythropus）体长53-66厘米，翼展120-135 厘米，体重1400-2300克。腿桔黄色，环嘴基有白斑，腹部具近黑色斑块。极似白额雁，冬季常与其混群。不同处在于体型较小，嘴、颈较短，嘴周围白色斑块延伸至额部，眼圈黄色，腹部暗色块较。飞行时双翼拍打用力，振翅频率高。小白额雁中型雁类，外形和白额雁相似，但体形较白额雁小，体色较深，嘴、脚亦较白额雁短；而额部白斑却较白额雁大，一直延伸到两眼之间的头顶部，不像白额雁仅及嘴基；另外小白额雁眼周金黄色，而白额雁不为金黄色，这些差异，足以将它们的野外区别开来。小白额雁与其他雁属鸟类一样，以绿色植物的茎叶和植物种子为食，湖岸附近生长的杂草、湖中的水草、农田中的绿色作物、谷物、草籽、树叶、嫩芽等皆为取食的对象。目前全球小白额雁的数量已不到3.5万只，在中国，虽然该物种并未列入保护名录，但实际上全球种群数量非常稀少，在中国境内更是难得一见。\\n                                                                          （科研救护中心）\\n",
      "Imageurl": "https://waterflow-scrapy.oss-cn-beijing.aliyuncs.com/HpWPWBR7wTCQSSxNVCVMUA"
    },
    {
      "newsname": "湖北圆满完成2019年越冬水鸟调查",
      "newsurl": "http://www.shidi.org/sf_FFE22AAFCA174CCF83EE1E1705423E1B_151_B002D016491.html",
      "Newstime": "2019/2/14 18:11:13",
      "Author": "媒体：原创  作者：朱湖国家湿地公园",
      "Body": "2月14日，湖北省林业局发布2019年越冬水鸟调查情况通报，通过全省31个点同步调查，共记录越冬水鸟 78 种，种群数量 427049 只，高质量完成调查任务。\\n \\n为进一步摸清全省越冬水鸟资源状况，掌握越冬水鸟资源的动态变化，湖北省林业局野生动植物保护总站牵头组织了 2019 年1月21日至23日的全省越冬水鸟同步调查，各相关市、州、县林业局、湿地公园管理局、自然保护区管理局按要求圆满完成了调查任务。\\n \\n此次调查综合考虑湖北省水鸟分布特点，共选定 31 个点同步调查地点，总面积 4653.26km 2 。调查期间，各调查队采用直接计数法，记录水鸟的种类、种群数量，并对水鸟的分布状况、栖息地状况、水域面积以及水鸟生存受干扰、受威胁因素和平保护现状等进行系统调查。\\n \\n经野外调查、内业整理，此次共调查记录到越冬水鸟 78 种，种群数量 427049 只。其中以雁鸭类为最多，共 29 种，种群数量达 261655 只，占总数量的 61.3%。在 78 种越冬水鸟中，国家重点保护物种 11 种，种群数量为 34639 只，占总数量的 8.1%。其中，国家Ⅰ级有中华秋沙鸭、白鹤、白头鹤、黑鹳、东方白鹳 5 种，Ⅱ级有白额雁、红胸黑雁、小天鹅、鸳鸯、灰鹤、白琵鹭 6 种。其中以小天鹅的种群数量为最大，达 29746 只，占国家重点保护鸟类种群数量的 85.9%。其次为白琵鹭。在 31 个调查地点中，水鸟种群数量超过 1 万只的调查地点有 11 个。龙感湖的水鸟种群数量最大，为 72450 只，其次为洪湖（69338 只）、网湖（57944 只）。网湖水鸟种类最多，达 58种，其次为沉湖和府河，均为 40 种。\\n",
      "Imageurl": "https://waterflow-scrapy.oss-cn-beijing.aliyuncs.com/HpWPWBqjkKCQSSxNVCVMUA"
    },
    {
      "newsname": "保护区2019年1月越冬水鸟同步调查工作顺利开展",
      "newsurl": "http://www.shidi.org/sf_D1241E451DB14414BB01C193E8F1874C_151_66FA58E1101.html",
      "Newstime": "2019/2/11 15:51:32",
      "Author": "媒体：盐城珍禽自然保护区  作者：吴爱鑫 周玲",
      "Body": "2019年1月25日，江苏盐城国家级珍禽自然保护区管理处组织开展了冬季越冬水鸟同步调查工作。此次调查主要目的是为了监测保护区内鹤类、雁鸭类等越冬水鸟的种群、数量和分布情况，揭示鸟类群落的动态变化，积累鸟类调查数据，为保护区的生态修复、巡护管理等提供科学依据。\\n23日上午同步调查开始前，保护区管理处召开冬季越冬水鸟同步调查动员大会。蒋巍副市长提出本次同步调查是空间小尺度，通过调查找准自然保护区核心价值，向科学考察方向体系加快靠拢，树立一流目标。会议由吕明光副主任主持；陈志会书记作动员讲话，陈书记从提高认识站位、把握重点要求和精心组织调查三个方面号召大家齐心协力、严谨细致、精益求精的做好此次调查工作；陈浩副主任作此次调查技术培训及宣读分组通知。\\n本次调查共分十个小组，调查区域包括保护区核心区、缓冲区及实验区，参与调查的人员为各科室主要负责人、全体专业技术人员、管护站工作人员、景区工作人员。此次调查遵循科学性、严谨性要求，采用路线调查法、定点直接计数法和视频监控法相结合的方法，借助双筒望远镜、单筒望远镜、相机、无人机等设备，监测鸟类活动、人为活动及土地开发情况。至2019年1月25日晚6点各调查小组全部完成调查工作，各小组调查数据汇总后由科研科同事整合编写调查报告并存档保存。\\n鸟调现场\\n鸟调现场\\n鸟调现场\\n野生丹顶鹤一家三口、银鸥\\n鹤戏金滩\\n普通鸬鹚集群\\n鸟类同步调查是一项专业性高、调查范围广、时间紧任务重、协作要求严的工作，保护区调查人员多年来始终坚持认真做好区域内鸟类调查工作并收集整理鸟类调查数据，为保护鸟类、生态湿地多做贡献。今后保护区全体职工也会一如既往开展好调查工作，久久为功，为建设世界一流生态湿地、鸟类家园而努力。\\n",
      "Imageurl": "https://waterflow-scrapy.oss-cn-beijing.aliyuncs.com/HpWPWBwW1lCQSSxNVCVMUA"
    }
  ]
  }`)
  json.Unmarshal(str, &s)
  res.Header.Set("Content-Type", "application/json; charset=utf-8")
  return res.WriteJSON(s)
}

func get_comments(req *air.Request, res *air.Response) error {
  pNUM := req.Param("page")
  if pNUM == nil {
    return a.NotFoundHandler(req, res)
  }
  p,_ := pNUM.Value().Int()

  s := Message_board_struct{}
  var json_file Message_board_struct
  json_file.PostTime = time.Now().Format("2006-01-02 15:04:05")

  i := 5*(p-1) + 1

  db, err := sql.Open("mysql", "root:123456@/test?charset=utf8")
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
  s := waterflow_info_struct{}
  var json_file waterflow_info_struct
  json_file.PostTime = time.Now().Format("2006-01-02")

  db, err := sql.Open("mysql", "root:123456@/waterflow_alpha?charset=utf8")
  checkErr(err)
  //查询数据
  rows, err := db.Query("SELECT * FROM waterflow_info")
  checkErr(err)

  for rows.Next() {
    var uid int
    var name string
    var Order string
    var Family string
    var Genus string
    err = rows.Scan(&uid, &name, &Order, &Family, &Genus)
    checkErr(err)
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
  s := waterflow_detail_struct{}
  var json_file waterflow_detail_struct
  json_file.PostTime = time.Now().Format("2006-01-02")

  db, err := sql.Open("mysql", "root:123456@/waterflow_alpha?charset=utf8")
  checkErr(err)
  //查询数据
  rows, err := db.Query("SELECT * FROM waterflow_detail")
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

func search(req *air.Request, res *air.Response) error {
  pNAME := req.Param("name")
  pAREA := req.Param("area")
  p_name := pNAME.Value().String()
  p_area := pAREA.Value().String()
  if pAREA == nil || pNAME == nil{
    return a.NotFoundHandler(req, res)
  } else if p_area == "all" && p_name != "all"{
      
      s := waterflow_info_struct{}
      var json_file waterflow_info_struct
      json_file.PostTime = time.Now().Format("2006-01-02")

      db, err := sql.Open("mysql", "root:123456@/waterflow_alpha?charset=utf8")
      checkErr(err)
      //查询数据
      rows, err := db.Query("SELECT * FROM waterflow_info where name like ?","%" + p_name + "%")
      checkErr(err)

      for rows.Next() {
        var uid int
        var name string
        var Order string
        var Family string
        var Genus string
        err = rows.Scan(&uid, &name, &Order, &Family, &Genus)
        checkErr(err)
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

  } else if p_area == "all" && p_name == "all"{
      s := waterflow_info_struct{}
      var json_file waterflow_info_struct
      json_file.PostTime = time.Now().Format("2006-01-02")

      db, err := sql.Open("mysql", "root:123456@/waterflow_alpha?charset=utf8")
      checkErr(err)
      //查询数据
      rows, err := db.Query("SELECT * FROM waterflow_info")
      checkErr(err)

      for rows.Next() {
        var uid int
        var name string
        var Order string
        var Family string
        var Genus string
        err = rows.Scan(&uid, &name, &Order, &Family, &Genus)
        checkErr(err)
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
  } else if p_area != "all" && p_name == "all"{
      s := waterflow_info_struct{}
      var json_file waterflow_info_struct
      json_file.PostTime = time.Now().Format("2006-01-02")

      db, err := sql.Open("mysql", "root:123456@/waterflow_alpha?charset=utf8")
      checkErr(err)
      //查询数据
      rows, err := db.Query("select waterflow_info.id, waterflow_info.name, waterflow_info.Order, waterflow_info.Family, waterflow_info.Genus from waterflow_info, area, waterflow_REF_area where waterflow_info.id = waterflow_REF_area.waterflow_id and Area.area_id=? and Area.area_id = waterflow_REF_area.area_id",p_area)
      checkErr(err)

      for rows.Next() {
        var uid int
        var name string
        var Order string
        var Family string
        var Genus string
        err = rows.Scan(&uid, &name, &Order, &Family, &Genus)
        checkErr(err)
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
  } else {
      s := waterflow_info_struct{}
      var json_file waterflow_info_struct
      json_file.PostTime = time.Now().Format("2006-01-02")

      db, err := sql.Open("mysql", "root:123456@/waterflow_alpha?charset=utf8")
      checkErr(err)
      //查询数据
      rows, err := db.Query("select waterflow_info.id, waterflow_info.name, waterflow_info.Order, waterflow_info.Family, waterflow_info.Genus from waterflow_info, area, waterflow_REF_area where waterflow_info.id = waterflow_REF_area.waterflow_id and Area.area_id=? and Area.area_id = waterflow_REF_area.area_id and name like ?",p_area, "%" + p_name + "%")
      checkErr(err)

      for rows.Next() {
        var uid int
        var name string
        var Order string
        var Family string
        var Genus string
        err = rows.Scan(&uid, &name, &Order, &Family, &Genus)
        checkErr(err)
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
  
}

func checkErr(err error) {
  if err != nil {
    panic(err)
  }
}

func Success(res *air.Response, data interface{}) error {
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
  i := rand.Intn(100)
  j := float64(i-50)/1000
  var ans float64 = data*(1.0+j)
  return ans
}

func Euclidean(x1 float64,x2 float64,x3 float64,x4 float64,y1 float64,y2 float64,y3 float64,y4 float64) float64 {

  sum := math.Pow(x1-y1,2) + math.Pow(x2-y2,2) + math.Pow(x3-y3,2) + math.Pow(x4-y4,2)
  return math.Sqrt(sum)
}

func Pearson(x1 float64,x2 float64,x3 float64,x4 float64,y1 float64,y2 float64,y3 float64,y4 float64) float64 {
  var avr_x, avr_y float64

  avr_x = (x1 + x2 + x3 + x4)/4
  avr_y = (y1 + y2 + y3 + y4)/4

  var num1 = (x1-avr_x)*(y1-avr_y) + (x2-avr_x)*(y2-avr_y) + (x3-avr_x)*(y3-avr_y) + (x4-avr_x)*(y4-avr_y)
  var num2 = math.Sqrt(math.Pow((x1-avr_x),2) + math.Pow((x2-avr_x),2) + math.Pow((x3-avr_x),2) + math.Pow((x4-avr_x),2))
  var num3 = math.Sqrt(math.Pow((y1-avr_y),2) + math.Pow((y2-avr_y),2) + math.Pow((y3-avr_y),2) + math.Pow((y4-avr_y),2))

  return num1/(num2*num3)
}
func Round2(f float64) float64 {
  n10 := math.Pow10(4)
  return math.Trunc((f+0.5/n10)*n10) / n10
}
