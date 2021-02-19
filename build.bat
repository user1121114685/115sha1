set buildtime=%date:~0,4%%date:~5,2%%date:~8,2%%time:~0,2%%time:~3,2%%time:~6,2%
echo %buildtime%>./version.txt 

go build -ldflags "-X main.Version=%buildtime%"


zip -q 115sha1_64λ.zip 115sha1.exe
del 115sha1.exe
rmdir /s/q logs