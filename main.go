package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
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
	filePath         = "./UserEye.txt"
	canDeadByAteFood = 15
	randomPercentage = 30
	critStateUp      = 2
)

type userItems struct {
	userID string
	food1  string
	food2  string
	food3  string
	hp     int
	str    int
	count  int
}

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
	state := *prefix + " help | " + strconv.Itoa(joinedServer) + "鯖にいるよ! データは消えるもの!"
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
	case isPrefix(message, "fd "):
		if isBotChannel(channel.Name) {
			giveFood(authorID, message, discord, channelID)
		}
		return
	case isPrefix(message, "st"):
		if isBotChannel(channel.Name) {
			sendState(authorID, message, discord, channelID)
		}
		return
	case isPrefix(message, "ad"):
		if isBotChannel(channel.Name) {
			goAdventure(authorID, discord, channelID)
		}
		return
	case isPrefix(message, "save"):
		if isBotChannel(channel.Name) {
			dataSave(authorID, discord, channelID)
		}
		return
	case isPrefix(message, "load "):
		if isBotChannel(channel.Name) {
			dataLoad(authorID, message, discord, channelID)
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
	//重複処理対策
	file.Lock()
	defer file.Unlock()

	//テキストデータ
	text := readFile(filePath)
	//ユーザーデータ
	userData := &userItems{
		userID: userID,
		food1:  "遺骨",
		food2:  "遺骨",
		food3:  "遺骨",
		hp:     1,
		str:    1,
		count:  0,
	}
	//書き込みデータ
	writeText := ""
	//データ変換
	for _, line := range strings.Split(text, "\n") {
		//userData変換
		if strings.HasPrefix(line, userData.userID) {
			_, err := fmt.Sscanf(line, "%s %s %s %s %d %d %d", &userData.userID, &userData.food1, &userData.food2, &userData.food3, &userData.hp, &userData.str, &userData.count)
			if err != nil {
				log.Println(err)
				log.Println("		Error: Failed Fmt Line To UserData")
			}
		}
		//書き込みデータ
		if !strings.HasPrefix(line, userData.userID) && line != "" {
			writeText = writeText + line + "\n"
		}
	}

	//embed用Textの生成
	embedText := ""

	//食べ物関連
	//渡した物の名前
	foodName := strings.Replace(message, *prefix+" fd ", "", -1)
	foodName = strings.ReplaceAll(foodName, " ", "")
	foodName = strings.ReplaceAll(foodName, "\n", "")
	embedText = embedText + "<@" + userData.userID + ">はアイに**" + foodName + "**を渡した\n"
	//物の順番を変える
	userData.food3 = userData.food2
	userData.food2 = userData.food1
	userData.food1 = foodName

	//ステータスアップ
	//どのステータスを上げるか乱数設定
	random := randomaizer(2)
	stateUp := randomaizer(20) - 5
	if userData.count+1 >= canDeadByAteFood && randomaizer(100) <= randomPercentage {
		stateUp = stateUp * 2
		embedText = embedText + "\\*いつもよりも変化した気がする\n"
	}
	switch random {
	//体力ステータス変更
	case 1:
		userData.hp = userData.hp + stateUp
		stateUpText := strconv.Itoa(stateUp)
		if stateUp > 0 {
			stateUpText = "+" + stateUpText
		}
		embedText = embedText + "体力が" + strconv.Itoa(userData.hp) + "(" + stateUpText + ") になった\n"
	//攻撃ステータス変更
	case 2:
		userData.str = userData.str + stateUp
		stateUpText := strconv.Itoa(stateUp)
		if stateUp > 0 {
			stateUpText = "+" + stateUpText
		}
		//攻撃 0以下 回避
		if userData.str < 1 {
			userData.str = 1
		}
		embedText = embedText + "攻撃が" + strconv.Itoa(userData.str) + "(" + stateUpText + ") になった\n"
	}

	//死亡判定
	userData.count++
	shouldDead := false
	//Count確認
	if userData.count >= canDeadByAteFood {
		if randomaizer(100) <= randomPercentage {
			embedText = embedText + "\\*おなかいっぱいで死んでしまった\\*\n"
			shouldDead = true
		} else {
			embedText = embedText + "**涙目になっている**\n"
		}
	}
	if userData.hp < 1 {
		embedText = embedText + "\\*体力がなくて死んでしまった\\*\n"
		shouldDead = true
	}

	//userData書き込み
	if !shouldDead {
		userText := userData.userID + " " +
			userData.food1 + " " +
			userData.food2 + " " +
			userData.food3 + " " +
			strconv.Itoa(userData.hp) + " " +
			strconv.Itoa(userData.str) + " " +
			strconv.Itoa(userData.count)
		writeText = writeText + userText + "\n"
	}
	//書き込み
	writeFile(filePath, writeText)

	//結果表示
	sendEmbed(discord, channelID, embedText)
}

func sendState(userID string, message string, discord *discordgo.Session, channelID string) {
	//重複処理対策
	file.Lock()
	defer file.Unlock()

	//テキストデータ
	text := readFile(filePath)
	//ユーザーデータ
	userData := &userItems{
		userID: userID,
		food1:  "遺骨",
		food2:  "遺骨",
		food3:  "遺骨",
		hp:     1,
		str:    1,
		count:  -1,
	}

	//もしIDが指定されてればそっちを表示
	if strings.HasPrefix(message, *prefix+" st ") {
		userData.userID = strings.Replace(message, *prefix+" st ", "", -1)
	}
	//データ変換
	for _, line := range strings.Split(text, "\n") {
		//userData変換
		if strings.HasPrefix(line, userData.userID) {
			_, err := fmt.Sscanf(line, "%s %s %s %s %d %d %d", &userData.userID, &userData.food1, &userData.food2, &userData.food3, &userData.hp, &userData.str, &userData.count)
			if err != nil {
				log.Println(err)
				log.Println("		Error: Failed Fmt Line To UserData")
			}
		}
	}

	//embed用Textの生成
	embedText := ""
	if userData.count >= 0 {
		//user名を平文に
		userName := ""
		isID := regexp.MustCompile(`[0-9]{18}`)
		if isID.MatchString(userData.userID) {
			userDataByID, _ := discord.User(userData.userID)
			userName = userDataByID.Username
		} else {
			userName = userData.userID
		}
		//固定部分
		embedText = "@" + userName + "のアイのステータス\n" +
			"食べたもの1:" + userData.food1 + "\n" +
			"食べたもの2:" + userData.food2 + "\n" +
			"食べたもの3:" + userData.food3 + "\n" +
			"体力:" + strconv.Itoa(userData.hp) + "  攻撃:" + strconv.Itoa(userData.str) + "\n"
		//可変部分
		//性格
		temperArr := []string{"", "主のことが好きのようだ", "????", "主のことが嫌いのようだ", "引きこもろうとしている", "何か恐れている", "????", "繧九↑縺｡繝ｼ", "お仕事が好き"}
		arrayLength := len(temperArr)
		arrayNumber := userData.hp*userData.str%(arrayLength-1) + 1
		temper := temperArr[arrayNumber]
		//味
		meetArr := []string{"", userData.food1, userData.food2, userData.food3}
		arrayLength = len(meetArr)
		meet := meetArr[randomaizer(arrayLength-1)]
		//属性
		attoributeArr := []string{"", "水", "火", "木", "光", "闇"}
		arrayLength = len(attoributeArr)
		arrayNumber = userData.hp*userData.str%(arrayLength-1) + 1
		attribute := attoributeArr[arrayNumber]
		//embed最終形
		embedText = embedText +
			"性格 : " + temper + "\n" +
			"味 : " + meet + "\n" +
			"属性 : " + attribute + "\n"
	} else {
		//user名を平文に
		userName := ""
		isID := regexp.MustCompile(`[0-9]{18}`)
		if isID.MatchString(userData.userID) {
			userDataByID, _ := discord.User(userData.userID)
			userName = userDataByID.Username
		} else {
			userName = userData.userID
		}
		//固定部分
		embedText = "@" + userName + "のアイのステータスはないよ?\n"
	}
	//結果表示
	sendEmbed(discord, channelID, embedText)

}

func goAdventure(userID string, discord *discordgo.Session, channelID string) {
	//重複処理対策
	file.Lock()
	defer file.Unlock()

	//テキストデータ
	text := readFile(filePath)
	//ユーザーデータ
	userData := &userItems{
		userID: userID,
		food1:  "遺骨",
		food2:  "遺骨",
		food3:  "遺骨",
		hp:     1,
		str:    1,
		count:  0,
	}
	enemyData := &userItems{
		userID: "unknown",
		food1:  "遺骨",
		food2:  "遺骨",
		food3:  "遺骨",
		hp:     1,
		str:    1,
		count:  0,
	}

	//書き込みデータ
	writeText := ""

	//敵指定
	lines := strings.Count(text, "\n")
	enemyLine := randomaizer(lines - 1)
	lineCount := 0

	//データ変換
	for _, line := range strings.Split(text, "\n") {
		//userData変換
		if strings.HasPrefix(line, userData.userID) {
			_, err := fmt.Sscanf(line, "%s %s %s %s %d %d %d", &userData.userID, &userData.food1, &userData.food2, &userData.food3, &userData.hp, &userData.str, &userData.count)
			if err != nil {
				log.Println(err)
				log.Println("		Error: Failed Fmt Line To UserData")
			}
		}
		//enemyData変換
		if lineCount == enemyLine {
			_, err := fmt.Sscanf(line, "%s %s %s %s %d %d %d", &enemyData.userID, &enemyData.food1, &enemyData.food2, &enemyData.food3, &enemyData.hp, &enemyData.str, &enemyData.count)
			if err != nil {
				log.Println("		Error: Failed Fmt Line To EnemyData")
				log.Println("		And Change Enemy to Herobrine")
				enemyData.userID = "MC: HeroBrine"
				enemyData.hp = enemyData.hp + randomaizer(5000)
				enemyData.str = enemyData.str + randomaizer(5000)
			}
		}
		lineCount++
		//書き込みデータ
		if !strings.HasPrefix(line, userData.userID) && line != "" {
			writeText = writeText + line + "\n"
		}
	}

	//宣言
	embedText := "<@" + userData.userID + "> のアイは冒険に出た\n" +
		"<@" + userData.userID + ">のアイ HP:" + strconv.Itoa(userData.hp) + " 攻撃力:" + strconv.Itoa(userData.str) + "\n"
	if !strings.Contains(enemyData.userID, "MC:") {
		enemyDataByID, _ := discord.User(enemyData.userID)
		enemyName := enemyDataByID.Username
		embedText = embedText + "@" + enemyName + "のアイ HP:" + strconv.Itoa(enemyData.hp) + " 攻撃力:" + strconv.Itoa(enemyData.str) + " に勝負を仕掛けた!\n\n"
	} else {
		embedText = embedText + enemyData.userID + "のアイ HP:" + strconv.Itoa(enemyData.hp) + " 攻撃力:" + strconv.Itoa(enemyData.str) + " に勝負を仕掛けた!\n\n"
	}

	//戦闘判定
	var isWin bool
	dummyHp := userData.hp
	dummyStr := userData.str
	damage := 0
	for {
		//自分攻撃
		damage = (randomaizer(3) - 1) * dummyStr
		enemyData.hp = enemyData.hp - damage
		embedText = embedText + "自分ターン: " + strconv.Itoa(damage) + "damage 相手HP:" + strconv.Itoa(enemyData.hp) + "\n"
		if enemyData.hp <= 0 {
			embedText = embedText + "勝負に勝った\n自分のアイのおなかがすいたようだ\n"
			isWin = true
			break
		}
		//敵攻撃
		crit := (randomaizer(10) - 1)
		if crit >= 3 || crit <= 9 {
			crit = 2
		}
		damage = crit * enemyData.str
		dummyHp = dummyHp - damage
		embedText = embedText + "相手ターン: " + strconv.Itoa(damage) + "damage 自分HP:" + strconv.Itoa(dummyHp) + "\n"
		if dummyHp <= 0 {
			embedText = embedText + "自分のアイは死んでしまった\n"
			isWin = false
			break
		}
	}

	//戦闘結果からstate変動
	userdata := ""
	if isWin {
		addHp := randomaizer(20)
		userData.hp = userData.hp + addHp
		addStrength := randomaizer(15)
		userData.str = userData.str + addStrength
		userData.count = 0
		userdata = userData.userID + " " + userData.food1 + " " + userData.food2 + " " + userData.food3 + " " + strconv.Itoa(userData.hp) + " " + strconv.Itoa(userData.str) + " " + strconv.Itoa(userData.count)
		embedText = embedText + "HP:" + strconv.Itoa(userData.hp) + "(+" + strconv.Itoa(addHp) + ") 攻撃力:" + strconv.Itoa(userData.str) + "(+" + strconv.Itoa(addStrength) + ")\n"
	} else {
		userData.count = canDeadByAteFood
		//確率でステータス返却
		if randomaizer(100) <= randomPercentage {
			userData.hp = userData.hp/2 + 1
			userData.str = userData.str/2 + 1
			userdata = userData.userID + " " + userData.food1 + " " + userData.food2 + " " + userData.food3 + " " + strconv.Itoa(userData.hp) + " " + strconv.Itoa(userData.str) + " " + strconv.Itoa(userData.count)
			embedText = embedText + "遺品を少し回収することができた\n"
		}
	}
	//最終書き込み内容
	writeText = writeText + userdata + "\n"

	//書き込み
	writeFile(filePath, writeText)

	//データ送信
	sendEmbed(discord, channelID, embedText)
}

func dataSave(userID string, discord *discordgo.Session, channelID string) {
	//重複処理対策
	file.Lock()
	defer file.Unlock()

	//テキストデータ
	text := readFile(filePath)

	//ユーザーデータ
	userData := &userItems{
		userID: userID,
		food1:  "遺骨",
		food2:  "遺骨",
		food3:  "遺骨",
		hp:     1,
		str:    1,
		count:  0,
	}

	//データ変換
	for _, line := range strings.Split(text, "\n") {
		//userData保存
		if strings.HasPrefix(line, userID) {
			_, err := fmt.Sscanf(line, "%s %s %s %s %d %d %d", &userData.userID, &userData.food1, &userData.food2, &userData.food3, &userData.hp, &userData.str, &userData.count)
			if err != nil {
				log.Println(err)
				log.Println("		Error: Failed Fmt Line To UserData")
			}
		}
	}

	//暗号化
	userSaveData := userData.userID + "," + strconv.Itoa(userData.hp) + "," + strconv.Itoa(userData.str) + "," + strconv.Itoa(userData.count) + ","
	saveData := hex.EncodeToString([]byte(userSaveData))

	//コード送信
	embedText := "<@" + userData.userID + "> コードができたよ!\n" +
		"Code: **" + saveData + "**\n" +
		"ユーザーごとに違うから注意してね!\n" +
		"データがないときにしかロードできないよ!"
	sendEmbed(discord, channelID, embedText)
}

func dataLoad(userID string, message string, discord *discordgo.Session, channelID string) {
	//重複処理対策
	file.Lock()
	defer file.Unlock()

	//テキストデータ
	text := readFile(filePath)

	//ユーザーデータ
	userData := &userItems{
		userID: userID,
		food1:  "遺骨",
		food2:  "遺骨",
		food3:  "遺骨",
		hp:     1,
		str:    1,
		count:  0,
	}

	//書き込みデータ
	writeText := ""
	//データ変換
	for _, line := range strings.Split(text, "\n") {
		//userData保存
		if strings.HasPrefix(line, userID) {
			embedText := "<@" + userID + "> の今のアイがかわいそうだよ...\n"
			sendEmbed(discord, channelID, embedText)
			return
		}
		//書き込みデータ
		if !strings.HasPrefix(line, userData.userID) && line != "" {
			writeText = writeText + line + "\n"
		}
	}

	//復号
	saveCode := strings.Replace(message, *prefix+" load ", "", -1)
	saveData, _ := hex.DecodeString(saveCode)
	saveUserData := strings.Split(string(saveData), ",")
	if len(saveUserData) == 5 {
		dummyHp, _ := strconv.Atoi(saveUserData[1])
		saveHp := dummyHp * 2 / 5 / randomaizer(10)
		dummyStr, _ := strconv.Atoi(saveUserData[2])
		saveStr := dummyStr * 2 / 5 / randomaizer(10)
		userData := saveUserData[0] + " LoadedThisEye LoadedThisEye LoadedThisEye " + strconv.Itoa(saveHp) + " " + strconv.Itoa(saveStr) + " " + saveUserData[3]
		writeText = writeText + userData + "\n"
		writeFile(filePath, writeText)
		embedText := "<@" + saveUserData[0] + "> のアイのデータを読み込んだよ!\n" +
			"よわくなってしまった\n" +
			"ステータス: 体力:" + strconv.Itoa(saveHp) + " 攻撃:" + strconv.Itoa(saveStr)
		sendEmbed(discord, channelID, embedText)
		log.Println("Loaded : " + userData)
	} else {
		embedText := "<@" + userID + "> さん 嘘ついてない?\n"
		sendEmbed(discord, channelID, embedText)

	}
}

func sendHelp(discord *discordgo.Session, channelID string) {
	embedText := "Bot Help\n" +
		*prefix + " fd <単語> : 自分のアイにご飯を上げます\n" +
		*prefix + " st : 自分のアイのステータスを確認します\n" +
		*prefix + " ad : ランダムなplayerに勝負をかけます\n" +
		*prefix + " save : 自分のアイのデータの保存コードを表示します\n" +
		*prefix + " load <コード>: 自分のアイのデータの保存コードから読み込みます\n" +
		"*help以外のコマンドは\"アイ育成\"を含む\n" +
		"名前のチャンネルでのみ反応します\n"
	sendEmbed(discord, channelID, embedText)
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
func readFile(filePath string) (text string) {
	//ファイルがあるか確認
	_, err := os.Stat(filePath)
	//ファイルがなかったら作成
	if os.IsNotExist(err) {
		_, err = os.Create(filePath)
		if err != nil {
			log.Println(err)
			log.Println("		Error: Faild Create File")
			return
		}
	}

	//読み込み
	byteText, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Println(err)
		log.Println("		Error: Faild Read File")
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
		log.Println("		Error: Faild Write File")
	}
}

//乱数入手用
func randomaizer(limit int) (random int) {
	//乱数のseed
	rand.Seed(time.Now().UnixNano())
	random = rand.Intn(limit) + 1
	return
}
