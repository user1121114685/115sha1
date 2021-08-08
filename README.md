# 115sha1
便捷的115sha1导入/导出功能。方便分享。   
## 开源说明

本项目旨在练习golang 和减少劳动力。    

本项目只是 fake115-go 的GUI包装，核心功能由fake115-go实现，对大佬进行感谢！   
[更新日志](https://github.com/user1121114685/115sha1/blob/main/update.md)  



## 使用方法

**需要Chrome**    
下载地址：https://shaoxia1991.coding.net/p/115sha1/d/115sha1/git/raw/main/115sha1_64%E4%BD%8D.zip    
使用简单方便，就不需要说明了，如果你想让该项目持续健康的发展请对我的付出进行捐赠。 

## 特别感谢
https://github.com/gawwo/fake115-go
https://github.com/getbuguai/gaihosts
---
## 如果对你有所帮助，也可以对我进行捐赠。那撒我也不废话，下面是二维码。
![微信](https://gitee.com/shaoxia1991/Blog/raw/master/me/%E5%BE%AE%E4%BF%A1%E6%94%B6%E6%AC%BE.png)  

![支付宝](https://gitee.com/shaoxia1991/Blog/raw/master/me/%E6%94%AF%E4%BB%98%E5%AE%9D%E6%94%B6%E6%AC%BE.jpg)  

---
## 编译 
<!-- `go env -w GOARCH=386` -->
`go build  -ldflags="-s -w -H windowsgui" main.go`

```
fyne package -os android -appID shaoxia.xyz.115sha1
```



## 尚未实现

1. IOS 
2. 多个json 选择

## License

[The MIT License (MIT)](https://raw.githubusercontent.com/user1121114685/115sha1/master/LICENSE)