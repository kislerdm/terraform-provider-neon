package main

import (
	"os"

	neon "github.com/kislerdm/neon-sdk-go"
)

func main() {
	client, err := neon.NewClient(neon.Config{Key: os.Getenv("NEON_API_KEY")})
	if err != nil {
		panic(err)
	}

	p, _ := client.ListProjects(nil, nil, nil, nil)
	for _, pr := range p.Projects {
		_, _ = client.DeleteProject(pr.ID)
	}
}
