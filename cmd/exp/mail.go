package main

import (
	"fmt"
	"net/smtp"
	"os"
)

func main() {
	// Gmailアカウントのメールアドレスとパスワード
	from := "foomittu@gmail.com"
	password := "your_password"
	to := "iza03275@nifty.com.com"
	subject := "GolangからGmail送信テスト"
	body := "Golangから送信されたテストメールです。"

	// SMTPサーバーの設定
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// 認証情報を作成
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// メールメッセージを作成
	msg := []byte(fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\n\r\n%s\r\n", to, from, subject, body))

	// メールを送信
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, msg)
	if err != nil {
		fmt.Println("メール送信失敗:", err)
		os.Exit(1)
	}

	fmt.Println("メール送信成功！")
}