package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"database/sql"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dbprepare()
	log.Printf("\x1b[31mНачали работать, проверяем сайт \x1b[32m%s\x1b[0m", site)
	go Request(site)
	go archive()
	TGBOTCONN()

}

func init() {
	f, err := os.Open("config.txt")
	if err != nil {
		st := "Не найден файл config.txt. Cейчас создан файл \"default config.txt\". Отредактируйте по своим параметрам, переименуйте в config.txt и запустите бота заново!!!"
		st1 := "\x1b[31mНе найден файл config.txt. Cейчас создан файл \"default config.txt\". Отредактируйте по своим параметрам, переименуйте в config.txt и запустите бота заново!!!\x1b[0m"
		writelog(st)
		log.Print(st1)
		cr, _ := os.Create("default config.txt")
		cr.WriteString("token=TELEGRAMTOKEN it can be created in @BotFather;\n")
		cr.WriteString("site=http://localhost;\n")
		cr.WriteString("period=5;\n")
		cr.WriteString("teststring=Пароль;")
		cr.Close()
		est := err.Error()
		writelog(est)
		os.Exit(0)

	}

	wr := bytes.Buffer{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		wr.WriteString(sc.Text())
	}
	site = (wr.String())
	f.Close()
	var config []string = strings.Split(site, ";")
	token = strings.TrimPrefix(config[0], "token=")
	site = strings.TrimPrefix(config[1], "site=")
	p := strings.TrimPrefix(config[2], "period=")
	per, _ := strconv.Atoi(p)
	period = int64(per)
	teststring = strings.TrimPrefix(config[3], "teststring=")
	mode = true
	log.Printf("\x1b[34m\nТокен: %s\nСайт: %s\nПроверяемая строка: %s\nПериод запрашивания в секундах: %d\x1b[0m", token, site, teststring, period)
}

func dbOpen(query string) {
	database, _ := sql.Open("sqlite3", "./chatid.db")
	statement, _ := database.Prepare(query)
	statement.Exec()
	statement.Close()
	database.Close()
}

func dbprepare() {
	str := "\x1b[34mПодготавливаем БД к работе, если нет-создаем\x1b[0m"
	log.Println(str)
	//database, _ := sql.Open("sqlite3", "./chatid.db")
	//statement3, _ := database.Prepare("CREATE TABLE IF NOT EXISTS chat (id INTEGER PRIMARY KEY)")
	dbOpen("CREATE TABLE IF NOT EXISTS chat (id INTEGER PRIMARY KEY)")
	//statement3.Exec()
	//statement3.Close()
	//database.Close()
}

var site, token, teststring string
var mode bool
var period int64

func Request(site string) {
	for {
		var str = [1]string{}
		time.Sleep(time.Second * time.Duration(period))
		if mode == true {
			log.Printf("\x1b[34mОтправляем GET запрос на %s\x1b[0m", site)
			client := &http.Client{}
			req, _ := http.NewRequest("GET", site, nil)
			req.Header.Add("User-Agent", "TEnergoOnlinebot")
			resp, err := client.Do(req)
			client.CloseIdleConnections()
			st := "отправили GET запрос на " + site
			writelog(st)
			if err != nil {
				log.Println("\x1b[34mСайт не отвечает\x1b[0m")
				database, _ := sql.Open("sqlite3", "./chatid.db")
				rows, _ := database.Query("select id from chat")
				for rows.Next() {
					var id int64
					err = rows.Scan(&id)
					if err != nil {
						est := err.Error()
						writelog(est)
						os.Exit(1)
					}
					newConnect, err := tgbotapi.NewBotAPI(token)
					if err != nil {
						writelog("Ошибка соединения с ботом, что то не так. Cмотреть ошибку")
						est := err.Error()
						writelog(est)
					}
					newConnect.Debug = false
					mestext := site + " нет ответа, сайт недоступен"
					msg := tgbotapi.NewMessage(id, mestext)
					newConnect.Send(msg)

				}
				rows.Close()
				st1 := site + " нет ответа, сайт недоступен"
				writelog(st1)
				continue
			}
			if resp.StatusCode == 200 {
				b, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				str = [1]string{(string(b))}
				if strings.Contains(str[0], teststring) == true {
					resp.Body.Close()
					continue
				} else {
					database, _ := sql.Open("sqlite3", "./chatid.db")
					rows, _ := database.Query("select id from chat")
					for rows.Next() {
						var id int64
						err = rows.Scan(&id)
						if err != nil {
							log.Println("ошибка при попытке запросить данные в БД")
							writelog("ошибка при попытке запросить данные в БД")
							est := err.Error()
							writelog(est)
						}
						newConnect, _ := tgbotapi.NewBotAPI(token)
						newConnect.Debug = false
						mestext := site + " что-то не так"
						msg := tgbotapi.NewMessage(id, mestext)
						newConnect.Send(msg)

					}
					database.Close()
					resp.Body.Close()
					client.CloseIdleConnections()
					st1 := site + " что-то не так"
					writelog(st1)
				}
			} else {
				database, _ := sql.Open("sqlite3", "./chatid.db")
				rows, _ := database.Query("select id from chat")
				for rows.Next() {
					var id int64
					err = rows.Scan(&id)
					if err != nil {
						log.Println("ERROR")
						est := err.Error()
						writelog(est)
					}
					newConnect, _ := tgbotapi.NewBotAPI(token)
					mestext := site + " Ответ сайта неправильный. " + resp.Status
					msg := tgbotapi.NewMessage(id, mestext)
					newConnect.Send(msg)
				}
				rows.Close()
				st1 := site + " Ответ сайта неправильный. " + resp.Status
				writelog(st1)
				database.Close()
				resp.Body.Close()
			}
		} else {
			continue
		}
	}
}
func TGBOTCONN() {
	newConnect, _ := tgbotapi.NewBotAPI(token)
	str := "Авторизовался в боте " + newConnect.Self.UserName + ". Начали обновлять сообщения."
	log.Println("\x1b[32m" + str + "\x1b[0m")
	writelog(str)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := newConnect.GetUpdatesChan(u)
	time.Sleep(time.Millisecond * 500)
	updates.Clear()
	for update := range updates {
		if update.Message == nil {
			continue
		}
		database, _ := sql.Open("sqlite3", "./chatid.db")
		chatid := update.Message.Chat.ID

		statement, _ := database.Prepare("INSERT INTO chat (id) VALUES (?) ")
		statement.Exec((chatid))
		statement.Close()
		database.Close()
		switch update.Message.Text {
		case "TGBOTSTOP":
			database, _ := sql.Open("sqlite3", "./chatid.db")
			rows, _ := database.Query("select id from chat")
			for rows.Next() {
				var id int64
				err := rows.Scan(&id)
				if err != nil {
					log.Println("Ошибка запроса в базе данных")
					est := err.Error()
					writelog(est)
				}
				mestext := "Остановили проверку по команде от ID " + strconv.FormatInt(chatid, 10) + " пользователь: " + update.Message.From.UserName
				msg := tgbotapi.NewMessage(id, mestext)
				writelog(mestext)
				newConnect.Send(msg)
				rows.Close()
				database.Close()
			}
			st := "Остановили проверку по команде от ID " + strconv.FormatInt(chatid, 10) + " пользователь: " + update.Message.From.UserName
			writelog(st)
			os.Exit(1)
		case "wait":
			mode = false
			writelog("Бот переведен в ждущий режим")
			log.Println("\x1b[32mБот переведен в ждущий режим\x1b[0m")
		case "go":
			mode = true
			writelog("Бот переведен в рабочий режим")
			log.Println("\x1b[32mБот переведен в рабочий режим\x1b[0m")
		case "status":
			if mode == true {
				mestext := "Бот работает, проверяет сайт " + site
				msg := tgbotapi.NewMessage(chatid, mestext)
				writelog(mestext + ". ID: " + strconv.FormatInt(chatid, 10))
				newConnect.Send(msg)
			} else {
				mestext := "Бот в режиме ожидания. " + site + " не проверяем"
				writelog(mestext + ". ID: " + strconv.FormatInt(chatid, 10))
				msg := tgbotapi.NewMessage(chatid, mestext)
				newConnect.Send(msg)
			}

		default:
			log.Printf("Записали id %d в базу, если еще не записано", chatid)
			mestext := "@" + update.Message.From.UserName + ", записали id если он не записан. Бот работает"
			msg := tgbotapi.NewMessage(chatid, mestext)
			newConnect.Send(msg)
			str := "Записали id: " + strconv.Itoa(int(chatid))
			writelog(str)
		}
	}
}

func writelog(str string) {
	str1 := time.Now().Format(time.DateTime) + "\t" + str
	filename := time.Now().Format(time.DateOnly) + " " + "bot.log"
	of, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		cr, _ := os.Create(filename)
		cr.WriteString(str1)
		cr.WriteString("\n")
		cr.Close()
	}

	of.WriteString(str1)
	of.WriteString("\n")
	of.Close()
}

func archive() {
	for {
		f, err := filepath.Glob("*.log")
		if err != nil {
			log.Println("нет таких файлов")
			est := err.Error()
			writelog(est)
		}

		for _, name := range f {
			if name != time.Now().Format(time.DateOnly)+" "+"bot.log" {
				a, _ := os.ReadFile(name)
				var buff bytes.Buffer
				zipW := zip.NewWriter(&buff)
				f, err := zipW.Create(name)
				if err != nil {
					est := err.Error()
					writelog(est)
					panic(err)
				}
				_, err = f.Write([]byte(a))
				if err != nil {
					est := err.Error()
					writelog(est)
					panic(err)
				}
				err = zipW.Close()
				if err != nil {
					est := err.Error()
					writelog(est)
					panic(err)
				}

				err = os.WriteFile("./archive/"+name+".zip", buff.Bytes(), os.ModePerm)
				if err != nil {
					os.Mkdir("./archive/", 0777)

					err = os.WriteFile("./archive/"+name+".zip", buff.Bytes(), os.ModePerm)
					os.Clearenv()
				}
				zipW.Close()

				os.Remove(name)
				os.Clearenv()
			}

		}
		os.Clearenv()
		time.Sleep(time.Second * 1)
		continue
	}
}
