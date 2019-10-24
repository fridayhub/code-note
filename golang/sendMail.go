package main

import (
	"fmt"
	"net/smtp"
	"strings"
)

func SendToMail(user, password, host, to, subject, body, mailtype string) error {
	hp := strings.Split(host, ":")
	auth := smtp.PlainAuth("", user, password, hp[0])
	var content_type string
	if mailtype == "html" {
		content_type = "Content-Type: text/" + mailtype + "; charset=UTF-8"
	} else {
		content_type = "Content-Type: text/plain" + "; charset=UTF-8"
	}

	msg := []byte("To: " + to + "\r\nFrom: " + user + ">\r\nSubject: " + subject + "\r\n" + content_type + "\r\n\r\n" + body)
	send_to := strings.Split(to, ";")
	err := smtp.SendMail(host, auth, user, send_to, msg)
	return err
}

func main() {
	user := "xxx@axxxa.com"
	password := "wwww"
	host := "smtp.exmail.qq.com:25"
	to := "xxx@qq.com"

	subject := "使用Golang发送邮件"

	body := `重置密码连接:http://www.sssxxx.com/pass/reset?thoke=xxxxwisdlfasldfowa,
              30分钟内有效`
	fmt.Println("send email")
	err := SendToMail(user, password, host, to, subject, body, "html")
	if err != nil {
		fmt.Println("Send mail error!")
		fmt.Println(err)
	} else {
		fmt.Println("Send mail success!")
	}
}
