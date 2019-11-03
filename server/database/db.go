package database

import (
	"github.com/go-sql-driver/mysql"
	"github.com/hackbeex/configcenter/local"
	"github.com/hackbeex/configcenter/util/log"
	"github.com/jinzhu/gorm"
	"os"
	"time"
)

var dbConn *gorm.DB

func init() {
	var err error
	var dbDebugMode = true
	if os.Getenv("DB_DEBUG") == "0" {
		dbDebugMode = false
	}

	dbConf := dbConfig()
	dbConn, err = gorm.Open("mysql", dbConf.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	dbConn.SingularTable(true)
	dbConn.LogMode(dbDebugMode)
}

func dbConfig() *mysql.Config {
	c := mysql.NewConfig()
	c.Net = "tcp"
	c.Collation = "utf8mb4_general_ci"
	c.Loc = time.Local
	c.MaxAllowedPacket = 20 << 20
	c.ParseTime = true
	c.Timeout = time.Second * 1
	c.ReadTimeout = time.Second * 2
	c.WriteTimeout = time.Second * 2

	conf := local.Conf.Server.Mysql
	c.User = conf.User
	c.Passwd = conf.Password
	c.Addr = conf.Addr
	c.DBName = conf.DBName
	return c
}

func Conn() *gorm.DB {
	return dbConn.New()
}
