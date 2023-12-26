package playerRegistation

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	generaldb "github.com/p-society/gCSB/connector/db"
	mailer "github.com/p-society/gCSB/connector/helpers"
	tempModel "github.com/p-society/gCSB/connector/model"
	"go.mongodb.org/mongo-driver/bson"
)

const otpValidityMinutes = 5

func generateOTP() (int, time.Time) {
	rand.Seed(time.Now().UnixNano())

	randomNumber := rand.Intn(10000) + 10000
	expirationTime := time.Now().Add(otpValidityMinutes * time.Minute)

	return randomNumber, expirationTime
}

func isOTPValid(expirationTime time.Time) bool {
	return time.Now().Before(expirationTime)
}

func PlayerRegistrationController(w http.ResponseWriter, r *http.Request) {
	var player tempModel.PlayerTemp
	err := json.NewDecoder(r.Body).Decode(&player)
	if err != nil {
		fmt.Println(err)
		return
	}

	OTP, expirationTime := generateOTP()
	player.OTP = OTP
	player.Verified = false
	player.OTPExpiration = expirationTime
	inserted, err := generaldb.Collection.InsertOne(r.Context(), player)

	content := `
	<html>
		<body style="font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif; background-color: #f5f5f5; color: #333; text-align: center; padding: 20px;">

			<img src="https://avatars.githubusercontent.com/u/29895434?s=400&u=59686b8af9d2476daa8886e6616fed4c76588596&v=4" alt="GCSB Logo" style="max-width: 200px; margin-bottom: 20px;">

			<h2 style="color: #007BFF;">Dear ` + player.Name + `,</h2>

			<p style="font-size: 16px;">Greetings from PSoc, IIIT-Bhubaneswar!</p>

			<div style="background-color: #ffffff; border-radius: 10px; padding: 20px; box-shadow: 0 0 10px rgba(0,0,0,0.1);">
				<p style="font-size: 18px;">Your One-Time Password (OTP) for GCSB Verification:</p>
				<h3 style="background-color: #007BFF; color: #fff; padding: 10px; border-radius: 5px; font-size: 24px;">` + strconv.Itoa(OTP) + `</h3>
				<p style="font-size: 16px;">This OTP is valid for a limited time to ensure the security of your account.</p>
			</div>

			<p style="font-size: 16px; margin-top: 20px;">If you did not request this OTP, please disregard this email.</p>

			<p style="font-size: 16px;">Best regards,<br/>
			The PSoc Team, IIIT-Bhubaneswar</p>

		</body>
	</html>`

	subject := "Player Email Verification - GCSB"

	var wg sync.WaitGroup

	ch := make(chan string)
	wg.Add(1)

	go func() {
		defer wg.Done()
		val := mailer.SendMail(content, subject, player.Email)
		ch <- val
		close(ch)
	}()

	fmt.Println("val:", <-ch)
	wg.Wait()

	if err != nil {
		fmt.Println(err)
		return
	}

	jsonPayload := map[string]interface{}{
		"inserted": inserted,
		"message":  "Verify yourself by plugging the OTP sent on you email provided.",
	}

	err = json.NewEncoder(w).Encode(jsonPayload)

	if err != nil {
		fmt.Println(err)
		return
	}

}

type CallbackMessage struct {
	Name        string `json:"name,omitempty" bson:"name,omitempty"`
	ReceivedOTP int    `json:"otp"`
}

func VerifyPlayerController(w http.ResponseWriter, r *http.Request) {

	var (
		message CallbackMessage
		player  tempModel.PlayerTemp
	)

	err := json.NewDecoder(r.Body).Decode(&message)

	if err != nil {
		fmt.Println(err)
		return
	}

	filter := bson.M{
		"name": message.Name,
	}

	res := generaldb.Collection.FindOne(r.Context(), filter)
	res.Decode(&player)

	if isOTPValid(player.OTPExpiration) && player.OTP == message.ReceivedOTP {

		filter := bson.M{"name": player.Name}

		update := bson.D{
			{"$set", bson.D{
				{"verified", true},
			}},
		}

		fmt.Println("User is Certified Stoner")
		updres, err := generaldb.Collection.UpdateOne(r.Context(), filter, update)

		if err != nil {
			fmt.Println("err:", err)
			json.NewEncoder(w).Encode(err)
		}
		fmt.Println(updres)

		w.Write([]byte("Verified!"))
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"err": "OTP validation failed,Either Timed Out or Invalid OTP!",
		})
	}

}

const CAPTAIN_UNIQUE_KEY = "iiitbbsr"

/*
Captain is manually added to the <sport>:<collection> by default.
Work of Captain is to select the players and create the roster.
*/

type CaptainResponseCreator struct {
	Team    string `json:"team,omitempty"`
	Sport   string `json:"sport,omitempty"`
	Passkey string `json:"passkey,omitempty"`
}

func CaptainFetchController(w http.ResponseWriter, r *http.Request) {

	var capFilter CaptainResponseCreator
	var players []tempModel.PlayerTemp
	err := json.NewDecoder(r.Body).Decode(&capFilter)

	if err != nil {
		json.NewEncoder(w).Encode(err)
		fmt.Println(err)
		return
	}

	if capFilter.Passkey != CAPTAIN_UNIQUE_KEY {
		errPayload := map[string]interface{}{
			"error": "Captain passkey error,Auth Denied.",
		}
		json.NewEncoder(w).Encode(errPayload)
		return
	}

	filter := bson.M{
		"team":     capFilter.Team,
		"sport":    capFilter.Sport,
		"verified": true,
	}

	cursor, err := generaldb.Collection.Find(r.Context(), filter)
	if err != nil {
		json.NewEncoder(w).Encode(err)
		fmt.Println(err)
		return
	}

	defer cursor.Close(r.Context())

	for cursor.Next(r.Context()) {
		var player tempModel.PlayerTemp
		err := cursor.Decode(&player)
		if err != nil {
			log.Fatal(err)
		}
		players = append(players, player)
	}

	json.NewEncoder(w).Encode(players)

	// Checking for errors during cursor iteration
	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	// Print or process the retrieved documents
	for _, doc := range players {
		fmt.Printf("%+v\n", doc)
	}
}

func CaptainPushController(w http.ResponseWriter, r *http.Request) {
	var players []tempModel.PlayerTemp

	err := json.NewDecoder(r.Body).Decode(&players)

	if err != nil {
		json.NewEncoder(w).Encode(err)
		fmt.Println(err)
		return
	}
	// fmt.Println(players);

	var insertPayload []interface{}
	for _, player := range players {
		insertPayload = append(insertPayload, player)
	}

	dbName := players[0].Sport + "db"
	colName := "players"
	// db := generaldb.Client.Database(dbName)
	// col := db.Collection("players")
	col := generaldb.ProvideAptCollection(dbName, colName)

	ins, err := col.InsertMany(r.Context(), insertPayload)

	if err != nil {
		json.NewEncoder(w).Encode(err)
		fmt.Println(err)
		return
	}

	fmt.Println(ins)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"roster_created": "true",
		"objects":        len(players),
	})
}
