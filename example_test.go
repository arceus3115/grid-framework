package grid_test

import (
	"context"
	"fmt"
	"log"

	"github.com/arceus3115/grid-framework"
)

func ExampleNewClientFromEnv() {
	client, err := grid.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	titles, err := client.CentralData.ListTitles(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, title := range titles {
		fmt.Println(title.ID, title.Name)
	}
}

func ExampleCentralDataService_ListTeams() {
	client, err := grid.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	for team, err := range client.CentralData.ListTeams(context.Background(), grid.TeamsParams{}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(team.ID, team.Name)
	}
}
