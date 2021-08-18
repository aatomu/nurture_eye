package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	//変数定義
	prefix           = flag.String("prefix", "", "call prefix")
	token            = flag.String("token", "", "bot token")
	clientID         = ""
	file             sync.Mutex
	canDeadByAteFood = 15
	randomPercentage = 30
	critStateUp      = 2
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
	authorID := m.Author.ID

	//表示
	log.Print("Guild:\"" + guild + "\"  Channel:\"" + channel.Name + "\"  " + author + ": " + message)

	//bot return
	if m.Author.Bot {
		return
	}

	switch {
	//分岐
	case isPrefix(message, "give "):
		if isBotChannel(channel.Name) {
			giveFood(authorID, message, discord, channelID)
		}
		return
	case isPrefix(message, "fd "):
		if isBotChannel(channel.Name) {
			giveFood(authorID, message, discord, channelID)
		}
		return
	case isPrefix(message, "state"):
		if isBotChannel(channel.Name) {
			sendState(authorID, message, discord, channelID)
		}
		return
	case isPrefix(message, "st"):
		if isBotChannel(channel.Name) {
			sendState(authorID, message, discord, channelID)
		}
		return
	case isPrefix(message, "adventure"):
		if isBotChannel(channel.Name) {
			goAdventure(authorID, discord, channelID)
		}
		return
	case isPrefix(message, "adv"):
		if isBotChannel(channel.Name) {
			goAdventure(authorID, discord, channelID)
		}
		return
	case isPrefix(message, "help"):
		sendHelp(discord, channelID)
		return
	}
}

func isPrefix(message, check string) bool {
	return strings.HasPrefix(message, *prefix+" "+check)
}

func isBotChannel(channelName string) bool {
	return strings.Contains(channelName, "アイ育成")
}

func giveFood(userID string, message string, discord *discordgo.Session, channelID string) {
	//並列処理対策
	file.Lock()
	defer file.Unlock()

	fileName := "./UserAi.txt"
	//データ一覧入手
	text, err := readFile(fileName)
	if err != nil {
		log.Println(err)
	}

	//代入先
	writeText := ""
	food := [5]string{"TUSB", "TUSB", "TUSB", "TUSB", "TUSB"}
	hp := 1
	sp := 1
	strength := 1
	temper := "くさ"
	count := 0
	//探索
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, "UserID:"+userID) {
			fmt.Sscanf(line, "UserID:"+userID+" Food 1:%s 2:%s 3:%s 4:%s 5:%s HP:%d SP:%d Strength:%d Temper:%s Count:%d", &food[1], &food[2], &food[3], &food[4], &food[0], &hp, &sp, &strength, &temper, &count)
		}
		if line != "" && !strings.Contains(line, "UserID:"+userID) {
			writeText = writeText + line + "\n"
		}
	}

	//食べ物
	giveFood := strings.Replace(message, *prefix+" give ", "", -1)
	food[0] = strings.Replace(giveFood, *prefix+" fd ", "", -1)
	//ステータス上昇
	state := "アイは\"" + food[0] + "\"を食べた\n"
	rand.Seed(time.Now().UnixNano())
	stateUp := rand.Intn(3)
	up := rand.Intn(20) - 5
	//涙目のとき増幅率をn倍
	if count+1 >= canDeadByAteFood && rand.Intn(101) <= randomPercentage {
		state = state + "**なにか強くなりそうだ**\n"
		up = up * critStateUp
	}
	switch {
	case stateUp == 0:
		hp = hp + up
		state = state + "HPが" + strconv.Itoa(hp) + "(変化:" + strconv.Itoa(up) + ")になった\n"
		if hp < 1 {
			state = state + "死んでしまった\n"
			count = -2
		}
		break
	case stateUp == 1:
		up = up - 3
		sp = sp + up
		//マイナス対策
		if sp <= 0 {
			sp = 1
		}
		state = state + "SPが" + strconv.Itoa(sp) + "(変化:" + strconv.Itoa(up) + ")になった \n"
		break
	case stateUp == 2:
		strength = strength + up
		//マイナス対策
		if strength <= 0 {
			strength = 1
		}
		state = state + "攻撃力が" + strconv.Itoa(strength) + "(変化:" + strconv.Itoa(up) + ")になった\n"
		break
	}
	//性格変更
	changeTemper := rand.Intn(8)
	arr := [8]string{"主のことが好きのようだ", "????", "主のことが嫌いのようだ", "引きこもろうとしている", "何か恐れている", "????", "繧九↑縺｡繝ｼ", "お仕事が好き"}
	temper = arr[changeTemper]

	//退化確認
	userdata := ""
	shouldLive := 100
	count++
	if count >= canDeadByAteFood {
		rand.Seed(time.Now().UnixNano())
		shouldLive = rand.Intn(101)
	}
	if shouldLive <= randomPercentage && hp >= 1 {
		state = "アイは食べ過ぎで死んでしまった!"
	}
	if shouldLive > randomPercentage && count > 0 {
		if count >= canDeadByAteFood {
			state = state + "**涙目になっている...**\n"
		}
		userdata = "UserID:" + userID + " Food 1:" + food[0] + " 2:" + food[1] + " 3:" + food[2] + " 4:" + food[3] + " 5:" + food[4] + " HP:" + strconv.Itoa(hp) + " SP:" + strconv.Itoa(sp) + " Strength:" + strconv.Itoa(strength) + " Temper:" + temper + " Count:" + strconv.Itoa(count)
	}
	//最終書き込み内容
	writeText = writeText + userdata + "\n"

	//書き込み
	writeFile(fileName, writeText)

	//結果表示
	embed := "<@" + userID + ">はアイにご飯を上げた!\n" + state
	sendEmbed(discord, channelID, embed)
}

func sendState(userID string, message string, discord *discordgo.Session, channelID string) {
	//並列処理対策
	file.Lock()
	defer file.Unlock()

	fileName := "./UserAi.txt"
	//データ一覧入手
	text, err := readFile(fileName)
	if err != nil {
		log.Println(err)
	}

	//代入先
	food := [5]string{}
	hp := 1
	sp := 1
	strength := 1
	temper := "くさ"
	count := 0

	//userID
	if strings.Contains(message, *prefix+" state <@!") || strings.Contains(message, *prefix+" state <@") || strings.Contains(message, *prefix+" st <@!") || strings.Contains(message, *prefix+" st <@") {
		otherID := strings.Replace(message, *prefix+" state <@!", "", -1)
		otherID = strings.Replace(otherID, *prefix+" state <@", "", -1)
		otherID = strings.Replace(otherID, *prefix+" st <@!", "", -1)
		otherID = strings.Replace(otherID, *prefix+" st <@", "", -1)
		otherID = strings.Replace(otherID, ">", "", -1)
		if otherID != "" {
			userID = otherID
		}
	}
	//探索
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, "UserID:"+userID) {
			fmt.Sscanf(line, "UserID:"+userID+" Food 1:%s 2:%s 3:%s 4:%s 5:%s HP:%d SP:%d Strength:%d Temper:%s Count:%d", &food[0], &food[1], &food[2], &food[3], &food[4], &hp, &sp, &strength, &temper, &count)
			break
		}
	}

	//結果表示
	embed := ""
	if food[0] != "" {
		rand.Seed(time.Now().UnixNano())
		meet := rand.Intn(4)
		attribute := hp % 5
		if attribute <= 0 {
			attribute = 0
		}
		attributeArray := [5]string{"水", "火", "木", "光", "闇"}
		embed = "<@" + userID + ">のアイのステータス\n" +
			"Food:\n" +
			"1." + food[0] + "\n" +
			"2." + food[1] + "\n" +
			"3." + food[2] + "\n" +
			"4." + food[3] + "\n" +
			"5." + food[4] + "\n" +
			"HP:" + strconv.Itoa(hp) + "　SP:" + strconv.Itoa(sp) + "　攻撃力:" + strconv.Itoa(strength) + "\n" +
			"性格:" + temper + "\n" +
			"味: たぶん" + food[meet] + "味\n" +
			"属性:" + attributeArray[attribute]
	} else {
		embed = "<@" + userID + ">のアイのステータスはなかったよ?"
	}
	sendEmbed(discord, channelID, embed)
}

func goAdventure(userID string, discord *discordgo.Session, channelID string) {
	//並列処理対策
	file.Lock()
	defer file.Unlock()

	fileName := "./UserAi.txt"
	//データ一覧入手
	text, err := readFile(fileName)
	if err != nil {
		log.Println(err)
	}

	//代入先
	enemyID := ""
	enemyHp := 1
	enemySp := 1
	enemyStrength := 1
	//書き込みとかよう
	writeText := ""
	food := [5]string{"TUSB", "TUSB", "TUSB", "TUSB", "TUSB"}
	hp := 1
	sp := 1
	strength := 1
	temper := "くさ"
	count := 0
	//敵指定
	rand.Seed(time.Now().UnixNano())
	lines := strings.Count(text, "\n")
	enemyLine := rand.Intn(lines - 1)
	counter := 0

	//探索
	for _, line := range strings.Split(text, "\n") {
		if counter == enemyLine {
			blank := ""
			fmt.Sscanf(line, "UserID:%s Food 1:%s 2:%s 3:%s 4:%s 5:%s HP:%d SP:%d Strength:%d Temper: Count:", &enemyID, &blank, &blank, &blank, &blank, &blank, &enemyHp, &enemySp, &enemyStrength)
		}
		if strings.Contains(line, "UserID:"+userID) {
			fmt.Sscanf(line, "UserID:"+userID+" Food 1:%s 2:%s 3:%s 4:%s 5:%s HP:%d SP:%d Strength:%d Temper:%s Count:%d", &food[0], &food[1], &food[2], &food[3], &food[4], &hp, &sp, &strength, &temper, &count)
		}
		counter++
		if line != "" && !strings.Contains(line, "UserID:"+userID) {
			writeText = writeText + line + "\n"
		}
	}

	//宣言
	embed := "<@" + userID + "> のアイは冒険に出た\n" +
		"<@" + userID + ">のアイ HP:" + strconv.Itoa(hp) + " SP:" + strconv.Itoa(sp) + " 攻撃力:" + strconv.Itoa(strength) + "\n"
	if !strings.Contains(enemyID, "MC:") {
		userData, _ := discord.User(enemyID)
		userName := userData.Username
		embed = embed + "@" + userName + "のアイ HP:" + strconv.Itoa(enemyHp) + " SP:" + strconv.Itoa(enemySp) + " 攻撃力:" + strconv.Itoa(enemyStrength) + " に勝負を仕掛けた!"
	} else {
		embed = embed + enemyID + "のアイ HP:" + strconv.Itoa(enemyHp) + " SP:" + strconv.Itoa(enemySp) + " 攻撃力:" + strconv.Itoa(enemyStrength) + " に勝負を仕掛けた!"
	}
	sendEmbed(discord, channelID, embed)

	//対決
	var isWin bool
	embed = ""
	dummyHp := hp
	dummySp := sp
	dummyStrength := strength
	for {
		//自分攻撃
		if dummySp >= 1 {
			rand.Seed(time.Now().UnixNano())
			damage := rand.Intn(3) * dummyStrength
			enemyHp = enemyHp - damage
			embed = embed + "自分ターン: " + strconv.Itoa(damage) + "damage 相手HP:" + strconv.Itoa(enemyHp) + "\n"
			if enemyHp <= 0 {
				embed = embed + "勝負に勝った\n自分のアイのおなかがすいたようだ\n"
				isWin = true
				break
			}
			dummySp = dummySp - 1
		}
		//敵攻撃
		if enemySp >= 1 {
			dummyHp = dummyHp - enemyStrength
			embed = embed + "相手ターン: " + strconv.Itoa(enemyStrength) + "damage 自分HP:" + strconv.Itoa(dummyHp) + "\n"
			if dummyHp <= 0 {
				embed = embed + "自分のアイは死んでしまった"
				isWin = false
				break
			}
			enemySp = enemySp - 1
		}
		if dummySp == 0 && enemySp == 0 {
			embed = embed + "相手のSPが切れて逃げた"
			isWin = true
			break
		}
	}

	userdata := ""
	if isWin {
		count = 0
		rand.Seed(time.Now().UnixNano())
		addHp := rand.Intn(20)
		hp = hp + addHp
		addSp := rand.Intn(5)
		sp = sp + addSp
		addStrength := rand.Intn(15)
		strength = strength + addStrength
		userdata = "UserID:" + userID + " Food 1:" + food[0] + " 2:" + food[1] + " 3:" + food[2] + " 4:" + food[3] + " 5:" + food[4] + " HP:" + strconv.Itoa(hp) + " SP:" + strconv.Itoa(sp) + " Strength:" + strconv.Itoa(strength) + " Temper:" + temper + " Count:" + strconv.Itoa(count)
		embed = embed + "HP:" + strconv.Itoa(hp) + "(+" + strconv.Itoa(addHp) + ") SP:" + strconv.Itoa(sp) + "(+" + strconv.Itoa(addSp) + ") 攻撃力:" + strconv.Itoa(strength) + "(+" + strconv.Itoa(addStrength) + ")\n"
	} else {
		count = canDeadByAteFood
		rand.Seed(time.Now().UnixNano())
		//確率でステータス返却
		if rand.Intn(101) <= randomPercentage {
			hp = hp/2 + 1
			sp = sp/2 + 1
			strength = strength/2 + 1
			userdata = "UserID:" + userID + " Food 1:" + food[0] + " 2:" + food[1] + " 3:" + food[2] + " 4:" + food[3] + " 5:" + food[4] + " HP:" + strconv.Itoa(hp) + " SP:" + strconv.Itoa(sp) + " Strength:" + strconv.Itoa(strength) + " Temper:" + temper + " Count:" + strconv.Itoa(count)
			embed = embed + "遺品を少し回収することができた\n"
		}
	}
	//最終書き込み内容
	writeText = writeText + userdata + "\n"

	//書き込み
	writeFile(fileName, writeText)

	//データ送信
	sendEmbed(discord, channelID, embed)
}

func sendHelp(discord *discordgo.Session, channelID string) {
	text := "Bot Help\n" +
		*prefix + " give <単語> : 自分のアイにご飯を上げます\n" +
		*prefix + " state : 自分のアイのステータスを確認します\n" +
		*prefix + " adventure : ランダムなplayerに勝負をかけます\n" +
		"*help以外のコマンドは\"アイ育成\"を含む\n" +
		"名前のチャンネルでのみ反応します\n"
	sendEmbed(discord, channelID, text)
}

//Embed送信
func sendEmbed(discord *discordgo.Session, channelID string, text string) {
	//embedのData作成
	embed := &discordgo.MessageEmbed{
		Type:        "rich",
		Description: text,
		Color:       1000,
	}
	//送信
	_, err := discord.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
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

//ファイル書き込み
func writeFile(filePath string, writeText string) {
	//書き込み
	err := ioutil.WriteFile(filePath, []byte(writeText), 0777)
	if err != nil {
		log.Println(err)
	}
}
