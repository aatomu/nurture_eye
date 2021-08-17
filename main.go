package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	//変数定義
	prefix   = flag.String("prefix", "", "call prefix")
	token    = flag.String("token", "", "bot token")
	clientID = ""
)

func main() {
	//flag入手
	flag.Parse()
	fmt.Println("prefix       :", *prefix)
	fmt.Println("token        :", *token)

	//bot起動準備
	discord, err := discordgo.New()
	if err != nil {
		fmt.Println("Error logging")
	}

	//token入手
	discord.Token = "Bot " + *token

	//eventトリガー設定
	discord.AddHandler(onReady)
	discord.AddHandler(onMessageCreate)

	//起動
	if err = discord.Open(); err != nil {
		fmt.Println(err)
	}
	defer func() {
		if err := discord.Close(); err != nil {
			log.Println(err)
		}
	}()
	//起動メッセージ表示
	fmt.Println("Listening...")

	//bot停止対策
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

}

//BOTの準備が終わったときにCall
func onReady(discord *discordgo.Session, r *discordgo.Ready) {
	clientID = discord.State.User.ID
	//1秒に1回呼び出す
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				botStateUpdate(discord)
			}
		}
	}()
}

func botStateUpdate(discord *discordgo.Session) {
	//botのステータスアップデート
	joinedServer := len(discord.State.Guilds)
	state := *prefix + " help | " + strconv.Itoa(joinedServer) + "鯖にいるよ!"
	discord.UpdateGameStatus(0, state)
}

//メッセージが送られたときにCall
func onMessageCreate(discord *discordgo.Session, m *discordgo.MessageCreate) {
	//一時変数
	guildID := m.GuildID
	guildData, _ := discord.Guild(guildID)
	guild := guildData.Name
	channelID := m.ChannelID
	channel, _ := discord.Channel(channelID)
	//	messageID := m.ID
	message := m.Content
	author := m.Author.Username
	//	authorID := m.Author.ID

	//表示
	log.Print("Guild:\"" + guild + "\"  Channel:\"" + channel.Name + "\"  " + author + ": " + message)

	switch {
	//分岐
	case prefixCheck(message, "join"):
		return
	case prefixCheck(message, "speed "):
		return
	case prefixCheck(message, "pitch "):
		return
	case prefixCheck(message, "help"):
		sendHelp(discord, channelID)
		return
	}
}

func prefixCheck(message, check string) bool {
	return strings.HasPrefix(message, *prefix+" "+check)
}

func sendHelp(discord *discordgo.Session, channelID string) {
	//embedのData作成
	embed := &discordgo.MessageEmbed{
		Type:        "rich",
		Title:       "BOT HELP",
		Description: "",
		Color:       1000,
	}
	Text := "--TTS--\n" +
		*prefix + " give : 自分のアイにご飯を上げます\n" +
		*prefix + " state : 自分のアイのステータスを確認します\n" +
		*prefix + " adventure : ランダムなplayerに勝負をかけます\n"
	embed.Description = Text
	//送信
	if _, err := discord.ChannelMessageSendEmbed(channelID, embed); err != nil {
		log.Println(err)
	}
}

//リアクション追加用
func addReaction(discord *discordgo.Session, channelID string, messageID string, reaction string) {
	err := discord.MessageReactionAdd(channelID, messageID, reaction)
	if err != nil {
		log.Print("Error: addReaction Failed")
		log.Println(err)
	}
}

//ファイル読み込み
func readFile(filePath string) (text string, returnErr error) {
	//ファイルがあるか確認
	_, err := os.Stat(filePath)
	//ファイルがなかったら作成
	if os.IsNotExist(err) {
		_, err = os.Create(filePath)
		if err != nil {
			log.Println(err)
			returnErr = fmt.Errorf("missing crate file")
			return
		}
	}

	//読み込み
	byteText, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Println(err)
		returnErr = fmt.Errorf("missing read file")
		return
	}

	//[]byteをstringに
	text = string(byteText)
	return
}
