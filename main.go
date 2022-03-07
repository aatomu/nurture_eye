package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"math/rand"
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
	prefix           = flag.String("prefix", "", "call prefix")
	token            = flag.String("token", "", "bot token")
	foodList         = []string{"遺骨", "遺骨", "遺骨", "遺骨", "遺骨", "遺骨", "遺骨", "遺骨", "遺骨", "遺骨"}
	usersData        = []*userItems{}
	randomPercentage = 30
	fromReplace      = "0123456789abcdef"
	fromArray        = strings.Split(fromReplace, "")
	toReplace        = flag.String("key", "", "bot token")
	toArray          = []string{}
)

type userItems struct {
	userID       string
	name         string
	staminaPoint int
	cutePoint    int
	intellPoint  int
	debufPoint   int
	speedPoint   int
}

func main() {
	//flag入手
	flag.Parse()
	fmt.Println("prefix       :", *prefix)
	fmt.Println("token        :", *token)
	fmt.Println("save key     :", *toReplace)
	toArray = strings.Split(*toReplace, "")

	//bot起動準備
	discord, err := discordgo.New("Bot " + *token)
	if err != nil {
		fmt.Println("Error logging")
	}

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
	//1秒に1回呼び出す
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			<-ticker.C
			botStateUpdate(discord)
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
	guildData, err := discord.Guild(guildID)
	guildName := ""
	if err == nil {
		guildName = guildData.Name
	} else {
		guildName = "DirectMessage"
	}
	channelID := m.ChannelID
	channelData, _ := discord.Channel(channelID)
	channelName := channelData.Name
	message := m.Content
	author := m.Author.Username
	authorID := m.Author.ID

	//表示
	log.Print("Guild:\"" + guildName + "\"  Channel:\"" + channelName + "\"  " + author + ": " + message)

	//bot return
	if m.Author.Bot {
		return
	}

	switch {
	//分岐
	case isPrefix(message, "fd "):
		if isBotChannel(channelName) || channelData.Type == 1 {
			giveFood(authorID, message, discord, channelID)
		}
		return
	case isPrefix(message, "name "):
		if isBotChannel(channelName) || channelData.Type == 1 {
			changeName(authorID, message, discord, channelID)
		}
		return
	case isPrefix(message, "le"):
		if isBotChannel(channelName) || channelData.Type == 1 {
			goLesson(authorID, message, discord, channelID)
		}
		return
	case isPrefix(message, "st"):
		if isBotChannel(channelName) || channelData.Type == 1 {
			sendState(authorID, message, discord, channelID)
		}
		return
	case isPrefix(message, "load "):
		if isBotChannel(channelName) || channelData.Type == 1 {
			userDataLoad(authorID, message, discord, channelID)
		}
		return
	case isPrefix(message, "fr"):
		if isBotChannel(channelName) || channelData.Type == 1 {
			goOut(authorID, discord, channelID)
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
	//SPの上り具合
	spUp := randomaizer(6)
	//UserData一覧
	shouldGenerateUserData := true
	ateEyeName := ""
	ateEyeStaminaPoint := 0

	//UsersDataにないか確認
	for _, user := range usersData {
		if user.userID == userID {
			user.staminaPoint = user.staminaPoint + spUp
			ateEyeName = user.name
			ateEyeStaminaPoint = user.staminaPoint
			shouldGenerateUserData = false
		}
	}
	//UsersDataになかったら新しく追加
	if shouldGenerateUserData {
		userData := generateUserData(userID)
		userData.staminaPoint = userData.staminaPoint + spUp
		ateEyeName = userData.name
		ateEyeStaminaPoint = userData.staminaPoint
		usersData = append(usersData, userData)
	}
	//渡したものの名前を入手
	foodName := strings.ReplaceAll(message, *prefix+" fd ", "")
	foodName = strings.ReplaceAll(foodName, "\n", "")
	//物の名前を追加
	for i := len(foodList) - 1; i >= 0; i = i - 1 {
		if i > 0 {
			foodList[i] = foodList[i-1]
		} else {
			foodList[0] = foodName
		}
	}
	//Embed用のデータ作成
	embedText := "<@" + userID + "> の**" + ateEyeName + "**は\n" +
		"**" + foodName + "**をたべて\n" +
		"StaminaPointが" + strconv.Itoa(ateEyeStaminaPoint) + "(+" + strconv.Itoa(spUp) + ") になった"
	//結果送信
	sendEmbed(discord, channelID, embedText)
}

func changeName(userID string, message string, discord *discordgo.Session, channelID string) {
	//UserData一覧
	shouldGenerateUserData := true
	eyeOldName := ""

	//新しい名前を入手
	newName := strings.ReplaceAll(message, *prefix+" name ", "")
	newName = strings.ReplaceAll(newName, "\n", "")
	newName = strings.ReplaceAll(newName, " ", "")

	//長さチェック
	if len(strings.Split(newName, "")) > 16 {
		embedText := "<@" + userID + "> のアイの新しい名前は長すぎるよ?"
		sendEmbed(discord, channelID, embedText)
		return
	}
	//UsersDataにないか確認
	for _, user := range usersData {
		if user.userID == userID {
			eyeOldName = user.name
			user.name = newName
			shouldGenerateUserData = false
		}
	}
	//UsersDataになかったら新しく追加
	if shouldGenerateUserData {
		userData := generateUserData(userID)
		eyeOldName = userData.name
		userData.name = newName
		usersData = append(usersData, userData)
	}

	//Embed用のデータ作成
	embedText := "<@" + userID + "> の**" + eyeOldName + "**は**" + newName + "**に名前が変わった!"
	//結果送信
	sendEmbed(discord, channelID, embedText)
}

func sendState(userID string, message string, discord *discordgo.Session, channelID string) {
	//もし指定されてたら変更
	if strings.HasPrefix(message, *prefix+" st ") {
		userID = strings.ReplaceAll(message, *prefix+" st ", "")
	}
	//UserIDからデータ探索
	shouldSendNotFoundUserEmbed := true
	for _, user := range usersData {
		if user.userID == userID {
			//Embed用テキスト作成
			userDataByUserID, err := discord.User(user.userID)
			if err != nil {
				log.Println(err)
				log.Println("		Error: NotFoundUserDataByUserID")
			}
			userName := userDataByUserID.Username
			embedText := "@" + userName + "のアイのすてーたす:\n" +
				"なまえ　　　　:　" + user.name + "\n" +
				"すたみな　　　:　" + strconv.Itoa(user.staminaPoint) + "\n" +
				"かわいさ　　　:　" + strconv.Itoa(user.cutePoint) + "\n" +
				"かしこさ　　　:　" + strconv.Itoa(user.intellPoint) + "\n" +
				"でばふぱわー　:　" + strconv.Itoa(user.debufPoint) + "\n" +
				"すばやさ　　　:　" + strconv.Itoa(user.speedPoint) + "\n" +
				"あじ　　　　　:　**" + foodList[randomaizer(len(foodList)-1)] + "** 味?\n"
			sendEmbed(discord, channelID, embedText)
			shouldSendNotFoundUserEmbed = false
		}
	}
	if shouldSendNotFoundUserEmbed {
		embedText := "その人のアイのすてーたすみつからなかったよ?"
		sendEmbed(discord, channelID, embedText)
	}
}

func goLesson(userID string, message string, discord *discordgo.Session, channelID string) {

	//UsersDataにないか確認
	shouldGenerateUserData := true
	for _, user := range usersData {
		if user.userID == userID {
			shouldGenerateUserData = false
			stateUp := randomaizer(10)
			selectStateUp := randomaizer(3)
			useStamina := randomaizer(15)
			//スタミナチェック
			if user.staminaPoint < useStamina {
				embedText := "せんせいはいたけど **" + user.name + "**のスタミナが足りなかったよ\n"
				sendEmbed(discord, channelID, embedText)
				return
			}
			user.staminaPoint = user.staminaPoint - useStamina

			//送るようembed
			embedText := ""

			//対面判定
			if randomaizer(100) <= randomPercentage {
				addSpeedPoint := randomaizer(3)
				user.speedPoint = user.speedPoint + addSpeedPoint
				embedText = "**" + user.name + "**は せんせいにあうことができなかった...\n" +
					"だけど すばやさが" + strconv.Itoa(user.speedPoint) + "(+" + strconv.Itoa(addSpeedPoint) + ")になった\n"
				selectStateUp = 0
			}

			switch selectStateUp {
			//きてぃせんせい
			case 1:
				user.cutePoint = user.cutePoint + stateUp
				embedText = "**" + user.name + "**は きてぃ せんせいにあうことができた!\n" +
					"かわいさが" + strconv.Itoa(user.cutePoint) + "(+" + strconv.Itoa(stateUp) + ")になった\n"
				//ぐーぐるせんせい
			case 2:
				user.intellPoint = user.intellPoint + stateUp
				embedText = "**" + user.name + "**は ぐーぐる せんせいにあうことができた!\n" +
					"かしこさが" + strconv.Itoa(user.intellPoint) + "(+" + strconv.Itoa(stateUp) + ")になった\n"
				//ないとめあせんせい
			case 3:
				user.debufPoint = user.debufPoint + stateUp
				embedText = "**" + user.name + "**は ないとめあ せんせいにあうことができた!\n" +
					"でばふぱわーが" + strconv.Itoa(user.debufPoint) + "(+" + strconv.Itoa(stateUp) + ")になった\n"

			}
			//暗号化
			dummyUserID := user.userID
			dummyStaminaPoint := fmt.Sprint(user.staminaPoint)
			dummyCutePoint := fmt.Sprint(user.cutePoint)
			dummyIntellPoint := fmt.Sprint(user.intellPoint)
			dummyDebufPoint := fmt.Sprint(user.debufPoint)
			dummySpeedPoint := fmt.Sprint(user.speedPoint)
			userSaveData := dummyUserID + " " + user.name + " " + dummyStaminaPoint + " " + dummyCutePoint + " " + dummyIntellPoint + " " + dummyDebufPoint + " " + dummySpeedPoint + " "
			//データをデコード
			saveData := hex.EncodeToString([]byte(userSaveData))
			sendCode := ""
			count := 0
			//分割してずらす
			for _, alpha := range strings.Split(saveData, "") {
				count++
				start := (strings.Index(fromReplace, alpha) + count) % len(fromArray)
				sendCode = sendCode + toArray[start]
			}
			embedText = embedText + "\n||SaveCode:\n**" + sendCode + "**||\n" +
				"コードは毎回変わるからね!"
			sendEmbed(discord, channelID, embedText)
			return
		}
	}

	//UsersDataになかったら新しく追加
	if shouldGenerateUserData {
		embedText := "せんせいはいたけど <@" + userID + ">のアイがこなかったよ"
		sendEmbed(discord, channelID, embedText)
		return
	}
}

func userDataLoad(userID string, message string, discord *discordgo.Session, channelID string) {
	//セーブコード入手
	saveCode := strings.ReplaceAll(message, *prefix+" load ", "")

	//復号
	loadData := ""
	count := 0
	//分割してずらす
	for _, alpha := range strings.Split(saveCode, "") {
		count++
		start := strings.Index(*toReplace, alpha) - count
		for i := 0; start < 0; i++ {
			start = start + len(toArray)
		}
		loadData = loadData + fromArray[start]
	}
	//データをエンコード
	data, _ := hex.DecodeString(loadData)
	log.Println("Load : \"" + string(data) + "\"")
	flatUserData := string(data)
	//Embed用
	embedText := ""

	//軽く確認
	if strings.Count(flatUserData, " ") == 7 && strings.HasPrefix(flatUserData, userID+" ") && strings.HasSuffix(flatUserData, " ") {
		//一時保存用
		dummyUserData := generateUserData("")

		//整形
		_, err := fmt.Sscanf(flatUserData, "%s %s %d %d %d %d %d ", &dummyUserData.userID, &dummyUserData.name, &dummyUserData.staminaPoint, &dummyUserData.cutePoint, &dummyUserData.intellPoint, &dummyUserData.debufPoint, &dummyUserData.speedPoint)
		if err != nil {
			log.Println(err)
			log.Println("		Error : Failed Encode LoadUserData")
			return
		}

		//UsersDataにないか確認
		shouldGenerateUserData := true
		for _, user := range usersData {
			if user.userID == dummyUserData.userID {
				//丸々移す
				user.userID = dummyUserData.userID
				user.name = dummyUserData.name
				user.staminaPoint = dummyUserData.staminaPoint
				user.cutePoint = dummyUserData.cutePoint
				user.intellPoint = dummyUserData.intellPoint
				user.debufPoint = dummyUserData.debufPoint
				user.speedPoint = dummyUserData.speedPoint
				//新しいデータを追加しない
				shouldGenerateUserData = false
				embedText = "<@" + user.userID + "> のアイのデータを読み込んだよ!\n" +
					"<@" + user.userID + "> のアイ(**" + user.name + "**)は嬉しそうだ!"
			}
		}
		//UsersDataになかったら新しく追加
		if shouldGenerateUserData {
			//dummyUserData.userID = userID
			usersData = append(usersData, dummyUserData)
			embedText = "<@" + dummyUserData.userID + "> のアイのデータを追加したよ!\n" +
				"<@" + dummyUserData.userID + "> のアイ(**" + dummyUserData.name + "**)は嬉しそうだ!"
		}
	} else {
		embedText = "<@" + userID + "> 嘘ついたりずるしたりしてない??\n"
	}
	//結果送信
	sendEmbed(discord, channelID, embedText)
}

func goOut(userID string, discord *discordgo.Session, channelID string) {
	//相手のユーザーデータ
	userCount := len(usersData)
	userNumber := randomaizer(userCount)
	log.Println(fmt.Sprint(userNumber))
	numberCount := 0
	playerData := generateUserData("")
	//自分のユーザーデータ
	userData := generateUserData("")

	//探索
	for _, user := range usersData {
		numberCount++
		//相手のユーザーデータを保存
		if userNumber == numberCount {
			if user.userID != userID {
				//丸々移す
				playerData.userID = user.userID
				playerData.name = user.name
				playerData.staminaPoint = user.staminaPoint
				playerData.cutePoint = user.cutePoint
				playerData.intellPoint = user.intellPoint
				playerData.debufPoint = user.debufPoint
				playerData.speedPoint = user.speedPoint
			} else {
				//少しごまかして移す
				playerData.userID = ""
				playerData.name = "????"
				playerData.staminaPoint = user.staminaPoint + randomaizer(200) - 100
				playerData.cutePoint = user.cutePoint + randomaizer(200) - 100
				playerData.intellPoint = user.intellPoint + randomaizer(200) - 100
				playerData.debufPoint = user.debufPoint + randomaizer(200) - 100
				playerData.speedPoint = user.speedPoint + randomaizer(200) - 100

			}
		}
		if user.userID == userID {
			//丸々移す
			userData.userID = user.userID
			userData.name = user.name
			userData.staminaPoint = user.staminaPoint
			userData.cutePoint = user.cutePoint
			userData.intellPoint = user.intellPoint
			userData.debufPoint = user.debufPoint
			userData.speedPoint = user.speedPoint
		}
	}
	//比較ステータス
	comparisonState := randomaizer(5)

	//表示用
	playerDataByID, err := discord.User(playerData.userID)
	playerName := ""
	if err == nil {
		playerName = playerDataByID.Username
	} else {
		playerName = "unknown"
	}
	embedText := "<@" + userData.userID + "> の **" + userData.name + "**は お出かけをした\n" +
		"**" + userData.name + "** は @" + playerName + " の **" + playerData.name + "**とあった\n\n"
	//比較
	if userData.staminaPoint >= 10 {
		switch comparisonState {
		case 1:
			//すたみなの比較
			switch {
			case userData.staminaPoint <= playerData.staminaPoint+50:
				embedText = embedText + "**" + playerData.name + "** のほうが 元気そうだ\n"
			case userData.staminaPoint >= playerData.staminaPoint-50:
				embedText = embedText + "**" + playerData.name + "** のは 元気がなさそうだ\n"
			default:
				embedText = embedText + "**" + playerData.name + "** は 一緒に遊んでくれるらしい\n"
			}
		case 2:
			//かわいさの比較
			switch {
			case userData.cutePoint <= playerData.cutePoint+30:
				embedText = embedText + "**" + playerData.name + "** は とても可愛かった\n"
			case userData.cutePoint >= playerData.cutePoint-30:
				embedText = embedText + "**" + playerData.name + "** と お買い物に出かけた\n"
			default:
				embedText = embedText + "**" + playerData.name + "** と お揃いの物を持ってた\n"
			}
		case 3:
			//かしこさの比較
			switch {
			case userData.intellPoint <= playerData.intellPoint+50:
				embedText = embedText + "**" + playerData.name + "** は とっても賢そう\n"
			case userData.intellPoint >= playerData.intellPoint-50:
				embedText = embedText + "**" + playerData.name + "** の　勉強を手伝った\n"
			default:
				embedText = embedText + "**" + playerData.name + "** と 会話を楽しんだ\n"
			}
		case 4:
			//でばふぱわーの比較
			switch {
			case userData.debufPoint <= playerData.debufPoint+70:
				embedText = embedText + "**" + playerData.name + "** は とても強そう\n"
			case userData.debufPoint >= playerData.debufPoint-70:
				embedText = embedText + "**" + playerData.name + "** の　どくを教えた\n"
			default:
				embedText = embedText + "**" + playerData.name + "** と なかよくなれそう\n"
			}
		case 5:
			//すばやさの比較
			switch {
			case userData.speedPoint <= playerData.speedPoint+100:
				embedText = embedText + "**" + playerData.name + "** は 走るのが好きらしい\n"
			case userData.speedPoint >= playerData.speedPoint-100:
				embedText = embedText + "**" + playerData.name + "** に 驚いてにげちゃった\n"
			default:
				embedText = embedText + "**" + playerData.name + "** と 追いかけっこをした\n"
			}
		}
		for _, user := range usersData {
			if user.name == userID {
				//スタミナを減らす
				user.staminaPoint = user.staminaPoint - 10
			}
		}
	} else {
		embedText = embedText + "だけど すたみな がなくって すぐに家に帰った\n"
	}
	sendEmbed(discord, channelID, embedText)
}

func sendHelp(discord *discordgo.Session, channelID string) {
	embedText := "Bot Help\n" +
		*prefix + " fd <単語> : 自分のアイにごはんを上げます\n" +
		*prefix + " lesson: 自分のアイがじゅぎょーをうけ\n" +
		*prefix + " state : 自分のアイのすてーたすを確認します\n" +
		*prefix + " load <コード>: 自分のアイのデータの保存コードから読み込みます\n" +
		*prefix + " free : 自分のアイがお出かけします\n" +
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

//乱数入手用
func randomaizer(limit int) (random int) {
	//乱数のseed
	rand.Seed(time.Now().UnixNano())
	random = rand.Intn(limit) + 1
	return
}

//デフォルトユーザーデータ
func generateUserData(userID string) (userData *userItems) {
	return &userItems{
		userID:       userID,
		name:         "unknown",
		staminaPoint: 1,
		cutePoint:    1,
		intellPoint:  1,
		debufPoint:   1,
		speedPoint:   1,
	}
}
