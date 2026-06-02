package main

import (
    "fmt"
    "net/smtp"
)

func main() {
    auth := smtp.PlainAuth("", "csin45159@gmail.com", "dluikppwjpwhzyvx", "smtp.gmail.com")
    err := smtp.SendMail(
        "smtp.gmail.com:587",
        auth,
        "csin45159@gmail.com",
        []string{"ronexlemon@gmail.com"},
        []byte("Subject: Test\r\n\r\nTest email"),
    )
    if err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println("Sent!")
    }
}