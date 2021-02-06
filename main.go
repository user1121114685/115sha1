package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
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

var MyFolderChoose string             // 最终选择的目录
var JsonData string                   //JSON 目录
var CooKie_115 string                 //115 Cookie
var ChooseFolder_115 []string         //115 文件目录
var ChooseFolderMap map[string]string /*创建集合 */

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

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// also set up a custom logger
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

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
	)
	choose115Folder()
}

func choose115Folder() {
	ChooseFolderMap = make(map[string]string)
	client := &http.Client{}
	reqest, err := http.NewRequest("GET", "https://webapi.115.com/files?aid=1&cid=0&o=user_ptime&asc=0&offset=0&show_dir=1&limit=56&code=&scid=&snap=0&natsort=1&record_open_time=1&source=&format=json", nil) //建立一个请求
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
	result := gjson.Get(string(body), "data.#.cid")
	for _, cidname := range result.Array() {
		println(cidname.String())
		if cidname.String() != "0" {
			nameStr := "data.#(" + "cid=\"" + cidname.String() + "\").ns"
			// name := gjson.Get(json, `programmers.#(lastName="Hunter").firstName`)
			filename := gjson.Get(string(body), nameStr)
			ChooseFolderMap[filename.String()] = cidname.String() // 将结果存入数组
			println(filename.String())
			// prints "Elliotte"
			ChooseFolder_115 = append(ChooseFolder_115, filename.String()) //将 名字保存到数组
		}
	}

}
func inputBoxChoose() {

	fdw := fyne.CurrentApp().NewWindow("导入功能")
	//	窗口大小
	fdw.Resize(fyne.NewSize(800, 800))
	fdw.SetContent(container.NewVBox(
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
			fd.Resize(fyne.NewSize(800, 800))

			fd.SetFilter(storage.NewExtensionFileFilter([]string{".json", ".jpg", ".jpeg"}))

			fd.Show()

		}),

		// 选择需要导入的目录
		widget.NewLabel("选择你需要导入的目录"),
		widget.NewSelect(ChooseFolder_115, func(s string) {

			fmt.Println("selected", s, "CID是", ChooseFolderMap[s])
			MyFolderChoose = ChooseFolderMap[s]
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
				dir, _ := os.Getwd()
				//dir = strings.Replace(dir, "\\", "\\\\", -1)
				//JsonData = strings.Replace(JsonData, "\\", "\\\\", -1)
				exec.Command("cmd", "/c", "start cmd /k"+dir+"\\fake115.exe -c \""+CooKie_115+"\" "+MyFolderChoose+" "+JsonData).Run()
			}

		}),
	))

	fdw.Show()
	//content.Show()
}

func outputBoxChoose() {

	fdw := fyne.CurrentApp().NewWindow("导入功能")
	//	窗口大小
	fdw.Resize(fyne.NewSize(800, 800))
	fdw.SetContent(container.NewVBox(

		// 选择需要导入的目录
		widget.NewLabel("选择你需要导出的目录"),
		widget.NewSelect(ChooseFolder_115, func(s string) {

			fmt.Println("selected", s, "CID是", ChooseFolderMap[s])
			MyFolderChoose = ChooseFolderMap[s]
		}),
		//开始导入
		widget.NewButton("开始导出", func() {
			if MyFolderChoose == "" {
				err := errors.New("请先选择 需要导出的115目录")
				dialog.ShowError(err, fdw)
			} else {
				dir, _ := os.Getwd()

				exec.Command("cmd", "/c", "start cmd /k"+dir+"\\fake115.exe -c \""+CooKie_115+"\" "+MyFolderChoose).Run()
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
	fdw := fyne.CurrentApp().NewWindow("我不信真的有人会捐赠")
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

func main() {

	a := app.New()
	w := a.NewWindow("115 Sha1备份/恢复 by 联盟少侠")
	//	窗口大小
	w.Resize(fyne.NewSize(200, 800))

	hello := widget.NewLabel("这是我的第一个GUI程序，运行本程序需要安装Chrome")
	// 第一个按钮
	w.SetContent(container.NewVBox(
		hello,
		widget.NewLabel(""),
		widget.NewLabel("版本号 2021年2月6日16:04:35"),
		widget.NewLabel(""),
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
				outputBoxChoose()
			}

		}),

		widget.NewButton("3.导入", func() {
			if CooKie_115 == "" {
				err := errors.New("你还没登陆，我猜你不知道需要先登陆")
				dialog.ShowError(err, w)
			} else {
				inputBoxChoose()
			}
		}),

		container.NewHBox(
			widget.NewLabel("地址"),
			widget.NewHyperlink("shaoxia.xyz", parseURL("https://shaoxia.xyz/")),
			widget.NewLabel("|"),
			widget.NewHyperlink("115sha1界面", parseURL("https://github.com/user1121114685/115sha1")),
			widget.NewLabel("|"),
			widget.NewHyperlink("115sha1核心", parseURL("https://github.com/gawwo/fake115-go/releases/latest")),
		),
		widget.NewButton("没有意义的功能", func() { QRcode() }),
	))

	//w.ShowAndRun()
	//第二个按钮

	w.ShowAndRun()
	os.Unsetenv("FYNE_FONT")
}
