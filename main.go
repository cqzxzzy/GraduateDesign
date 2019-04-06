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

  doc.Find("ul").Find("li").Each(func(i int, s *goquery.Selection) {
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

    indoc.Find("#endText").Find("p").Each(func(i_tmp int, s_tmp *goquery.Selection) {
        str = str + s_tmp.Text() + "\\n"
        str = strings.Replace(str, "\\n\\n", "\\n", -1)
    })

    json_file.NewsList = append(json_file.NewsList, NewsLists{NewsName: band,NewsUrl: href,Newstime: newstime, Author:author, Body:str})
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
  "get_time": "2019-04-06",
  "news": [
    {
      "newsname": "美丽的漳江河畔成为鸥类水鸟栖息觅食的天堂",
      "newsurl": "http://www.shidi.org/sf_2D9645FDF1564F6A86E27E7038764723_151_weishanlin.html",
      "Newstime": "2019/4/1 11:03:25",
      "Author": "媒体：原创  作者：福建漳江口",
      "Body": "福建漳江口红树林国家级自然保护区位于云霄县漳江入海口，是以红树植物、湿地水鸟、珍稀水产种质资源为主要对象的湿地类型自然保护区，近年来，保护区始终坚持“共建与管理先行，保护与发展并重”，致力于“人与自然和谐发展”，各项工作均取得长足发展，区内红树林面积不断扩大，生态环境日益改善，湿地水鸟显著增多，保护区的东升埭海域滩涂及养殖池塘已经成为了鸥类水鸟活动的场所，大量的红嘴鸥、红嘴巨鸥、黑嘴鸥等都在这里栖息觅食，形成一幅和谐的美丽画卷。\\n"
    },
    {
      "newsname": "东洞庭湖保护区客人到访升金湖观测小白额雁等水鸟越冬情况",
      "newsurl": "http://www.shidi.org/sf_7A1A784D66AA47188CA258B0B370C93A_151_8129996.html",
      "Newstime": "2019/3/15 10:14:42",
      "Author": "媒体：原创  作者：升金湖",
      "Body": "3月9日-10日，湖南东洞庭湖国家级管理局姚毅总工、张鸿副局长等一行4人到安徽升金湖国家级自然保护区就小白额雁等越冬水鸟情况开展调查。升金湖管理局科研救护中心工作人员进行了陪同。东洞庭湖管理局一行先后到升金湖保护区雁类栖息地、觅食地等雁类集中地进行观测，共观测到小白额雁2300余只，小天鹅370只，白头鹤240只，鸿雁6500只，一行人员对升金湖优质的越冬环境表示了高度认可。之后双方就小白额雁的数量、越冬迁飞时间、种群分布做了深入的讨论交流，并达成2019年年底候鸟越冬期同步监测合作意向，同步监测数据将为长江中下游候鸟越冬迁飞变化分析提供有力支撑。\\n注：小白额雁（学名：Anser erythropus）体长53-66厘米，翼展120-135 厘米，体重1400-2300克。腿桔黄色，环嘴基有白斑，腹部具近黑色斑块。极似白额雁，冬季常与其混群。不同处在于体型较小，嘴、颈较短，嘴周围白色斑块延伸至额部，眼圈黄色，腹部暗色块较。飞行时双翼拍打用力，振翅频率高。小白额雁中型雁类，外形和白额雁相似，但体形较白额雁小，体色较深，嘴、脚亦较白额雁短；而额部白斑却较白额雁大，一直延伸到两眼之间的头顶部，不像白额雁仅及嘴基；另外小白额雁眼周金黄色，而白额雁不为金黄色，这些差异，足以将它们的野外区别开来。小白额雁与其他雁属鸟类一样，以绿色植物的茎叶和植物种子为食，湖岸附近生长的杂草、湖中的水草、农田中的绿色作物、谷物、草籽、树叶、嫩芽等皆为取食的对象。目前全球小白额雁的数量已不到3.5万只，在中国，虽然该物种并未列入保护名录，但实际上全球种群数量非常稀少，在中国境内更是难得一见。\\n                                                                          （科研救护中心）\\n"
    },
    {
      "newsname": "湖北圆满完成2019年越冬水鸟调查",
      "newsurl": "http://www.shidi.org/sf_FFE22AAFCA174CCF83EE1E1705423E1B_151_B002D016491.html",
      "Newstime": "2019/2/14 18:11:13",
      "Author": "媒体：原创  作者：朱湖国家湿地公园",
      "Body": "2月14日，湖北省林业局发布2019年越冬水鸟调查情况通报，通过全省31个点同步调查，共记录越冬水鸟 78 种，种群数量 427049 只，高质量完成调查任务。\\n \\n为进一步摸清全省越冬水鸟资源状况，掌握越冬水鸟资源的动态变化，湖北省林业局野生动植物保护总站牵头组织了 2019 年1月21日至23日的全省越冬水鸟同步调查，各相关市、州、县林业局、湿地公园管理局、自然保护区管理局按要求圆满完成了调查任务。\\n \\n此次调查综合考虑湖北省水鸟分布特点，共选定 31 个点同步调查地点，总面积 4653.26km 2 。调查期间，各调查队采用直接计数法，记录水鸟的种类、种群数量，并对水鸟的分布状况、栖息地状况、水域面积以及水鸟生存受干扰、受威胁因素和平保护现状等进行系统调查。\\n \\n经野外调查、内业整理，此次共调查记录到越冬水鸟 78 种，种群数量 427049 只。其中以雁鸭类为最多，共 29 种，种群数量达 261655 只，占总数量的 61.3%。在 78 种越冬水鸟中，国家重点保护物种 11 种，种群数量为 34639 只，占总数量的 8.1%。其中，国家Ⅰ级有中华秋沙鸭、白鹤、白头鹤、黑鹳、东方白鹳 5 种，Ⅱ级有白额雁、红胸黑雁、小天鹅、鸳鸯、灰鹤、白琵鹭 6 种。其中以小天鹅的种群数量为最大，达 29746 只，占国家重点保护鸟类种群数量的 85.9%。其次为白琵鹭。在 31 个调查地点中，水鸟种群数量超过 1 万只的调查地点有 11 个。龙感湖的水鸟种群数量最大，为 72450 只，其次为洪湖（69338 只）、网湖（57944 只）。网湖水鸟种类最多，达 58种，其次为沉湖和府河，均为 40 种。\\n"
    },
    {
      "newsname": "99种467686只 北大港湿地水鸟分布新记录！",
      "newsurl": "http://www.shidi.org/sf_D0062DE46DD74D178BB187720A43B3EF_151_66FA58E1101.html",
      "Newstime": "2019/2/11 15:55:04",
      "Author": "媒体：天津日报  作者：王睿",
      "Body": "从2018年开始到现在，由保尔森基金会和河仁慈善基金会支持的《天津北大港湿地生态环境及候鸟资源监测项目》在一年多的调查期间共记录到99种467686只水鸟在北大港湿地停歇、繁殖或越冬，其中不乏世界易危或濒危物种。这刷新了北大港湿地水鸟分布的记录，也意味着这里已成为东亚—澳大利西亚水鸟迁飞路线上最为重要的栖息地之一。\\n此次调查有不少重要发现，其中在北大港单次记录的易危水鸟遗鸥最大数量为12767只，已超过湿地国际对该物种全球种群数量的估计值，而在滨海湿地滩涂越冬的遗鸥也已达上万只；在北大港单次记录到濒危水鸟东方白鹳最大数量为1347只，占其全球种群数量的45%；在北大港单次记录到濒危物种白枕鹤293只，占其全国种群数量的20%左右，北大港湿地正在成为白枕鹤重要的栖息地。与此同时，在2018年10月至11月期间，调查还发现了一些区域新鸟种，包括白头硬尾鸭、白额鹱、红喉潜鸟和雪雁。\\n据了解，基于调研成果，调查团队将制定《北大港湿地监测方案和技术规程》，指导北大港湿地自然保护区管理中心以此为依据，开展长期持续的监测，更好地了解东亚—澳大利西亚迁飞路线上迁徙水鸟的现有种群及其发展趋势，以便对水鸟面临的威胁提供早期预警。\\n"
    },
    {
      "newsname": "保护区2019年1月越冬水鸟同步调查工作顺利开展",
      "newsurl": "http://www.shidi.org/sf_D1241E451DB14414BB01C193E8F1874C_151_66FA58E1101.html",
      "Newstime": "2019/2/11 15:51:32",
      "Author": "媒体：盐城珍禽自然保护区  作者：吴爱鑫 周玲",
      "Body": "2019年1月25日，江苏盐城国家级珍禽自然保护区管理处组织开展了冬季越冬水鸟同步调查工作。此次调查主要目的是为了监测保护区内鹤类、雁鸭类等越冬水鸟的种群、数量和分布情况，揭示鸟类群落的动态变化，积累鸟类调查数据，为保护区的生态修复、巡护管理等提供科学依据。\\n23日上午同步调查开始前，保护区管理处召开冬季越冬水鸟同步调查动员大会。蒋巍副市长提出本次同步调查是空间小尺度，通过调查找准自然保护区核心价值，向科学考察方向体系加快靠拢，树立一流目标。会议由吕明光副主任主持；陈志会书记作动员讲话，陈书记从提高认识站位、把握重点要求和精心组织调查三个方面号召大家齐心协力、严谨细致、精益求精的做好此次调查工作；陈浩副主任作此次调查技术培训及宣读分组通知。\\n本次调查共分十个小组，调查区域包括保护区核心区、缓冲区及实验区，参与调查的人员为各科室主要负责人、全体专业技术人员、管护站工作人员、景区工作人员。此次调查遵循科学性、严谨性要求，采用路线调查法、定点直接计数法和视频监控法相结合的方法，借助双筒望远镜、单筒望远镜、相机、无人机等设备，监测鸟类活动、人为活动及土地开发情况。至2019年1月25日晚6点各调查小组全部完成调查工作，各小组调查数据汇总后由科研科同事整合编写调查报告并存档保存。\\n鸟调现场\\n鸟调现场\\n鸟调现场\\n野生丹顶鹤一家三口、银鸥\\n鹤戏金滩\\n普通鸬鹚集群\\n鸟类同步调查是一项专业性高、调查范围广、时间紧任务重、协作要求严的工作，保护区调查人员多年来始终坚持认真做好区域内鸟类调查工作并收集整理鸟类调查数据，为保护鸟类、生态湿地多做贡献。今后保护区全体职工也会一如既往开展好调查工作，久久为功，为建设世界一流生态湿地、鸟类家园而努力。\\n"
    },
    {
      "newsname": "洈水国家湿地公园开展2019年越冬水鸟同步调查",
      "newsurl": "http://www.shidi.org/sf_7EEFA30AB92B4DA8A078FF34CF1F0CA3_151_04BD4275188.html",
      "Newstime": "2019/1/24 17:10:16",
      "Author": "媒体：原创  作者：洈水国家湿地公园",
      "Body": "湖北省越冬水鸟同步调查，2019年1月21日-23日，湖北松滋洈水国家湿地公园共监测到越冬水鸟16种900只。\\n      洈水国家湿地公园是水鸟的天堂。1月，湿地上空阳光明媚，湖汊纵横，滩涂裸露，水草肆意铺展，让越冬水鸟有了更广阔的“嬉戏”舞台。\\n大批越冬水鸟陆续光临，在此安营扎寨，迎接又一个春天的到来。这是一年一度的洈水国家湿地公园越冬水鸟同步调查的最好时节。1月21日至23日，湖北松滋洈水国家湿地公园工作人员联合荆州观鸟协会、松滋市林业局等单位，开始了一次爱鸟之人与鸟儿的约会，延续人与自然和谐相处。\\n两路人马下江湖，经过10多位伙伴的艰辛努力，调查人员共监测到越冬水鸟16种900只，其中监测到“国宝”级动物中华秋沙鸭136只。\\n今年的鸟类调查与往年有何不同?“方案更加成熟，监测点多面广，更科学有效。”洈水国家湿地公园管理处工作人员说，“水鸟的数量，是洈水生态环境的重要指标。调查区域在整个洈水水域，特别引人注目的是，为保证鸟类监测的科学有效，此次鸟类调查采用了许多先进设备，单筒望远镜、双筒望远镜、GPS定位跟踪仪、鸟类图谱、数据记录工具等一应俱全。”\\n众里寻它千百度，“黑鹳!黑鹳！”1月22日13时，第一组在洈水湿地首次发现一只国家一级保护动物黑鹳，它泛着黑色光泽，眼红色，太美了!监测队员越看越喜欢，久久不肯走远。\\n"
    },
    {
      "newsname": "金沙湖湿地水鸟",
      "newsurl": "http://www.shidi.org/sf_74D86A3F628A49A193C32682BF7143EC_151_249EA316522.html",
      "Newstime": "2019/1/16 9:06:30",
      "Author": "媒体：原创  作者：金沙湖湿地",
      "Body": " \\n \\n"
    },
    {
      "newsname": "湖北开展越冬水鸟同步调查",
      "newsurl": "http://www.shidi.org/sf_3E80868D0317463BBD0E7A63EB008F81_151_B002D016491.html",
      "Newstime": "2019/1/11 16:45:22",
      "Author": "媒体：原创  作者：朱湖国家湿地公园",
      "Body": " \\n为摸清全省越冬水鸟资源状况，掌握越冬水鸟资源的动态变化，1月9日，湖北省林业局下发通知：定于2019年1月21-23日在全省开展越冬水鸟同步调查工作。\\n要求各有关地区和国家湿地公园切实做好统筹协调，加强组织领导，明确专人负责，认真做好本次越冬水鸟同步调查工作。同时，本次未涉及调查的湿地自然保护区和湿地公园，要以本次越冬水鸟同步调查工作为契机，按照要求，自行做好相关调查工作，并形成调查原始数据和调查成果（含影像资料），便于以后工作使用。\\n为落实省林业局通知要求，朱湖国家湿地公园管理处积极组织人员开展越冬水鸟同步调查的准备工作。\\n \\n \\n"
    },
    {
      "newsname": "朱湖湿地让越冬水鸟过好年",
      "newsurl": "http://www.shidi.org/sf_2363C2BDCE10485BB703947C444CE7DD_151_B002D016491.html",
      "Newstime": "2019/1/9 19:52:39",
      "Author": "媒体：原创  作者：朱湖国家湿地公园",
      "Body": " \\n1月9日，在湖北孝感朱湖国家湿地公园漫天的鸟儿令人眼花缭乱、应接不暇。为进一步加强对湿地鸟类的保护，管理处工作人员正积极准备开展越冬水鸟调查，认真摸清家底，投其所好，让鸟儿在朱湖湿地过一个安全、温馨的春节。\\n \\n朱湖湿地地处东亚—澳大利亚候鸟迁徙带，鸟类资源十分丰富，优良的生态环境成为鸟类安家落户和过境栖息的乐园。在这里发现的鸟类数量达194种，占我国鸟类总数量1371种的14%。常年栖息着各种野生鸟类5万多只，其中，列入国家二级以上重点保护鸟类有中华锦鸡、火烈斑鸠、水草鸡等10种。\\n目前，朱湖湿地鸟类主要由以下种类组成。雀形目61种，占总数的47.3%；鸻形目16种，占总数的12.4%；雁形目9种，占总数的7%；鹳形目8种，占总数的6.2%；隼形目7种，占总数的5.4%；鹤形目和佛法僧目均为5种，占总数的3.9%；鹃形目4种，占总数的3.1%；鸽形目和鸮形目均为3种，占总数的2.3%；鸊鹈目、夜鹰目、雨燕目各为1种，各占0.7%；其它。\\n \\n保护鸟类就是保护我们共同的美好家园。朱湖国家湿地公园不断加大对野生鸟类的保护力度，2018年，专门出台了《湖北孝感朱湖国家湿地公园野生动物保护制度》， 严禁任何组织和个人以收容救护野生动物为名买卖野生动物及其制品，严禁在湿地公园区域内非法狩猎、捕捞、张网捕鸟、售卖野生动物。湿地管理处成立专门巡查队伍，全天候严格防控伤害野生动物的行为。\\n同时，加大鸟类保护宣传教育工作力度。每年4月开展“国际爱鸟周”活动，专门组织周边学生游湿地公园、看鸟类生境，亲身感受生态怡人的鸟类家园，邀请野生鸟类保护专家在湿地现场上爱鸟护鸟课，并高调开展收缴、救助的野生鸟类放飞活动，形成浓厚的鸟类保护氛围。\\n \\n朱湖还立足于湿地保护和生态修复，通过挖掘古代圩田文化，顺应自然开展退田还湖，使湿地公园回归水乡古境，再现阡陌纵横。芡实、水葱、莼菜等10多种一度在朱湖区域销声匿迹的野生植物，青头潜鸭、小天鹅、白额雁、苍鹭等20多种数十年不见的野生鸟类再现河湖渠汊，生机盎然。“才闻鸟鸣在东陌，又见野莲卧菱湖”的生态美景，在这里悠然舒展。（文图供稿  万清平）\\n"
    },
    {
      "newsname": "水鸟对气候变化的反应与其首选的越冬栖息地有关",
      "newsurl": "http://www.shidi.org/sf_E845C488C8CD4C2C98DA3E9A9B7A8A9C_151_66FA58E1101.html",
      "Newstime": "2019/1/6 20:04:38",
      "Author": "媒体：sciencedaily  作者：sciencedai",
      "Body": "一篇新的科学文章表明，25种欧洲水鸟物种可以根据冬季天气改变其越冬地区。温暖的冬天允许他们将他们的越冬区域向东北方向移动，而寒冷的法术将鸟类向西南移动。在深水区越冬的物种表现出最快的长期变化：在过去24年中，它们的丰度每年向东北移动约5公里。\\n21个欧洲国家最近的一项合作研究为水鸟如何应对大规模冬季天气条件的变化提供了新的见解。该研究表明，水鸟对冬季天气条件的年度和长期变化都有反应，冬季当地丰度的变化就是明证。\\n“我们的研究强调并非所有的水鸟都对天气条件的变化做出同样的反应。喜欢浅水和深水的物种对温度的年变化反应最快，而像鹅一样的农田物种反应很弱，”赫尔辛基的DiegoPavón-Jordán说。芬兰自然历史博物馆鸟类学实验室，本研究的主要作者。\\n除了年度变化之外，该研究还表明，在20世纪90年代和21世纪90年代和20世纪90年代逐渐向东北移动的深海水域的越冬种群中心有一个长期的变化。在20世纪90年代和21世纪初期，偏爱浅水的物种的中心向东北方向移动，而在2000年代中期以后向西南移动，与欧洲连续几个严寒的冬季相吻合。\\n“根据2018年10月发布的IPCC最新报告，冬季将在不久的将来变得温和，这肯定会影响整个欧洲水鸟的丰富程度。许多物种分布的南部边缘的一些湿地可能会出现局部灭绝芬兰自然历史博物馆赫尔辛基鸟类学实验室负责人Aleksi Lehikoinen说，该分布的北部边缘有新的湿地的殖民地。\\nIPCC报告的冬季天气条件的变化增加也可能导致大量的逐年波动，沿着迁徙飞行路线向北和向南推动和拉动个体。所有这些分布区域的变化和水鸟的丰富度都为保护和监测物种带来了挑战。例如，由于气候条件可能在该地区变得不利，因此某些保护区内的物种可能不再冬天。\\n"
    }
  ]
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