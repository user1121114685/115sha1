package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/flopp/go-findfont"
	"github.com/gawwo/fake115-go/config"
	"github.com/gawwo/fake115-go/core"
	"github.com/gogf/gf/text/gregex"

	//"github.com/gawwo/fake115-go/log"
	"crypto/sha1"

	"github.com/tidwall/gjson"
)

func init() {

	//获取中文字体列表
	fontPaths := findfont.List()
	for _, path := range fontPaths {
		//设置字体
		if strings.Contains(path, "simkai.ttf") {
			os.Setenv("FYNE_FONT", path)
			break
		}
	}
}

var Version = ""

var MyFolderChoose string             // 最终选择的目录
var MyFolderChooseName string         // 最终选择的目录
var JsonData string                   //JSON 目录
var CooKie_115 string                 //115 Cookie
var ChooseFolder_115 []string         //115 文件目录
var ChooseFolderMap map[string]string /*创建集合 */
var NewVersion string

func getcookie() {
	dir, err := ioutil.TempDir("", "chromedp-example")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", false),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("window-size", "1280,720"),
		chromedp.UserDataDir(dir),
	)

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	//defer cancel()

	// also set up a custom logger
	taskCtx, _ := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	//defer cancel()
	chromedp.ListenTarget(taskCtx, func(ev interface{}) {
		switch ev := ev.(type) {

		case *network.EventResponseReceived:
			resp := ev.Response
			if len(resp.Headers) != 0 {
				// log.Printf("received headers: %s", resp.Headers)
				//https://115.com/?cid=0&offset=0&tab=&mode=wangpan
				//https://115.com/?cid=    &offset=0&mode=wangpan
				// https://webapi.115.com/files?aid=1&cid=   &o=user_ptime&asc=0&offset=0&show_dir=1&limit=40&code=&scid=&snap=0&natsort=1&record_open_time=1&source=&format=json&type=&star=&is_share=&suffix=&custom_order=&fc_mix=
				if strings.Index(resp.URL, "https://115.com/?cid=") != -1 {
					fmt.Println("找到API啦！！  " + resp.URL)

					respURL, err := gregex.MatchString(`cid=\d+`, resp.URL)
					respCid, err := gregex.MatchString(`\d+`, respURL[0])
					if err == nil {
						MyFolderChoose = respCid[0]
						// document.querySelector("#js_top_header_file_path_box > div.top-file-path > div")
						choose115FolderName()
						fmt.Println("选择了   " + MyFolderChooseName + MyFolderChoose)
						//choose115FolderName()

					} else {
						fmt.Println("API 提取错误。。 请GitHub联系 " + resp.URL)

						var exitScan string
						_, _ = fmt.Scan(&exitScan)
					}

				}
			}

		}
		// other needed network Event
	})
	chromedp.Run(taskCtx,
		network.Enable(),
		chromedp.Navigate(`https://115.com/`),
		chromedp.WaitVisible(`#js-main_leftUI`, chromedp.BySearch),
		chromedp.Navigate(`https://webapi.115.com/history/receive_list`),
		chromedp.ActionFunc(func(ctx context.Context) error {
			cookies, err := network.GetCookies().Do(ctx)
			if err != nil {
				return err
			}

			//var coo string
			for _, v := range cookies {
				CooKie_115 = CooKie_115 + v.Name + "=" + v.Value + ";"
			}
			// // 将保存的字符串转换为字节流
			// str := []byte(coo)

			// // 保存到文件
			// ioutil.WriteFile(`cookies.txt`, str, 0775)

			return nil
		}),
		chromedp.Navigate(`https://115.com/`),
	)

	//choose115Folder()

}

func choose115FolderName() {
	ChooseFolderMap = make(map[string]string)
	client := &http.Client{}
	urlstring := "https://webapi.115.com/files?aid=1&cid=" + MyFolderChoose + "&o=user_ptime&asc=0&offset=0&show_dir=1&limit=40&code=&scid=&snap=0&natsort=1&record_open_time=1&source=&format=json"
	// https://webapi.115.com/files?aid=1&cid=    &o=user_ptime&asc=0&offset=0&show_dir=1&limit=40&code=&scid=&snap=0&natsort=1&record_open_time=1&source=&format=json
	reqest, err := http.NewRequest("GET", urlstring, nil) //建立一个请求
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(0)
	}
	//Add 头协议
	reqest.Header.Add("Accept", "*/*")
	reqest.Header.Add("Accept-Language", "ja,zh-CN;q=0.8,zh;q=0.6")
	reqest.Header.Add("Connection", "keep-alive")
	reqest.Header.Add("Cookie", CooKie_115)
	reqest.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.146 Safari/537.36")
	response, err := client.Do(reqest) //提交

	defer response.Body.Close()
	body, err1 := ioutil.ReadAll(response.Body)
	if err1 != nil {
		// handle error
	}
	fmt.Println(string(body)) //网页源码
	//gjson.GetMany(string(body))
	//gjson.Parse(json).Get("name").Get("last")
	//gjson.Get(string(body), "data.data.cid").Get("ns")
	result := gjson.Get(string(body), "path.#.name")
	MyFolderChooseName = "\r\n"
	for _, pathname := range result.Array() {
		println(pathname.String())
		MyFolderChooseName = MyFolderChooseName + "/" + pathname.String()

	}
	MyFolderChooseName = MyFolderChooseName + "\r\n"

}

func inputBoxChoose() {

	fdw := fyne.CurrentApp().NewWindow("导入功能")
	//	窗口大小
	fdw.Resize(fyne.NewSize(800, 400))
	//selectEntry := widget.NewSelectEntry(ChooseFolder_115)
	//selectEntry.PlaceHolder = "请输入CID或者选择文件夹"
	fdw.SetContent(container.NewVBox(
		// 选择需要导入的目录
		widget.NewLabel("选择导入目录请在Chrome浏览器上进行\r\n按f5刷新，确认导入目录"),

		//content := fdw.SetContent(container.NewVBox(
		widget.NewButton("选择需要导入的Json文件", func() {

			fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
				if err == nil && reader == nil {
					return
				}

				if err != nil {
					dialog.ShowError(err, fdw)
					return
				}
				JsonData = reader.URI().Path()
				//imageOpened(reader)
			}, fdw)
			fd.Resize(fyne.NewSize(800, 400))

			fd.SetFilter(storage.NewExtensionFileFilter([]string{".json", ".txt"}))

			fd.Show()

		}),

		//开始导入
		widget.NewButton("开始导入", func() {
			if JsonData == "" {
				err := errors.New("请先选择 需要导入的Json文件")
				dialog.ShowError(err, fdw)
			} else if MyFolderChoose == "" {
				err := errors.New("请先选择 需要导入到115的目录")
				dialog.ShowError(err, fdw)
			} else {

				_, err := os.Stat("./已导入sha1.txt") //os.Stat获取文件信息
				if err != nil {
					ioutil.WriteFile("./已导入sha1.txt", nil, 0777)
				}
				file, err := os.Open("./已导入sha1.txt")

				if err != nil {
					println(err.Error())

				}
				defer file.Close()
				// 计算导入文件的sha1
				f, err := os.Open(JsonData)
				if err != nil {
					log.Fatal(err)
				}
				defer f.Close()
				h1 := sha1.New()
				if _, err := io.Copy(h1, f); err != nil {
					log.Fatal(err)
				}

				ScannerSha1 := bufio.NewScanner(file)
				var WriteToSha1 bool
				WriteToSha1 = true
				for ScannerSha1.Scan() {

					sha1s := ScannerSha1.Text()
					if fmt.Sprintf("%x", h1.Sum(nil)) == sha1s {

						WriteToSha1 = false
					}

				}
				// dir, _ := os.Getwd()

				//exec.Command("cmd", "/c", "start cmd /k"+dir+"\\fake115.exe -c \""+CooKie_115+"\" "+MyFolderChoose+" "+JsonData).Run()

				//go core.Import(MyFolderChoose, JsonData)

				cnf := dialog.NewConfirm(JsonData, "确定导入到  "+MyFolderChooseName+"CID是 "+MyFolderChoose, func(s bool) {

					if s == true {
						go core.Import(MyFolderChoose, JsonData)

						if WriteToSha1 != false {
							file, err := os.OpenFile("./已导入sha1.txt", os.O_APPEND, 0777)
							defer file.Close()

							if err != nil {
								println(err.Error())

							}
							write := bufio.NewWriter(file)

							write.WriteString(fmt.Sprintf("%x\r\n", h1.Sum(nil)))

							//Flush将缓存的文件真正写入到文件中
							write.Flush()
							file.Sync()
							fmt.Println(fmt.Sprintf("%x\r\n", h1.Sum(nil)))

						}

					}
				}, fdw)
				cnf.SetDismissText("不")
				cnf.SetConfirmText("确定")
				cnf.Show()
				if WriteToSha1 == false {

					dialog.ShowInformation("发现该文件已导入", "请注意是否需要重复导入 "+JsonData, fdw)
				}
			}

		}),
	))

	fdw.Show()
	//content.Show()
}

func outputBoxChoose() {

	fdw := fyne.CurrentApp().NewWindow("导出功能")

	//	窗口大小
	fdw.Resize(fyne.NewSize(800, 400))
	fdw.SetContent(container.NewVBox(

		// 选择需要导入的目录
		widget.NewLabel("选择导出目录请在Chrome浏览器上进行\r\n按f5刷新，确认导出目录"),

		//开始导入
		widget.NewButton("开始导出", func() {
			if MyFolderChoose == "" {
				err := errors.New("请先选择 需要导出的115目录")
				dialog.ShowError(err, fdw)
			} else {
				//dir, _ := os.Getwd()
				cnf := dialog.NewConfirm("导出文件夹确认", "确定导出  "+MyFolderChooseName+"CID是 "+MyFolderChoose, func(s bool) {
					if s == true {
						go core.Export(MyFolderChoose)
					}
				}, fdw)
				cnf.SetDismissText("不")
				cnf.SetConfirmText("确定")
				cnf.Show()
				//exec.Command("cmd", "/c", "start cmd /k"+dir+"\\fake115.exe -c \""+CooKie_115+"\" "+MyFolderChoose).Run()

			}

		}),
	))

	fdw.Show()
	//content.Show()
}

func parseURL(urlStr string) *url.URL {
	link, err := url.Parse(urlStr)
	if err != nil {
		fyne.LogError("Could not parse URL", err)
	}

	return link
}

func QRcode() {
	fdw := fyne.CurrentApp().NewWindow("给开源一份阳光")
	//	窗口大小
	//fdw.Resize(fyne.NewSize(800, 800))  storage.NewURI("https://gitee.com/shaoxia1991/Blog/raw/master/me/%E6%94%AF%E4%BB%98%E5%AE%9D%E6%94%B6%E6%AC%BE.jpg")
	//
	aliQRcode, _ := http.Get("https://gitee.com/shaoxia1991/Blog/raw/master/me/%E6%94%AF%E4%BB%98%E5%AE%9D%E6%94%B6%E6%AC%BE.jpg")
	defer aliQRcode.Body.Close()
	weChatQRcode, _ := http.Get("https://gitee.com/shaoxia1991/Blog/raw/master/me/%E5%BE%AE%E4%BF%A1%E6%94%B6%E6%AC%BE.png")
	defer weChatQRcode.Body.Close()
	f_ali, err := os.Create("./ali.jpg")
	if err != nil {
		panic(err)
	}
	f_wechat, err := os.Create("./wechat.png")
	if err != nil {
		panic(err)
	}
	io.Copy(f_ali, aliQRcode.Body)
	io.Copy(f_wechat, weChatQRcode.Body)
	// // 将保存的字符串转换为字节流
	// str := []byte(coo)

	// // 保存到文件
	// ioutil.WriteFile(`cookies.txt`, str, 0775)
	img1 := canvas.NewImageFromFile("./ali.jpg")
	img1.FillMode = canvas.ImageFillOriginal
	img2 := canvas.NewImageFromFile("./wechat.png")
	img2.FillMode = canvas.ImageFillOriginal

	container := fyne.NewContainerWithLayout(
		layout.NewGridWrapLayout(fyne.NewSize(300, 300)),
		img1, img2)
	fdw.SetContent(container)
	fdw.Show()
}

func checkNewVersion() {
	//https://shaoxia1991.coding.net/p/115sha1/d/115sha1/git/raw/main/
	//https://raw.githubusercontent.com/user1121114685/115sha1/main/main.go
	newVersion, err := http.Get("https://shaoxia1991.coding.net/p/115sha1/d/115sha1/git/raw/main/version.txt?download=true")

	if err != nil {
		fmt.Println("获取新版本失败  ....")
		// handle error
	}
	defer newVersion.Body.Close()
	body, err := ioutil.ReadAll(newVersion.Body)
	if err != nil {
		fmt.Println("获取新版本失败  ....")
		// handle error
	}
	fmt.Println(string(body)) //网页源码
	NewVersion = string(body)
}

func fake115(w fyne.Window) {
	config.Cookie = CooKie_115
	//config.WorkerNum = 10
	// 确保cookie在登录状态
	loggedIn := core.SetUserInfoConfig()
	if !loggedIn {
		fmt.Println("Login expire or fail...")
		err := errors.New("登陆失效，或者登陆失败")
		dialog.ShowError(err, w)
		os.Exit(1)
	}

}
func main() {

	checkNewVersion()
	a := app.New()
	w := a.NewWindow("115 Sha1备份/恢复 by 联盟少侠")
	//	窗口大小
	w.Resize(fyne.NewSize(200, 600))

	hello := widget.NewLabel("这是我的第一个GUI程序，运行本程序需要安装Chrome")

	// 第一个按钮
	w.SetContent(container.NewVBox(
		hello,
		widget.NewLabel(""),
		// widget.NewLabel("版本号 2021年2月6日16:04:35"),
		// widget.NewLabel(""),
		widget.NewButton("1.登陆", func() {
			if CooKie_115 != "" {
				err := errors.New("你已经登陆了！！！  再次登陆是为了锻炼身体吗？")
				dialog.ShowError(err, w)
			} else {
				getcookie()
			}

		}),
		widget.NewButton("2.导出", func() {
			if CooKie_115 == "" {
				err := errors.New("你还没登陆，我猜你不知道需要先登陆")
				dialog.ShowError(err, w)
			} else {
				fake115(w)
				outputBoxChoose()
			}

		}),
		widget.NewLabel("导入1次后崩溃，请重新打开软件，下版本修复"),
		widget.NewButton("3.导入", func() {
			if CooKie_115 == "" {
				err := errors.New("你还没登陆，我猜你不知道需要先登陆")
				dialog.ShowError(err, w)
			} else {
				fake115(w)
				inputBoxChoose()
			}
		}),

		container.NewHBox(
			widget.NewLabel("项目地址:"),
			widget.NewHyperlink("115sha1", parseURL("https://github.com/user1121114685/115sha1")),
			widget.NewLabel(" "),
			widget.NewHyperlink("fake115", parseURL("https://github.com/gawwo/fake115-go/releases/latest")),
		),
		container.NewHBox(
			widget.NewLabel("作者博客:"),
			widget.NewHyperlink("shaoxia.xyz", parseURL("https://shaoxia.xyz/")),
			widget.NewLabel(" "),
			widget.NewHyperlink("TG分享交流群", parseURL("https://t.me/Resources115")),
		),
		container.NewHBox(
			widget.NewLabel("当前版本:"+Version),
		),
		container.NewHBox(
			widget.NewLabel("最新版本:"),
			widget.NewHyperlink(NewVersion, parseURL("https://shaoxia1991.coding.net/p/115sha1/d/115sha1/git/raw/main/115sha1_64%E4%BD%8D.zip")),
		),
		widget.NewButton("115SHA1加油", func() { QRcode() }),

		//widget.NewButton("合二为一的版本", func() { fake115(w) }),

	))

	//w.ShowAndRun()
	//第二个按钮

	w.ShowAndRun()
	os.Unsetenv("FYNE_FONT")
}

// 选择文件不好用，那就拖动输入
