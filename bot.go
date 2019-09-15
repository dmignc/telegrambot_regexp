package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api" //get -u github.com/go-telegram-bot-api/telegram-bot-api
)

var (
	// переменная с токеном
	telegramBotToken string
	// имя файла с настройками
	settingsfile string = "settings.json"
	//карта, в которую считываеются все настройки из файла
	settingsmap map[string]interface{}
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

	fmt.Println("platform:", runtime.GOOS)
}

func main() {
	settinsreading() //читаем настройки

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

//функци чтения настроек
func settinsreading() {
	data, err := ioutil.ReadFile(settingsfile) //читаем файл settingsfile
	if err != nil {
		fmt.Print(err)
	}
	if err := json.Unmarshal(data, &settingsmap); err != nil { //запихиваем его в settingsmap
		log.Println(err)
	}

	//fmt.Println(settingsmap)
	allowedids := settingsmap["allowedids"].([]interface{}) //получаем массив с разрешенным ID пользователей
	fmt.Printf("%s\n", "Allowed IDs:")                      //и выводим их
	for k, id := range allowedids {
		fmt.Printf("%d: %s\n", k, id)
	}
}

//функция чтения сообщения в чате
func readingmessage_func(update tgbotapi.Update, bot *tgbotapi.BotAPI) {

	var reply string //переменная с текстом ответа

	// логируем от кого какое сообщение пришло
	log.Printf("[%s:%s] %s", update.Message.From.UserName, strconv.Itoa(update.Message.From.ID), update.Message.Text)
	// свитч на обработку комманд
	// комманда - сообщение, начинающееся с "/"
	switch update.Message.Command() {
	case "start":
		reply = "Привет. Я телеграм-бот"
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		bot.Send(msg)
	}

	allowedids := settingsmap["allowedids"].([]interface{}) //помещаем список разрешенных ID в массив
	allowed := false                                        //флаг, указывающий на то, разрешен ID или нет

	for _, id := range allowedids { //перебираем массив с ID и определяем, если ли user в списке разрешенных ID
		chatidstr := strconv.FormatInt(update.Message.Chat.ID, 10)
		//fmt.Println(chatidstr)
		if chatidstr == id {
			allowed = true //если нашли ID в списке, ставим флаг
			fmt.Printf("%s разрешен\n", id)
		}
	}

	if allowed { //если ID пользователя в списке разрешенных, выполняем код:
		regexp_result := regexp_func(update.Message.Text)
		reply = "Результат от регулярки " + settingsmap["userregexp"].(string) + ": " + regexp_result
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		bot.Send(msg)

		usercommand_func(regexp_result)

	} else { //если ID не нашелся в списке разрешенных, выполняем это:
		reply = "Мы не знакомы :("
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		bot.Send(msg)
	}
}

//функция обработки регулярного выражения (сама регулярка берется из файла с настройками)
func regexp_func(inputtext string) string {
	userregexp := settingsmap["userregexp"].(string)
	r, _ := regexp.Compile(userregexp)
	match := r.FindAllString(inputtext, -1)
	output := strings.Join(match, "") //склейка всех результатов regexp
	if output == "" {
		output = "null"
	}
	return output //возвращаем результат работы функции
}

//выполнение пользовательской функции на локальной машине
func usercommand_func(inputtext string) {

	commandstr := strings.Replace(settingsmap["usercommand"].(string), "<arg>", inputtext, -1) //заменяем текст <arg> на результат нашего регулярного выражения
	//cmd := exec.Command("/bin/bash", "-c", commandstr) //формируем команду для выполнения
	cmd := exec.Command("cmd", "/C", commandstr) //формируем команду для выполнения
	stdout, err := cmd.Output()                  //выполняем команду
	if err != nil {
		println(err.Error())
		return
	}

	fmt.Printf("%s", stdout) //вывод stdout в консоль
}
