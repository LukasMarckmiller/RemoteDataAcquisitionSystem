package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/semihalev/gin-stats"
	"os"
)

var (
	router *gin.Engine
	app    *config
)

const (
	portDefault = "8000"
)

type config struct {
	Server     string
	DeviceName string
}

func main() {
	//A unique name for the device can configured. It is used as prefix for img names
	name := os.Getenv("RFANAME")
	server := os.Getenv("RFASERVER")
	if server == "" {
		//Set default
		server = "root@192.168.0.13"
	}
	app = &config{server, name}
	gin.SetMode(gin.DebugMode)
	//Uncomment the next line if you need gin debug output like handled requests by gin
	//gin.DefaultWriter = ioutil.Discard
	//Set the router as default
	router = gin.Default()
	//Load static html content
	router.LoadHTMLFiles("web/index.html")
	router.Static("/css", "web/css")
	router.Static("/js", "web/js")
	router.Static("/fonts", "web/fonts")
	/*USE ONLY FOR LAB ENVIRONMENT, BUILD OWN CONFIG FOR PRODUCTIVE BUILD:
	https://github.com/gin-contrib/cors*/
	router.Use(cors.Default())

	router.Use(stats.RequestStats())

	// Define Simple test route
	initRoutes()

	//Hardware Rec

	/*err,disks := getDisksWithoutBootPart()
	if (err != nil){
		fmt.Printf("Err while trying to retriev Block/Disk info", err)
	}
	for _,disk := range disks{
		fmt.Printf(disk.String())l
	}
	*/
	port := os.Getenv("RFAPORT")
	if port == "" {
		port = portDefault
	}
	router.Run(":" + port)
	//router.RunTLS(":5443", "/etc/ssl/certs/server.crt","/etc/ssl/private/server.key")
}

/*
DC3DD needs to be installed on client and server

To avoid password prompting for ssh while using dd or dc3dd generate ssh key pair and share to client:

$ ssh-keygen -t rsa -b 2048
Generating public/private rsa key pair.
Enter file in which to save the key (/home/username/.ssh/id_rsa):
Enter passphrase (empty for no passphrase):
Enter same passphrase again:
Your identification has been saved in /home/username/.ssh/id_rsa.
Your public key has been saved in /home/username/.ssh/id_rsa.pub.

Copy your keys to the target server:

$ ssh-copy-id id@server
id@server's password:

Now try logging into the machine, with ssh 'id@server', and check in:

.ssh/authorized_keys

to make sure we haven’t added extra keys that you weren’t expecting.

Finally check logging in…

$ ssh id@server

id@server:~$
"
*/
