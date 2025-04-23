<div align="center">
<p>
  <a href="#">
    <img width="200" alt="Extendx" src="https://raw.githubusercontent.com/Marsudii/extendx/refs/heads/main/docs/favicon/icon.png" />
  </a>
</p>

![GitHub Release Date](https://img.shields.io/github/release-date/Marsudii/extendx)
![platform](https://img.shields.io/badge/platform-Windows%20%7C%20MacOS%20%7C%20Linux-lightgrey)
![GitHub last commit (branch)](https://img.shields.io/github/last-commit/Marsudii/extendx/main)


Stream Application Cross platforms

</div>





## Content
- [About Extendx](#About-Extendx)
- [Contribute Extendx](#Contribute)



## About Extendx
Extendx is an open-source Go (Golang) based application designed to help streamers do screen sharing between devices/platforms without the need for a capture card.
- **📡.** Capture screen from macOS, Windows, or Linux.
- **🔄.** Stream real-time to other devices over a local network (LAN).
- **💻.** Display the captured results directly in the viewer application (without a browser).
- **🚫.** No need for additional hardware such as a capture card

## Running
### MacOS
```
```

### WIN
```
Klik Run Start extend-server.exe or extend-client.exe
```



## Build MacOS(M1)
### MacOS
```
GOOS=darwin GOARCH=amd64 go build -o app_intel main.go

GOOS=darwin GOARCH=arm64 go build -o app_m1 main.go

```

## ANDROID

```
fyne package -os android -icon Icon.png -app-id com.extendx.server
```

### WIN
```
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o nama_app.exe main.go

```


## Contribute
This is an open source project. If you want to implement new features, feel free to create a pull request.