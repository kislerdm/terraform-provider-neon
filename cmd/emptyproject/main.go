// An illustration of the logic to provision a "clean" project,
// i.e., a project with no default role, database and endpoint.
// Note: the default branch is preserved as it cannot be deleted.
package main

import (
	"log"
	"os"
	"time"

	neon "github.com/kislerdm/neon-sdk-go"
)

func main() {
	t0 := time.Now()
	c, err := neon.NewClient(neon.Config{Key: os.Getenv("NEON_API_KEY")})
	if err != nil {
		log.Fatalln(err)
	}

	name := "qux"
	rProject, err := c.CreateProject(neon.ProjectCreateRequest{Project: neon.ProjectCreateRequestProject{
		Name: &name,
	}})
	if err != nil {
		log.Fatalln(err)
	}

	waitRunningOps(c, rProject.Project.ID, rProject.OperationsResponse.Operations)

	projectID := rProject.Project.ID

	branchID := rProject.Branch.ID
	for _, db := range rProject.DatabasesResponse.Databases {
		dbName := db.Name
		e := c.DeleteProjectBranchDatabase(projectID, branchID, dbName)
		if e != nil {
			log.Printf("error deleting database %s: %v\n", dbName, e)
		}
	}

	for _, role := range rProject.RolesResponse.Roles {
		e := c.DeleteProjectBranchRole(projectID, branchID, role.Name)
		if e != nil {
			log.Printf("error role %s: %v\n", role.Name, e)
		}
	}

	for _, endpoint := range rProject.EndpointsResponse.Endpoints {
		e := c.DeleteProjectEndpoint(projectID, endpoint.ID)
		if e != nil {
			log.Printf("error deleting endpoint %s: %v\n", endpoint.ID, e)
		}
	}

	if !rProject.Branch.Default {
		hardDelete := true
		e := c.DeleteProjectBranch(projectID, branchID, &hardDelete)
		if e != nil {
			log.Printf("error deleting branch %s: %v\n", branchID, e)
		}
	}

	log.Printf("lapsed: %d ms.\n", time.Since(t0).Milliseconds())
}

func waitRunningOps(c *neon.Client, projectID string, ops []neon.Operation) {
	if len(ops) > 0 {
		var wait bool
		for _, op := range ops {
			switch op.Status {
			case neon.OperationStatusRunning, neon.OperationStatusScheduling:
				wait = true
				break
			}
		}

		if wait {
			log.Println("waiting running operations")
			time.Sleep(50 * time.Millisecond)
			newOps, _ := c.ListProjectOperations(nil, nil, projectID)
			waitRunningOps(c, projectID, newOps.OperationsResponse.Operations)
		}
	}
}
