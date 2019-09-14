package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api" //get -u github.com/go-telegram-bot-api/telegram-bot-api
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var (
	// глобальная переменная в которой храним токен
	telegramBotToken string
	settingsfile string = "settings.json"
	//settingsmap map[string]interface{} //карта, в которую считываеются все настройки
	settingsmap map[string]interface{} //карта, в которую считываеются все настройки

)

func init() {
	// принимаем на входе флаг -telegrambottoken
	flag.StringVar(&telegramBotToken, "telegrambottoken", "", "Telegram Bot Token")
	flag.Parse()

	// без него не запускаемся
	if telegramBotToken == "" {
		log.Print("-telegrambottoken is required")
		os.Exit(1)
	}
}

func main() {
	settinsreading()
	// используя токен создаем новый инстанс бота
	bot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)
	// u - структура с конфигом для получения апдейтов
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// используя конфиг u создаем канал в который будут прилетать новые сообщения
	updates, err := bot.GetUpdatesChan(u)
	// в канал updates прилетают структуры типа Update
	// вычитываем их и обрабатываем
	for update := range updates {
		readingmessage_func(update, bot)
	}
}

func settinsreading(){
	//r1, _ := regexp.Compile("(^.*)=")
	// read file
	data, err := ioutil.ReadFile(settingsfile)
	if err != nil {
		fmt.Print(err)
	}
	if err := json.Unmarshal(data, &settingsmap); err != nil {
		//panic(err)
		log.Println(err)
	}

	fmt.Println(settingsmap)
	allowedids := settingsmap["allowedids"].([]interface{})
	fmt.Printf("%s\n", "Allowed IDs:")
	for k , id := range allowedids{
		fmt.Printf("%d: %s\n", k, id)
	}
}

func readingmessage_func(update tgbotapi.Update, bot *tgbotapi.BotAPI) {

	// универсальный ответ на любое сообщение
	var reply string
	//	if update.Message == nil {
	//		return
	//	}

	//regexp_func(update.Message.Text)
	// логируем от кого какое сообщение пришло
	log.Printf("[%s:%s] %s", update.Message.From.UserName, strconv.Itoa(update.Message.From.ID), update.Message.Text)
	// свитч на обработку комманд
	// комманда - сообщение, начинающееся с "/"
	switch update.Message.Command() {
	case "start":
		reply = "Привет. Я телеграм-бот"
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		bot.Send(msg)
		//	case "hello":
		//		reply = "world"
	}

	allowedids := settingsmap["allowedids"].([]interface{}) //.([]interface{}
	allowed := false

	for _ , id := range allowedids {
		chatidstr := strconv.FormatInt(update.Message.Chat.ID,10)
		//fmt.Println(chatidstr)
		if chatidstr == id{
			allowed = true
			fmt.Printf("%s разрешен\n", id)
		}
	}

	//fmt.Println(allowed, update.Message.Chat.ID)
	if allowed{ //если id в списке разрешенных, выполняем код:

		//reply = "О, я тебя знаю!"
		//usermsg[] string := update.Message.Text
		regexp_result := regexp_func(update.Message.Text)
		reply = "Результат от регулярки " + settingsmap["userregexp"].(string) + ": " + regexp_result
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		bot.Send(msg)

		usercommand_func(regexp_result)



	} else { //если id не нашелся в списке разрешенных, выполняем это:
		//fmt.Println("В настройках нет такого ID")
		reply = "Мы не знакомы :("
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		bot.Send(msg)
	}
}

func regexp_func(inputtext string) string{
	//var pattern string
	//var text string
	//text := inputtext
	userregexp := settingsmap["userregexp"].(string)
	r, _ := regexp.Compile(userregexp)
	match := r.FindAllString(inputtext, -1)
	output := strings.Join(match, "")
	if output == ""{
		output = "null"
	}
	return output
}

func usercommand_func(inputtext string) {

	commandstr := strings.Replace(settingsmap["usercommand"].(string),"<arg>", inputtext, -1)
	cmd := exec.Command("/bin/bash", "-c", commandstr)
	stdout, err := cmd.Output()
	if err != nil {
		println(err.Error())
		return
	}

	//s := string([]byte(stdout))
	fmt.Printf("%s",stdout)
}