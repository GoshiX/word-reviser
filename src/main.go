package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	tb "gopkg.in/telebot.v4"
)

type Word struct {
	Id     int
	WordRu string
	WordEn string
	Author string
	Show   bool
	Total  int
	Corr   int
}

func Logger(next tb.HandlerFunc) tb.HandlerFunc {
	return func(c tb.Context) error {
		startTime := time.Now()

		err := next(c)
		if err != nil {
			log.Printf("@%s: %s WITH ERROR: %v", c.Sender().Username, c.Text(), err)
		} else {
			log.Printf("@%s: %s (%s)", c.Sender().Username, c.Text(), time.Since(startTime))
		}
		return err

	}
}

var (
	db               DB
	waitingTranslion map[int]string
)

func startQuiz(c tb.Context) error {
	word := db.GetRandomWord(int(c.Sender().ID))
	if word.Id == 0 {
		return c.Send("You have no words😭")
	}

	wordSend := ""

	if rand.Intn(2) == 0 {
		wordSend = word.WordEn
	} else {
		wordSend = word.WordRu
	}

	inlineKeys := [][]tb.InlineButton{
		{
			tb.InlineButton{Data: "check_word\n" + strconv.Itoa(word.Id), Text: "Check"},
		},
	}
	inlineMarkup := &tb.ReplyMarkup{InlineKeyboard: inlineKeys}
	return c.Send(fmt.Sprintf("Word: *%s*\n\nTranslate it please🥺", wordSend), &tb.SendOptions{ParseMode: tb.ModeMarkdown}, inlineMarkup)
}

func main() {

	waitingTranslion = make(map[int]string)

	pref := tb.Settings{
		Token:  os.Getenv("BOT_TOKEN"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tb.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	b.SetCommands([]tb.Command{
		{Text: "/stat", Description: "Show statistics"},
		{Text: "/quiz", Description: "Start quiz"},
	})

	db = initDB()
	log.Println("Connected to database")

	b.Use(Logger)

	b.Handle("/start", func(c tb.Context) error {
		return c.Send("Hello! This is a bot for learning English words.🤯\nWrite any word to translate it and add to your dictionary.📝\n\nBy @goshanmorev with ❤️")
	})

	b.Handle("/stat", func(c tb.Context) error {
		total, learning := db.GetWordsAmount(int(c.Sender().ID))
		return c.Send(fmt.Sprintf("Total words📚: %d\nLearning words🧠: %d", total, learning))
	})

	b.Handle("/quiz", startQuiz)

	b.Handle(tb.OnText, func(c tb.Context) error {

		if _, ok := waitingTranslion[int(c.Sender().ID)]; ok {
			inlineKeys := [][]tb.InlineButton{{
				tb.InlineButton{Data: "add_word\n" + waitingTranslion[int(c.Sender().ID)] + "\n" + c.Text(), Text: "Add " + c.Text()},
			}}
			inlineMarkup := &tb.ReplyMarkup{InlineKeyboard: inlineKeys}
			c.Send("Add new word❓\n\nWord: *"+waitingTranslion[int(c.Sender().ID)]+"*\nTranslation: *"+c.Text()+"*", &tb.SendOptions{ParseMode: tb.ModeMarkdown}, inlineMarkup)

			delete(waitingTranslion, int(c.Sender().ID))
			return nil
		}

		word_en := c.Text()
		word_ru, err := translate(word_en)
		if err != nil {
			log.Println(err)
		}

		inlineKeys := [][]tb.InlineButton{
			{tb.InlineButton{Data: "add_word\n" + word_en + "\n" + word_ru, Text: "Add " + word_ru}},
			{tb.InlineButton{Data: "change_translation\n" + word_en, Text: "Change translation"}},
		}
		inlineMarkup := &tb.ReplyMarkup{InlineKeyboard: inlineKeys}
		return c.Send("Add new word❓\n\nWord: *"+word_en+"*\nTranslation: *"+word_ru+"*", &tb.SendOptions{ParseMode: tb.ModeMarkdown}, inlineMarkup)

	})

	b.Handle(tb.OnCallback, func(c tb.Context) error {
		data := strings.Split(c.Callback().Data, "\n")
		switch data[0] {

		case "add_word":
			err := db.AddWord(data[1], data[2], int(c.Sender().ID))
			if err != nil {
				c.Send(err.Error())
			} else {
				c.Send("Word added🎉\n\nWord: *"+data[1]+"*\nTranslation: *"+data[2]+"*", &tb.SendOptions{ParseMode: tb.ModeMarkdown})
			}

		case "change_translation":
			waitingTranslion[int(c.Sender().ID)] = data[1]
			c.Send("Choose translation:")

		case "check_word":
			wordId, _ := strconv.Atoi(data[1])
			word := db.GetWord(wordId)
			inlineKeys := [][]tb.InlineButton{{
				tb.InlineButton{Data: "corr_word\n" + data[1], Text: "✅"},
				tb.InlineButton{Data: "wrong_word\n" + data[1], Text: "❌"},
			}}
			inlineMarkup := &tb.ReplyMarkup{InlineKeyboard: inlineKeys}
			c.Send(fmt.Sprintf("Answer📝:\n\nWord: *%s*\nTranslation: *%s*", word.WordEn, word.WordRu), &tb.SendOptions{ParseMode: tb.ModeMarkdown}, inlineMarkup)

		case "corr_word":
			wordId, _ := strconv.Atoi(data[1])
			word := db.GetWord(wordId)
			word.Corr += 1
			word.Total += 1
			db.UpdateWord(word)
			if word.Corr > 3 && word.Corr%2 == 0 {
				inlineKeys := [][]tb.InlineButton{
					{tb.InlineButton{Data: "del_word\n" + data[1], Text: "Delete it 🗑️"}},
					{tb.InlineButton{Data: "leave_word\n" + data[1], Text: "Leave it ✏️"}},
				}
				inlineMarkup := &tb.ReplyMarkup{InlineKeyboard: inlineKeys}
				return c.Send(fmt.Sprintf("You guessed this word: *%d out of %d times*", word.Corr, word.Total), &tb.SendOptions{ParseMode: tb.ModeMarkdown}, inlineMarkup)
			}
			return startQuiz(c)

		case "wrong_word":
			wordId, _ := strconv.Atoi(data[1])
			word := db.GetWord(wordId)
			word.Total += 1
			db.UpdateWord(word)
			return startQuiz(c)

		case "del_word":
			wordId, _ := strconv.Atoi(data[1])
			word := db.GetWord(wordId)
			word.Show = false
			db.UpdateWord(word)
			fallthrough

		case "leave_word":
			return startQuiz(c)

		default:
			c.Send("Unknown command 😔")
		}
		return nil
	})

	log.Println("Bot is running...")
	b.Start()

}
