package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/pkg/sftp"
	sftpsender "github.com/yudgxe/beans-go-test"
	"github.com/yudgxe/beans-go-test/model"
	"golang.org/x/crypto/ssh"
)

var (
	addr     string
	user     string
	password string
)

func init() {
	flag.StringVar(&addr, "a", "localhost:22", "Адрес сервера")
	flag.StringVar(&user, "u", "dev", "Имя пользователя")
	flag.StringVar(&password, "p", "123", "Пароль пользователя")

	flag.Parse()
}

func main() {
	if len(os.Args) < 3*flag.NFlag()*2 {
		fmt.Println("Недостаточно аргументов")
		return
	}
	config := ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", addr, &config)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	sc, err := sftp.NewClient(conn)
	if err != nil {
		log.Fatal(err)
	}

	subcommand := os.Args[1+flag.NFlag()*2]
	jsonFile := os.Args[2+flag.NFlag()*2]

	jsonData, err := os.ReadFile(jsonFile)
	if err != nil {
		log.Fatal(err)
	}

	switch subcommand {
	case "create":
		var create model.Create
		if err := json.Unmarshal(jsonData, &create); err != nil {
			log.Fatal(err)
		}
		if err := sftpsender.Create(sc, create); err != nil {
			log.Fatal(err)
		}
	case "update":
		var update model.Update
		if err := json.Unmarshal(jsonData, &update); err != nil {
			log.Fatal(err)
		}
		if err := sftpsender.Update(sc, update); err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}
}
