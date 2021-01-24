package main

// код писать тут

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

type XmlUser struct {
	Id        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

type XmlUsers struct {
	Version string    `xml:"version,attr"`
	List    []XmlUser `xml:"row"`
}

const GoodToken = "good_token"
const BadToken = "bad_token"
const TimeoutErrorQuerry = "timeout_query"
const InternalErrorQuerry = "fatal_query"
const BadRequestErrorQuerry = "bad_request_query"
const InvalidJsonErrorQuery = "Invalid_json_query"
const BadRequestUnknownErrorQuery = "bad_request_unknown_query"
const WrongJsonErrorQuery = "wrong_json_query"

func SearchServer(w http.ResponseWriter, r *http.Request) {
	// проверка токена
	token := r.Header.Get("AccessToken")
	if token != GoodToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// limit
	limit, err := strconv.Atoi(r.FormValue("limit"))
	checkError(err)

	// offset
	offset, err := strconv.Atoi(r.FormValue("offset"))
	checkError(err)

	// query
	query := r.FormValue("query")
	if query == TimeoutErrorQuerry {
		time.Sleep(time.Second * 5)
	}
	if query == InternalErrorQuerry {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if query == BadRequestErrorQuerry {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if query == BadRequestUnknownErrorQuery {
		resp, _ := json.Marshal(SearchErrorResponse{"UnknownError"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}
	if query == WrongJsonErrorQuery {
		w.Write([]byte("abc"))
		return
	}

	// order_field
	order_field := r.FormValue("order_field")
	if order_field == "" {
		order_field = "Name"
	}
	if order_field != "Id" && order_field != "Age" && order_field != "Name" {
		resp, _ := json.Marshal(SearchErrorResponse{"ErrorBadOrderField"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}

	// order_by
	order_by, err := strconv.Atoi(r.FormValue("order_by"))
	checkError(err)

	// чтение исходного xml файла
	data, err := ioutil.ReadFile("dataset.xml")
	checkError(err)

	xmlUsers := new(XmlUsers)
	err = xml.Unmarshal(data, &xmlUsers)
	checkError(err)

	users := make([]User, 0, len(xmlUsers.List))
	for _, user := range xmlUsers.List {
		item := User{
			Id:     user.Id,
			Name:   user.FirstName + " " + user.LastName,
			Age:    user.Age,
			About:  user.About,
			Gender: user.Gender,
		}
		users = append(users, item)
	}

	filteredUsers := []User{}
	if query != "" {
		for _, user := range users {
			if strings.Contains(user.Name, query) {
				filteredUsers = append(filteredUsers, user)
			}
		}
	} else {
		filteredUsers = users
	}

	//сортировка
	switch order_field {
	case "Id":
		switch order_by {
		case 1:
			sort.Slice(filteredUsers, func(i, j int) bool {
				return filteredUsers[i].Id < filteredUsers[j].Id
			})
		case -1:
			sort.Slice(filteredUsers, func(i, j int) bool {
				return filteredUsers[i].Id > filteredUsers[j].Id
			})
		}
	case "Age":
		switch order_by {
		case 1:
			sort.Slice(filteredUsers, func(i, j int) bool {
				return filteredUsers[i].Age < filteredUsers[j].Age
			})
		case -1:
			sort.Slice(filteredUsers, func(i, j int) bool {
				return filteredUsers[i].Age > filteredUsers[j].Age
			})
		}
	case "Name":
		switch order_by {
		case 1:
			sort.Slice(filteredUsers, func(i, j int) bool {
				return filteredUsers[i].Name < filteredUsers[j].Name
			})
		case -1:
			sort.Slice(filteredUsers, func(i, j int) bool {
				return filteredUsers[i].Name > filteredUsers[j].Name
			})
		}
	default:
		fmt.Println("Wrong order_filed value!")
	}

	resp, err := json.Marshal(users[offset:limit])
	checkError(err)
	w.Write(resp)
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func TestToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := SearchClient{
		URL:         ts.URL,
		AccessToken: BadToken,
	}

	result, err := client.FindUsers(SearchRequest{Query: "on", OrderField: "Age", OrderBy: 1})
	if result != nil && err.Error() != "Bad AccessToken" {
		t.Errorf("Token auth not working")
	}
}

func TestLimitAndOffset(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := SearchClient{
		URL:         ts.URL,
		AccessToken: GoodToken,
	}

	result, err := client.FindUsers(SearchRequest{Limit: -1})
	if result != nil && err.Error() != "limit must be > 0" {
		t.Errorf("Wrong limit parameter, must be > 0")
	}

	result, err = client.FindUsers(SearchRequest{Limit: 26})
	if result == nil && err != nil && len(result.Users) != 25 {
		t.Errorf("Limit not working for max 25")
	}

	result, err = client.FindUsers(SearchRequest{Offset: -1})
	if result != nil && err.Error() != "offset must be > 0" {
		t.Errorf("Wrong offset parameter, must be > 0")
	}
}

func TestNextPage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := SearchClient{
		URL:         ts.URL,
		AccessToken: GoodToken,
	}

	result, err := client.FindUsers(SearchRequest{Query: "o"})
	if result == nil && err != nil {
		t.Errorf("Next Page error")
	}

	result, err = client.FindUsers(SearchRequest{Limit: 1})
	if result == nil && err != nil {
		t.Errorf("Next Page error")
	}
	result, err = client.FindUsers(SearchRequest{Limit: 30})
	if result == nil && err != nil {
		t.Errorf("Next Page error")
	}
	result, err = client.FindUsers(SearchRequest{Limit: 25, Offset: 1})
	if result == nil && err != nil {
		t.Errorf("Next Page error")
	}
}

func TestTimeoutAndUnkonown(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := SearchClient{
		URL:         ts.URL,
		AccessToken: GoodToken,
	}

	result, err := client.FindUsers(SearchRequest{Query: TimeoutErrorQuerry})
	if result != nil && err == nil {
		t.Errorf("Timeout not working")
	}

	client = SearchClient{
		URL:         "",
		AccessToken: GoodToken,
	}

	result, err = client.FindUsers(SearchRequest{})
	if result != nil && err == nil {
		t.Errorf("Unknown error")
	}
}

func TestStatusCodes(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := SearchClient{
		URL:         ts.URL,
		AccessToken: GoodToken,
	}

	result, err := client.FindUsers(SearchRequest{Query: InternalErrorQuerry})
	if result != nil && err == nil {
		t.Errorf("Bad request error not working")
	}

	result, err = client.FindUsers(SearchRequest{Query: BadRequestErrorQuerry})
	if result != nil && err == nil {
		t.Errorf("Bad request error not working")
	}

	result, err = client.FindUsers(SearchRequest{OrderField: "bad field"})
	if result != nil && err == nil {
		t.Errorf("Bad order field")
	}

	result, err = client.FindUsers(SearchRequest{Query: BadRequestUnknownErrorQuery})
	if result != nil && err == nil {
		t.Errorf("Unknown error not working")
	}

	result, err = client.FindUsers(SearchRequest{Query: WrongJsonErrorQuery})
	if result != nil && err == nil {
		t.Errorf("Wrong json not working")
	}

}

// func main() {
// 	http.HandleFunc("/", SearchServer)

// 	fmt.Println("starting server at :8080")
// 	http.ListenAndServe(":8080", nil)
// }
