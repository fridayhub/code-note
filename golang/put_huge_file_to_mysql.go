package main

import (
	"bufio"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

const head = "INSERT INTO db.table (a,b,c,d,e,v,f) VALUES "

var (
	valueChan  chan string
	concurChan chan struct{}
	myConn     *sql.DB
)

func initDb() {
	myConn, _ = sql.Open("mysql", "root:xxxx@tcp(xx.xx.xx.xxx:3306)/?parseTime=true&charset=utf8mb4,utf8")
	myConn.SetMaxIdleConns(1024)
	myConn.SetMaxOpenConns(1024)
	if err := myConn.Ping(); err != nil {
		log.Printf("Db init err:%v\n", err)
		os.Exit(-1)
	}
}

func saveToDb(done chan struct{}) {
	var (
		sqlss strings.Builder
		n     int
	)
	n = 0
	for value := range valueChan {
		sqlss.WriteString(value)
		sqlss.WriteString(",")
		n++
		if n == 2000 {
			sqll := strings.TrimRight(sqlss.String(), ",")
			sqlss.Reset()
			concurChan <- struct{}{}
			n = 0
			go func(sqls string) {
				defer func() {
					<-concurChan
				}()
				//fmt.Println(head + sqls)
				_, err := myConn.Exec(head + sqls)
				handleErr("inser error:", err)
			}(sqll)
		}
	}
	done <- struct{}{}
}

func ParseLine(line string) {
	//fmt.Println("ininin..:", []byte(line))
	lineItem := strings.Split(line, "\r")
	var (
		value strings.Builder
		n     int
		item  string
	)

	value.WriteString("(")
	n = 0
	for n, item = range lineItem {
		if n > 7 {
			break
		}

		if strings.Index(item, "\\") != -1 {
			item = strings.ReplaceAll(item, "\\", "\\\\'")
		}
		if strings.Index(item, "'") != -1 {
			item = strings.ReplaceAll(item, "'", "\\'")
		}

		//fmt.Printf("co:%d, content:%s\n", n, item)
		value.WriteString("'")

		value.WriteString(item)
		value.WriteString("',")
	}

	for i := n; i < 7; i++ {
		value.WriteString("'")
		value.WriteString("',")
	}
	sqlss := strings.TrimRight(value.String(), ",") + ")"
	valueChan <- sqlss
}

func main() {
	start := time.Now().Unix()
	fmt.Println(time.Now().Format("2006-01-02 15:04:05")) //2019-11-17 20:20:41
	valueChan = make(chan string, 1000)
	concurChan = make(chan struct{}, 1000)
	initDb()

	done := make(chan struct{})
	go saveToDb(done)

	fd, err := os.Open("d:\\u3.txt")
	handleErr("open file", err)
	buf := bufio.NewReader(fd)
	var n uint64
	n = 0
	for {
		line, err := buf.ReadString('\n')
		n++
		if err != nil {
			if err == io.EOF {
				//end := time.Now().Unix()
				//fmt.Println(time.Now().Format("2006-01-02 15:04:05")) //2019-11-17 20:25:45
				//fmt.Println("use second:", end - start) //use second: 304
				//fmt.Println("read lines:", n)  //read lines: 436525670
				break
			}
			handleErr("read string:", err)
		}
		ParseLine(strings.TrimRight(line, "\n"))
		//fmt.Printf("value:%s\n",value)
	}

	<-done
	end := time.Now().Unix()
	fmt.Println(time.Now().Format("2006-01-02 15:04:05")) //2019-11-17 20:25:45
	fmt.Println("use second:", end-start)                 //use second: 304
}

func handleErr(msg string, err error) {
	if err != nil {
		log.Fatal(msg, err)
	}
}
