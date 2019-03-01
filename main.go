package main

import (
  "fmt"
  "log"
  "net/http"
  "github.com/PuerkitoBio/goquery"
  "encoding/json"
  "time"
  "github.com/aofei/air"
)

var a = air.Default

type NewsMessage struct {
  PostTime string `json:"get_time"`
  NewsList []News `json:"news"`
}

type News struct {
  NewsName string `json:"newsname"`
  NewsUrl   string  `json:"newsurl"`
}

func main() {
  //Scrape()
  a.DebugMode = true
  a.GET("/getnews", jsontest)
  a.Serve()
}

func Scrape(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
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
  for i := 0; i <= 9; i++{
    fmt.Fprintf(w,"<p>")
    fmt.Fprintf(w,"<a href=" + s.NewsList[i].NewsUrl + ">")
    fmt.Fprintf(w,s.NewsList[i].NewsName)
    fmt.Fprintf(w,"</a>")
    fmt.Fprintf(w,"</p>")
    
  }
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

