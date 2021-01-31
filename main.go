package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopxe/config"
	"gopxe/httpd"
	"gopxe/model"
	"gopxe/tftp"
	"gopxe/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io/ioutil"
	"os"
	"path"
)

func init() {
	logrus.SetReportCaller(true)
}

func startHttp(listen string) {
	if err := httpd.Server(listen); err != nil {
		logrus.Fatalln(err)
	}
}

func startTftp(context *cli.Context, listen string) {
	go func() {
		tf := tftp.NewDefaultTFTP(config.Tftp)
		defer tf.Shutdown()
		if err := tf.ListenAndServe(listen); err != nil {
			logrus.Fatalln(err)
		}
	}()
}

func initDB() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	if err := db.AutoMigrate(&model.OS{}); err != nil {
		logrus.Fatalln("auto migrate: ", err.Error())
	}
	config.Db = db
}

func initDefaultFile() {
	sys := `DEFAULT vesamenu.c32
TIMEOUT {{ timeout|default: "300" }}
MENU BACKGROUND {{ img| default: "splash.jpg" }}
MENU TITLE {{ title|default: "GoPXE" }}`

	if !utils.IsFile(path.Join(config.Tftp, "pxelinux.cfg", "default")) {
		if err := ioutil.WriteFile(path.Join(config.Tftp, "pxelinux.cfg", "default"), []byte(sys), 0644); err != nil {
			logrus.Fatalln(err)
		}
	}
}

func initDirs(dirs []string) {
	for _, dir := range dirs {
		if !utils.IsFile(dir) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				logrus.Fatalln(err)
			}
		}
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "GoPXE"
	app.Usage = ""
	app.Version = "v0.1.0"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:     "device",
			Usage:    "Service bound network card",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "work-dir",
			Usage: "Program working directory",
			Value: "/opt/GoPXE/",
		},
		&cli.IntFlag{
			Name:  "port",
			Value: 8888,
			Usage: "Http server port",
		},
	}
	app.Action = func(context *cli.Context) error {
		workDir := context.String("work-dir")
		initDirs([]string{workDir})

		if err := os.Chdir(workDir); err != nil {
			logrus.Fatalln(err)
		}

		initDB()
		initDirs([]string{path.Join(config.Tftp, "pxelinux.cfg"), config.IsoSave, config.IsoMount})
		initDefaultFile()

		config.Device = context.String("device")
		address, err := utils.GetAddress(config.Device)
		if err != nil {
			logrus.Fatalln(err)
		}
		config.Address = address
		config.Port = context.Int("port")
		startTftp(context, fmt.Sprintf("%s:%d", address, 69))
		startHttp(fmt.Sprintf("%s:%d", address, context.Int("port")))
		return nil
	}
	if err := app.Run(os.Args); err != nil {
		logrus.Fatalln(err)
	}
}
