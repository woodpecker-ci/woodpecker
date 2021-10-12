# woodpecker-go

```Go
import (
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
	"golang.org/x/oauth2"
)

const (
	token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
	host  = "http://woodpecker.company.tld"
)

func main() {
	// create an http client with oauth authentication.
	config := new(oauth2.Config)
	authenticator := config.Client(
		oauth2.NoContext,
		&oauth2.Token{
			AccessToken: token,
		},
	)

	// create the woodpecker client with authenticator
	client := woodpecker.NewClient(host, authenticator)

	// gets the current user
	user, err := client.Self()
	fmt.Println(user, err)

	// gets the named repository information
	repo, err := client.Repo("woodpecker-ci", "woodpecker")
	fmt.Println(repo, err)
}
```
