package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
)

// User .
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// JwtToken .
type JwtToken struct {
	Token string `json:"token"`
}

// Exception .
type Exception struct {
	Message string `json:"message"`
}

var signingKey = []byte(os.Getenv("SIGNING_KEY"))

// Person .
type Person struct {
	ID        string   `json:"id,omitempty"`
	Firstname string   `json:"firstname,omitempty"`
	Lastname  string   `json:"lastname,omitempty"`
	Address   *Address `json:"address,omitempty"`
}

// Address .
type Address struct {
	City  string `json:"city,omitempty"`
	State string `json:"state,omitempty"`
}

var people = make(map[string]Person)

// GetPersonEndpoint .
func GetPersonEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	person, _ := people[params["id"]]

	// person might be empty if it wasn't found in the map
	json.NewEncoder(w).Encode(person)
}

// GetPeopleEndpoint .
func GetPeopleEndpoint(w http.ResponseWriter, req *http.Request) {
	json.NewEncoder(w).Encode(people)
}

// CreatePersonEndpoint {"id":"2","firstname":"Mohan","lastname":"Pratap","address":{"city":"Chennai","state":"TN"}}
func CreatePersonEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	//fmt.Fprintf(os.Stderr, "Params: %v\n", params);
	var person Person
	result := json.NewDecoder(req.Body).Decode(&person)
	if result != nil {
		fmt.Fprintf(os.Stderr, "result=%v\n", result)
		return
	}

	//fmt.Fprintf(os.Stderr, "Person: %v\n", person);
	person.ID, _ = params["id"]
	people[person.ID] = person
	json.NewEncoder(w).Encode(people)
}

// DeletePersonEndpoint .
func DeletePersonEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	id, _ := params["id"]

	delete(people, id)
	json.NewEncoder(w).Encode(people)
}

// CreateTokenEndpoint .
func CreateTokenEndpoint(w http.ResponseWriter, req *http.Request) {
	var user User
	_ = json.NewDecoder(req.Body).Decode(&user)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"password": user.Password,
	})
	tokenString, error := token.SignedString(signingKey)
	if error != nil {
		fmt.Println(error)
	}
	json.NewEncoder(w).Encode(JwtToken{Token: tokenString})
}

// ProtectedEndpoint .
func ProtectedEndpoint(w http.ResponseWriter, req *http.Request) {
	params := req.URL.Query()
	token, _ := jwt.Parse(params["token"][0], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an Error")
		}
		return signingKey, nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		var user User
		mapstructure.Decode(claims, &user)
		json.NewEncoder(w).Encode(user)
	} else {
		json.NewEncoder(w).Encode(Exception{Message: "Invalid authorization token"})
	}
}

// ValidateMiddleware .
func ValidateMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		authorizationHeader := req.Header.Get("authorization")
		if authorizationHeader != "" {
			bearerToken := strings.Split(authorizationHeader, " ")
			if len(bearerToken) == 2 {
				token, error := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("There was an Error")
					}
					return signingKey, nil
				})
				if error != nil {
					json.NewEncoder(w).Encode(Exception{Message: error.Error()})
					return
				}
				if token.Valid {
					context.Set(req, "decoded", token.Claims)
					next(w, req)
				} else {
					json.NewEncoder(w).Encode(Exception{Message: "Invalid authorization token"})
				}
			}
		} else {
			json.NewEncoder(w).Encode(Exception{Message: "An Authorization Header is Required"})
		}
	})
}

// TestEndpoint .
func TestEndpoint(w http.ResponseWriter, req *http.Request) {
	decoded := context.Get(req, "decoded")
	var user User
	mapstructure.Decode(decoded.(jwt.MapClaims), &user)
	json.NewEncoder(w).Encode(user)
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/authenticate", CreateTokenEndpoint).Methods("POST")
	router.HandleFunc("/protected", ProtectedEndpoint).Methods("GET")
	router.HandleFunc("/test", ValidateMiddleware(TestEndpoint)).Methods("GET")
	router.HandleFunc("/people", ValidateMiddleware(GetPeopleEndpoint)).Methods("GET")
	router.HandleFunc("/people/{id}", ValidateMiddleware(GetPersonEndpoint)).Methods("GET")
	router.HandleFunc("/people/{id}", ValidateMiddleware(CreatePersonEndpoint)).Methods("POST")
	router.HandleFunc("/people/{id}", ValidateMiddleware(DeletePersonEndpoint)).Methods("DELETE")

	fmt.Println("Starting the application on Port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
