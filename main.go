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
)

var (
  a = air.Default
  errorhtml string
)

type NewsMessage struct {
  PostTime string `json:"get_time"`
  NewsList []News `json:"news"`
}

type News struct {
  NewsName string `json:"newsname"`
  NewsUrl   string  `json:"newsurl"`
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

type Map map[string]interface{}

func main() {
  a.DebugMode = true
  a.FILE("/error", "templates/error.html") //var errorhtml string
  a.FILE("/building", "templates/build.html") //var errorhtml string
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
  a.GET("/waterflow_info/name=:NAME",namesearch)
  a.GET("/waterflow_info/area=:ID",areasearch)
  a.POST("/api/v1/comments",commentsHandler)
  a.POST("/api/v1/dailypush",dailypushHandler)
  a.POST("/api/v1/test",testHandler)
  a.Serve()
}

func Scrape(req *air.Request, respon *air.Response) error {
  s := NewsMessage{}
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

  doc.Find("ul").Find("li").Each(func(i int, s *goquery.Selection) {
    // For each item found, get the band and title
    band := s.Find("a").Text()
    href := s.Find(".siteurl").Text()
    json_file.NewsList = append(json_file.NewsList, News{NewsName: band,NewsUrl: href})
    //fmt.Printf("Review %d: %s - %s\n", i, band, href)
  })

  json_fin, err := json.Marshal(json_file)
  if err != nil {
    fmt.Println("json err:", err)
  }

  json.Unmarshal(json_fin, &s)
  respon.Header.Set("Content-Type", "application/json; charset=utf-8")
  return respon.WriteJSON(s)
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


  user_gender, err := req.Param("gender").Value().Float64()
  if err != nil{
    panic(err)
  }

  t := 2.0
  t = RandonFloat(t)
  fmt.Println(user_gender)
  fmt.Println(t)
  return Success(res, "")
}

func jsontest(req *air.Request, res *air.Response) error {
  //r.ParseForm()
  s := NewsMessage{}
  str := []byte(`
{
  "get_time": "2019-02-28",
  "news": [{
    "newsname": "湖北圆满完成2019年越冬水鸟调查",
    "newsurl": "http://www.shidi.org/sf_FFE22AAFCA174CCF83EE1E1705423E1B_151_B002D016491.html"
  }, {
    "newsname": "99种467686只 北大港湿地水鸟分布新记录！",
    "newsurl": "http://www.shidi.org/sf_D0062DE46DD74D178BB187720A43B3EF_151_66FA58E1101.html"
  }, {
    "newsname": "保护区2019年1月越冬水鸟同步调查工作顺利开展",
    "newsurl": "http://www.shidi.org/sf_D1241E451DB14414BB01C193E8F1874C_151_66FA58E1101.html"
  }, {
    "newsname": "洈水国家湿地公园开展2019年越冬水鸟同步调查",
    "newsurl": "http://www.shidi.org/sf_7EEFA30AB92B4DA8A078FF34CF1F0CA3_151_04BD4275188.html"
  }, {
    "newsname": "金沙湖湿地水鸟",
    "newsurl": "http://www.shidi.org/sf_74D86A3F628A49A193C32682BF7143EC_151_249EA316522.html"
  }, {
    "newsname": "湖北开展越冬水鸟同步调查",
    "newsurl": "http://www.shidi.org/sf_3E80868D0317463BBD0E7A63EB008F81_151_B002D016491.html"
  }, {
    "newsname": "朱湖湿地让越冬水鸟过好年",
    "newsurl": "http://www.shidi.org/sf_2363C2BDCE10485BB703947C444CE7DD_151_B002D016491.html"
  }, {
    "newsname": "水鸟对气候变化的反应与其首选的越冬栖息地有关",
    "newsurl": "http://www.shidi.org/sf_E845C488C8CD4C2C98DA3E9A9B7A8A9C_151_66FA58E1101.html"
  }, {
    "newsname": "《上海水鸟观察入门指南》开放申领",
    "newsurl": "http://www.shidi.org/sf_C403AA3E348540BD9817C17CC20DA7EF_151_66FA58E1101.html"
  }, {
    "newsname": "临泽双泉湖湿地成水鸟冬季乐园",
    "newsurl": "http://www.shidi.org/sf_E805A4CCB5CC4DAAAC559196FD36607F_151_DDC156A0949.html"
  }]
}`)
  json.Unmarshal(str, &s)
  res.Header.Set("Content-Type", "application/json; charset=utf-8")
  return res.WriteJSON(s)
}

func get_comments(req *air.Request, res *air.Response) error {
  s := Message_board_struct{}
  var json_file Message_board_struct
  json_file.PostTime = time.Now().Format("2006-01-02 15:04:05")

  db, err := sql.Open("mysql", "root:123456@/test?charset=utf8")
  checkErr(err)
  //查询数据
  rows, err := db.Query("SELECT * FROM message_board")
  checkErr(err)

  for rows.Next() {
    var uid int
    var user_name string
    var comment string
    var mail string
    var sendtime string
    err = rows.Scan(&uid, &user_name, &comment, &mail, &sendtime)
    checkErr(err)
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

func namesearch(req *air.Request, res *air.Response) error {
  pID := req.Param("NAME")
  if pID == nil {
    return a.NotFoundHandler(req, res)
  }
  p := pID.Value().String()
  s := waterflow_info_struct{}
  var json_file waterflow_info_struct
  json_file.PostTime = time.Now().Format("2006-01-02")

  db, err := sql.Open("mysql", "root:123456@/waterflow_alpha?charset=utf8")
  checkErr(err)
  //查询数据
  rows, err := db.Query("SELECT * FROM waterflow_info where name like ?","%" + p + "%")
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

func areasearch(req *air.Request, res *air.Response) error {
  pID := req.Param("ID")
  if pID == nil {
    return a.NotFoundHandler(req, res)
  }
  p := pID.Value().String()
  fmt.Println(p)
  s := waterflow_info_struct{}
  var json_file waterflow_info_struct
  json_file.PostTime = time.Now().Format("2006-01-02")

  db, err := sql.Open("mysql", "root:123456@/waterflow_alpha?charset=utf8")
  checkErr(err)
  //查询数据
  rows, err := db.Query("select waterflow_info.id, waterflow_info.name, waterflow_info.Order, waterflow_info.Family, waterflow_info.Genus from waterflow_info, area, waterflow_REF_area where waterflow_info.id = waterflow_REF_area.waterflow_id and Area.area_id=? and Area.area_id = waterflow_REF_area.area_id",p)
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