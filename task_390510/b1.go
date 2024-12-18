package main  
import (  
    "fmt"
    "io"
    "log"
    "os"

    "github.com/mailru/easyjson"
)

// define a struct that represents the structure of your JSON data
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Age  int    `json:"age"`
}

// implement the easyjson.Unmarshaler interface for User
func (u *User) UnmarshalEasyJSON(in *jlexer.Lexer) {
    in.Delim('{')
    for !in.IsDelim('}') {
        key := in.UnsafeString()
        in.WantColon()
        switch key {
        case "id":
            u.ID = int(in.UnsafeInt64())
        case "name":
            u.Name = in.UnsafeString()
        case "age":
            u.Age = int(in.UnsafeInt64())
        default:
            in.SkipRecursive()
        }
        in.WantComma()
    }
    in.Delim('}')
}

func main() {
    // open the JSON file
    file, err := os.Open("large_file.json")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    // create a new easyjson decoder
    dec := easyjson.NewDecoder(file)

    // initialize a slice to store the users
    var users []*User

    // parse the JSON file streamingly
    for {
        var user User
        if err := dec.Decode(&user); err != nil {
            if err == io.EOF {
                break
            }
            log.Fatal(err)
        }
        users = append(users, &user)
    }

    // do something with the parsed data
    fmt.Println("Number of users:", len(users))
}