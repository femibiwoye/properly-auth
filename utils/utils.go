package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	gomail "gopkg.in/mail.v2"

	"github.com/haibeey/struct2Map"
)

//BearerTokenHeader header token for jwt
type BearerTokenHeader struct {
	Token string `header:"Authorization"`
}

//CreateToken returns a jwt string used for authetication
func CreateToken(userid string) (string, error) {
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["user_id"] = userid
	atClaims["exp"] = time.Now().Add(time.Minute * 1500).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		return "", err
	}
	return token, nil
}

//DecodeJWTToken decodes a jwt auth token
func DecodeJWTToken(tokenString string) (map[string]string, error) {
	if strings.HasPrefix(tokenString, "Bearer ") || strings.HasPrefix(tokenString, "bearer ") {

		tokArr := strings.Split(tokenString, " ")

		if len(tokArr) != 2 {
			return nil, fmt.Errorf("Bad Request couldn't parse token   na here")
		}
		tokenString = tokArr[1]
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("SECRET_KEY")), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("Invaliad token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("Token Expired")
	}

	res := make(map[string]string)
	res["user_id"] = claims["user_id"].(string)

	return res, nil
}

//DecodeJWT parse a jwt token . To be used only in a request
func DecodeJWT(c *gin.Context) (map[string]string, error) {
	token := BearerTokenHeader{}

	err := c.ShouldBindHeader(&token)

	if err != nil {
		return nil, fmt.Errorf("Bad Request couldn't parse data")
	}

	tokArr := strings.Split(token.Token, " ")

	if len(tokArr) != 2 {
		return nil, fmt.Errorf("Bad Request couldn't parse token")
	}

	res, err := DecodeJWTToken(tokArr[1])

	if err != nil {
		return nil, fmt.Errorf("Couldn't parse token %v", err)
	}
	return res, nil
}

//GenerateRandomDigit genrates random strings of numbers
//Values generated are exclusively digit
func GenerateRandomDigit(size int) string {
	if len(os.Getenv("TESTING")) > 0 {
		return "111111"
	}
	b := make([]byte, size)
	rand.Read(b)

	result := ""
	for i := 0; i < size; i++ {
		curValue := b[i]
		for curValue < 48 || curValue > 57 {
			if curValue < 48 {
				curValue += 5
			}
			if curValue > 57 {
				curValue -= 5
			}
		}
		result += string(curValue)
	}
	return result
}

//GeneratePUMCCode genrates pumc code for each user
//Values generated are exclusively digit and characters
func GeneratePUMCCode(size int) string {
	b := make([]byte, size)
	rand.Read(b)

	result := ""
	for i := 0; i < size; i++ {
		curValue := b[i]
		for curValue < 48 || curValue > 57 && curValue < 65 || curValue > 90 && curValue < 97 || curValue > 122 {
			if curValue < 48 {
				curValue += 5
			}
			if curValue > 57 {
				curValue += 8
			}
			if curValue > 122 {
				curValue -= 5
			}
		}
		result += string(curValue)
	}
	return result
}

//SendMail use to authenticate user
func SendMail(emailRecipent, subject, body string) error {
	if os.Getenv("TESTING") == "TESTING" {
		return nil
	}
	m := gomail.NewMessage()

	// Set E-Mail sender
	m.SetHeader("From", os.Getenv("EMAIL_SENDER"))
	m.SetHeader("To", emailRecipent)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer("smtp.gmail.com", 465, os.Getenv("EMAIL_SENDER"), os.Getenv("EMAIL_SENDER_PASSWORD"))
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

//SHA256Hash hash of a string
func SHA256Hash(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return fmt.Sprintf("%x", h.Sum(nil))
}

//MissingDataResponse returns the required field for fil that are nil
func MissingDataResponse(dataModel interface{}) (map[string][]string, error) {
	response := make(map[string][]string)
	value, err := struct2map.Struct2Map(dataModel)
	if err != nil {
		return nil, err
	}

	for i, v := range value {
		if v == 0 || v == "" || v == nil {
			response[i] = []string{fmt.Sprintf("%s cannot be blank.", i)}
		}
	}
	return response, nil
}
