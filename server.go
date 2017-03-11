package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"

	"github.com/gorilla/context"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var OTP string

const connectionString = "mongodb://localhost/"

type endUserInfo struct {
	EndUsername string `bson:"endusername"`
	Password    string `bson:"password"`
	City        string `bson:"city"`
	State       string `bson:"state"`
	PinCode     string `bson:"pincode"`
	Email       string `bson:"email"`
	Status      string `bson:"status"`

	// Similarities above

	FirstName string `bson:"firstname"`
	LastName  string `bson:"lastname"`

	// time at end

	//	timestamp string `bson:"timestamp"`
}

type salonInfo struct {
	Username string `bson:"username"`
	Password string `bson:"password"`
	City     string `bson:"city"`
	State    string `bson:"state"`
	PinCode  string `bson:"pincode"`
	Email    string `bson:"email"`
	Status   string `bson:"status"`

	// Similarities above

	SalonName string `bson:"salonname"`
	ShortDesc string `bson:"shortdesc"`

	// time at end

	//	timestamp string `bson:"timestamp"`
}

type commentsSection struct {
	SalonName   string `bson:"salonname"`
	EndUsername string `bson:"endusername"`
	Comment     string `bson:"comment"`
	Star        string `bson:"star"`

	// time at end

	//	timestamp string `bson:"timestamp"`
}

func init() {
	OTP = ""
}

func main() {
	http.HandleFunc("/regUser", regUser)
	http.HandleFunc("/updateUser", updateUser)
	http.HandleFunc("/updateUserPassword", updateUserPassword)

	http.HandleFunc("/regSalon", regSalon)
	http.HandleFunc("/getComments", getComments)
	http.HandleFunc("/salComment", salComment)
	http.HandleFunc("/updateComments", updateComments)
	http.HandleFunc("/searchSaloon", searchSaloon)
	http.HandleFunc("/updateSalonPassword", updateSalonPassword)

	http.HandleFunc("/login", login)
	http.HandleFunc("/getUser", getUser)
	http.HandleFunc("/verify", verify)
	http.HandleFunc("/upload", uploadPage)

	fmt.Println("Server listening at port 80")
	http.ListenAndServe(":80", context.ClearHandler(http.DefaultServeMux))
	// using context.ClearHandler(http.DefaultServeMux) instead of nil to avoid memory leak
}

func verify(w http.ResponseWriter, r *http.Request) {
	var response string

	if r.Method == http.MethodPost {

		OTP = r.FormValue("otp")

		if OTP != "" && r.FormValue("type") == "Saloon" {
			var sal salonInfo

			sesson, err := mgo.Dial(connectionString)
			if err != nil {
				panic(err)
			}
			defer sesson.Close()

			sesson.SetMode(mgo.Monotonic, true)
			sal.Username = r.FormValue("user")

			c := sesson.DB("SaloonData").C("Saloon")

			result := salonInfo{}
			err = c.Find(bson.M{"username": sal.Username}).One(&result)
			if err != nil {
				response = "failed :("
			} else {
				err = sendMail(`
			<!DOCTYPE html>
			<html>
				<body>
					<h2> Hey there </h2>
					<p> Thank you for taking interest in Truudus. Here's your One Time Password </p> <br>
					<h3> `+OTP+`</h3>
				</body>
			</html>`, result.Email)
			}

			response = "Email Sent"
		} else if OTP != "" && r.FormValue("type") == "EndUser" {

			var user endUserInfo

			sesson, err := mgo.Dial(connectionString)
			if err != nil {
				panic(err)
			}
			defer sesson.Close()

			sesson.SetMode(mgo.Monotonic, true)
			user.EndUsername = r.FormValue("user")

			c := sesson.DB("SaloonData").C("Saloon")

			result := salonInfo{}
			err = c.Find(bson.M{"username": user.EndUsername}).One(&result)
			if err != nil {
				response = "failed :("
			} else {
				err = sendMail(`
			<!DOCTYPE html>
			<html>
				<body>
					<h2> Hey there </h2>
					<p> Thank you for taking interest in Truudus. Here's your One Time Password </p> <br>
					<h3> `+OTP+`</h3>
				</body>
			</html>`, result.Email)
			}

			response = "Email Sent"
			//******************************************************************************************************************//
			//												Status Updation Below											   //
			//****************************************************************************************************************//
		} else if r.FormValue("type") == "Saloon" {
			var sal salonInfo
			result := salonInfo{}

			sesson, err := mgo.Dial(connectionString)
			if err != nil {
				panic(err)
			}
			defer sesson.Close()

			sesson.SetMode(mgo.Monotonic, true)
			sal.Username = r.FormValue("user")
			sal.Status = "true"
			c := sesson.DB("SaloonData").C("Saloon")

			err = c.Find(bson.M{"username": sal.Username}).One(&result)
			if err != nil {
				log.Println(err)
			}

			sal.Password = result.Password
			sal.City = result.City
			sal.State = result.State
			sal.PinCode = result.PinCode
			sal.Email = result.Email
			sal.SalonName = result.SalonName
			sal.ShortDesc = result.ShortDesc

			colQuerier := bson.M{"username": sal.Username}
			err = c.Update(colQuerier, sal)
			if err != nil {
				response = "Failed :("
			} else {
				response = "Success"
			}

			sal.Status = "true"
		} else if r.FormValue("type") == "EndUser" {
			var user endUserInfo
			result := endUserInfo{}

			sesson, err := mgo.Dial(connectionString)
			if err != nil {
				panic(err)
			}
			defer sesson.Close()

			sesson.SetMode(mgo.Monotonic, true)
			user.EndUsername = r.FormValue("user")
			user.Status = "true"
			c := sesson.DB("SaloonData").C("EndUser")

			err = c.Find(bson.M{"endusername": user.EndUsername}).One(&result)
			if err != nil {
				log.Println(err)
			}

			user.Password = result.Password
			user.City = result.City
			user.State = result.State
			user.PinCode = result.PinCode
			user.Email = result.Email
			user.FirstName = result.FirstName
			user.LastName = result.LastName

			colQuerier := bson.M{"endusername": user.EndUsername}
			err = c.Update(colQuerier, user)
			if err != nil {
				response = "Failed :("
			} else {
				response = "Success"
			}
		}

		fmt.Fprintf(w, `{"response":"%s" }`, response)
	}

	log.Println(r.URL.Path)
}

func regSalon(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		respond := helpeReg("Saloon", r)

		fmt.Fprintf(w, `{
			"response":"%s"
		}`, respond)
	}

	log.Println(r.URL.Path)
}

func regUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		respond := helpeReg("user", r)

		fmt.Fprintf(w, `{
			"response":"%s"
		}`, respond)
	}

	log.Println(r.URL.Path)
}

func login(w http.ResponseWriter, r *http.Request) {
	var response string

	if r.Method == http.MethodPost {
		session, err := mgo.Dial(connectionString)

		if err != nil {
			panic(err)
		}
		defer session.Close()

		session.SetMode(mgo.Monotonic, true)

		c := session.DB("SaloonData").C(r.FormValue("type"))

		if r.FormValue("type") == "EndUser" {
			var user endUserInfo

			user.EndUsername = r.FormValue("user")
			pass := r.FormValue("pass")

			user.Password = stringToBits(pass)

			result := endUserInfo{}
			err = c.Find(bson.M{"endusername": user.EndUsername, "password": user.Password}).One(&result)

			if err == nil {
				if result.Status != "true" {

				} else {
					response = "Success"
				}
			} else {
				response = "Wrong Username or Password"
			}

		} else {
			var sal salonInfo
			sal.Username = r.FormValue("user")
			pass := r.FormValue("pass")

			sal.Password = stringToBits(pass)

			result := salonInfo{}
			err = c.Find(bson.M{"username": sal.Username}).One(&result)

			if err == nil {
				if sal.Password == result.Password {
					if result.Status != "true" {
						response = "Invalid Account"
					} else {
						response = "Success"
					}
				} else {
					response = "Wrong Username or Password"
				}
			} else {
				response = "Wrong Username or Password"
			}
		}

		fmt.Fprintf(w, `{
		"response":"%s"
	}`, response)

		log.Println(r.URL.Path)
	}
}

func salComment(w http.ResponseWriter, r *http.Request) {
	var comment commentsSection
	var res []commentsSection

	if r.Method == http.MethodPost {
		sesson, err := mgo.Dial(connectionString)

		if err != nil {
			panic(err)
		}
		defer sesson.Close()

		sesson.SetMode(mgo.Monotonic, true)

		c := sesson.DB("SaloonData").C("commentsSection")

		comment.SalonName = r.FormValue("sname")
		comment.EndUsername = r.FormValue("user")
		comment.Comment = r.FormValue("comment")
		comment.Star = r.FormValue("star")
		//comment.timestamp = time.Now().String()

		err = c.Insert(comment)

		c.Find(bson.M{"salonname": comment.SalonName}).All(&res)
		respond, _ := json.MarshalIndent(res, "", " ")

		fmt.Fprintf(w, `{
			"response": %s}`, string(respond))
	}

	log.Println(r.URL.Path)
}

func updateComments(w http.ResponseWriter, r *http.Request) {
	var comment commentsSection
	var res []commentsSection

	if r.Method == http.MethodPost {
		sesson, err := mgo.Dial(connectionString)

		if err != nil {
			panic(err)
		}
		defer sesson.Close()

		sesson.SetMode(mgo.Monotonic, true)

		c := sesson.DB("SaloonData").C("commentsSection")

		comment.SalonName = r.FormValue("sname")
		comment.EndUsername = r.FormValue("user")
		comment.Comment = r.FormValue("comment")
		comment.Star = r.FormValue("star")

		colQuerier := bson.M{"endusername": comment.EndUsername, "salonname": comment.SalonName}
		err = c.Update(colQuerier, comment)

		c.Find(bson.M{"salonname": comment.SalonName}).All(&res)
		respond, _ := json.MarshalIndent(res, "", " ")

		fmt.Fprintf(w, `{
			"response": %s}`, string(respond))
	}

	log.Println(r.URL.Path)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	var user endUserInfo
	var sal salonInfo

	if r.Method == http.MethodPost {
		session, err := mgo.Dial(connectionString)

		if err != nil {
			panic(err)
		}
		defer session.Close()

		session.SetMode(mgo.Monotonic, true)

		c := session.DB("SaloonData").C(r.FormValue("type"))

		if r.FormValue("type") == "EndUser" {
			user.EndUsername = r.FormValue("user")
			result := endUserInfo{}
			err := c.Find(bson.M{"endusername": user.EndUsername}).One(&result)

			if err != nil {
				log.Println(err)
			} else {
				response, _ := json.MarshalIndent(result, "", " ")
				fmt.Fprintf(w, string(response))
			}

		} else {
			sal.Username = r.FormValue("user")
			result := salonInfo{}
			err := c.Find(bson.M{"username": sal.Username}).One(&result)

			if err != nil {
				log.Println(err)
			} else {
				response, _ := json.MarshalIndent(result, "", " ")
				fmt.Fprintf(w, string(response))
			}
		}
	}

	log.Println(r.URL.Path)
}

func getComments(w http.ResponseWriter, r *http.Request) {
	var getComments []commentsSection
	var com commentsSection

	if r.Method == http.MethodPost {
		sesson, err := mgo.Dial(connectionString)

		if err != nil {
			panic(err)
		}
		defer sesson.Close()
		sesson.SetMode(mgo.Monotonic, true)

		c := sesson.DB("SaloonData").C("commentsSection")

		com.SalonName = r.FormValue("sname")

		c.Find(bson.M{"salonname": com.SalonName}).All(&getComments)

		respond, _ := json.MarshalIndent(getComments, "", " ")

		fmt.Fprintf(w, `{
			"response": %s}`, string(respond))
	}

	log.Println(r.URL.Path)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	var user endUserInfo
	var response string
	result := endUserInfo{}

	if r.Method == http.MethodPost {
		sesson, err := mgo.Dial(connectionString)

		if err != nil {
			panic(err)
		}
		defer sesson.Close()
		sesson.SetMode(mgo.Monotonic, true)

		c := sesson.DB("SaloonData").C("EndUser")

		user.EndUsername = r.FormValue("user")
		colQuerier := bson.M{"endusername": user.EndUsername}

		err = c.Find(bson.M{"endusername": user.EndUsername}).One(&result)

		user.FirstName = r.FormValue("fname")
		user.LastName = r.FormValue("lname")
		user.Email = r.FormValue("email")
		user.City = r.FormValue("city")
		user.State = r.FormValue("state")
		user.PinCode = r.FormValue("pin")
		user.Password = result.Password
		user.State = result.Status

		err = c.Update(colQuerier, user)

		if err != nil {
			response = "Success"
		} else {
			response = "Failed :("
		}

		fmt.Fprintf(w, `{
			"response":"%s"
		}`, response)
	}

	log.Println(r.URL.Path)
}

func searchSaloon(w http.ResponseWriter, r *http.Request) {
	var searchReasults []salonInfo

	sesson, err := mgo.Dial(connectionString)

	if err != nil {
		panic(err)
	}
	defer sesson.Close()
	sesson.SetMode(mgo.Monotonic, true)

	c := sesson.DB("SaloonData").C("Saloon")

	c.Find(nil).All(&searchReasults)

	respond, _ := json.MarshalIndent(searchReasults, "", " ")

	fmt.Fprintf(w, `{
			"SearchResults": %s}`, string(respond))

	log.Println(r.URL.Path)
}

func updateUserPassword(w http.ResponseWriter, r *http.Request) {
	response := "Failed :("

	if r.Method == http.MethodPost {

		response = helpUpdate("EndUser", r)

		fmt.Fprintf(w, `{
			"response":"%s"
		}`, response)
	}

	log.Println(r.URL.Path)
}

func updateSalonPassword(w http.ResponseWriter, r *http.Request) {
	response := "Failed :("

	if r.Method == http.MethodPost {

		response = helpUpdate("Saloon", r)

		fmt.Fprintf(w, `{
			"response":"%s"
		}`, response)
	}

	log.Println(r.URL.Path)
}

//******************************************************//
// Helpers
//******************************************************//

func helpUpdate(useSalon string, r *http.Request) string {
	var response string

	sesson, err := mgo.Dial(connectionString)

	if err != nil {
		log.Println(err)
	}
	defer sesson.Close()
	sesson.SetMode(mgo.Monotonic, true)
	c := sesson.DB("SaloonData").C(useSalon)

	if useSalon == "EndUser" {

		var user endUserInfo
		result := endUserInfo{}

		user.EndUsername = r.FormValue("user")
		pass := r.FormValue("newPass")
		passCheck := r.FormValue("oldPass")

		user.Password = stringToBits(pass)
		passCheck = stringToBits(passCheck)

		err = c.Find(bson.M{"endusername": user.EndUsername, "password": passCheck}).One(&result)
		if err != nil {
			response = "Failed :("
		} else {
			colQuerier := bson.M{"endusername": user.EndUsername}

			user.FirstName = result.FirstName
			user.LastName = result.LastName
			user.City = result.City
			user.State = result.State
			user.PinCode = result.PinCode
			user.Email = result.Email
			user.Status = result.Status

			err = c.Update(colQuerier, user)
			if err != nil {
				response = "Failed :("
			} else {
				response = "Success"
			}
		}

	} else {
		var sal salonInfo
		result := salonInfo{}

		sal.Username = r.FormValue("user")
		pass := r.FormValue("newPass")
		passCheck := r.FormValue("oldPass")

		passCheck = stringToBits(passCheck)
		sal.Password = stringToBits(pass)

		err = c.Find(bson.M{"username": sal.Username, "password": passCheck}).One(&result)
		if err != nil {
			response = "Failed :("
		} else {
			colQuerier := bson.M{"username": sal.Username}

			sal.SalonName = result.SalonName
			sal.ShortDesc = result.SalonName
			sal.City = result.City
			sal.State = result.State
			sal.PinCode = result.PinCode
			sal.Email = result.Email
			sal.Status = result.Status

			err = c.Update(colQuerier, sal)
			if err != nil {
				response = "Failed :("
			} else {
				response = "Success"
			}
		}

	}

	return response
}

func helpeReg(useSalon string, r *http.Request) string {

	var response string

	session, err := mgo.Dial(connectionString)

	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	if useSalon == "Saloon" {

		var sal salonInfo
		c := session.DB("SaloonData").C("Saloon")

		sal.Username = r.FormValue("user")
		pass := r.FormValue("pass")
		sal.City = r.FormValue("city")
		sal.State = r.FormValue("state")
		sal.PinCode = r.FormValue("pin")
		sal.SalonName = r.FormValue("sname")
		sal.Email = r.FormValue("email")
		sal.ShortDesc = r.FormValue("desc")
		sal.Status = "false"
		OTP = r.FormValue("otp")

		sal.Password = stringToBits(pass)
		//		sal.timestamp = time.Now().String()

		result := salonInfo{}

		err := c.Find(bson.M{"username": sal.Username}).One(&result)
		err1 := c.Find(bson.M{"email": sal.Email}).One(&result)

		if err == nil || err1 == nil {
			response = "User already exists"
		} else {
			err = sendMail(`
			<!DOCTYPE html>
			<html>
				<body>
					<h2> Hey there </h2>
					<p> Thank you for taking interest in Truudus. Here's your One Time Password </p> <br>
					<h3> `+OTP+`</h3>
				</body>
			</html>`, sal.Email)

			if err != nil {
				response = "Invalid Email :( "
				log.Println(err)
			} else {
				err = c.Insert(sal)

				if err != nil {
					response = "Failed :("
					log.Println(err)
				} else {
					response = "Success"
				}
			}
		}

	} else {

		var user endUserInfo
		c := session.DB("SaloonData").C("EndUser")

		user.FirstName = r.FormValue("fname")
		user.LastName = r.FormValue("lname")
		user.EndUsername = r.FormValue("user")
		pass := r.FormValue("pass")
		user.City = r.FormValue("city")
		user.State = r.FormValue("state")
		user.PinCode = r.FormValue("pin")
		user.Email = r.FormValue("email")
		user.Status = "false"
		OTP = r.FormValue("otp")

		user.Password = stringToBits(pass)
		//		user.timestamp = time.Now().String()

		result := endUserInfo{}

		err := c.Find(bson.M{"endusername": user.EndUsername}).One(&result)
		err1 := c.Find(bson.M{"email": user.Email}).One(&result)

		if err == nil || err1 == nil {
			response = "User already exists"
		} else {
			err = sendMail(`
			<!DOCTYPE html>
			<html>
				<body>
					<h2> Hey there </h2>
					<p> Thank you for taking interest in Truudus. Here's your One Time Password </p> <br>
					<h3> `+OTP+`</h3>
				</body>
			</html>`, user.Email)

			if err != nil {
				response = "Invalid Email :( "
				log.Println(err)
			} else {
				err = c.Insert(user)
				if err != nil {
					response = "Failed :( "
					log.Println(err)
				} else {
					response = "Success"
				}
			}
		}
	}

	return response
}

func convertToBits(n, pad int) string {
	var result string

	for ; n > 0; n = n / 2 {
		if n%2 == 0 {
			result = "1" + result
		} else if n%3 == 0 {
			result = "1" + result
		} else {
			result = "0" + result
		}
	}

	for i := len(result); i < pad; i++ {
		result = "0" + result
	}

	return result
}

func stringToBits(str string) string {
	var result string

	data := []rune(str)

	for _, i := range data {
		result = convertToBits(int(i), 8) + result
	}

	return result
}

func sendMail(Body, to string) error {
	from := "truuduss"
	password := "desmond_TRUUDUS12"

	msg := "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"MIME-Version: 1.0" + "\r\n" +
		"Content-type: text/html" + "\r\n" +
		"Subject: Reigstration Success" + "\r\n\r\n" +
		Body + "\r\n"

	err := smtp.SendMail("smtp.gmail.com:587", smtp.PlainAuth("", from, password, "smtp.gmail.com"), from, []string{to}, []byte(msg))
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Verification Message Sent")
	return nil
}

func uploadPage(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		// To recieve a file, for html its going to be input type="file" name="file"
		src, hdr, err := r.FormFile("blogFile")
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 500)
			return
		}
		defer src.Close()

		//writing file by creating one
		dst, err := os.Create("./assets/images/" + hdr.Filename)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 500)
			return
		}
		defer dst.Close()

		// copy the uploaded file
		_, err = io.Copy(dst, src)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintf(w, `{"response":"Success"}`)
	}

	log.Println(r.URL.Path)
}
