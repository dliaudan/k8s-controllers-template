package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var goBasicCmd = &cobra.Command{
	Use:   "go-basic",
	Short: "Run golang basic code",
	Run: func(cmd *cobra.Command, args []string) {
		//Go basic code to run functions
		k8s := Kubernetes{
			Name:    "k8s-demo-cluster",
			Version: "1.31",
			Users:   []string{"alex", "den"},
			NodeNumber: func() int {
				return 10
			},
		}

		//print users
		k8s.GetUsers()

		//add new user to struct
		k8s.AddNewUser("anonymous")

		//print users one more time
		k8s.GetUsers()
	},
}

var addNewUser = &cobra.Command{
	Use:   "add-user [user]",
	Short: "Add new user to kubernetes cluster",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Error: please provide a username")
			return
		}
		//Go basic code to run functions
		k8s := Kubernetes{
			Name:    "k8s-demo-cluster",
			Version: "1.31",
			Users:   []string{"alex", "den"},
			NodeNumber: func() int {
				return 10
			},
		}

		//add new user to struct
		for _, username := range args {
			k8s.AddNewUser(username)
		}

		//print users one more time
		k8s.GetUsers()
	},
}

var describeCluster = &cobra.Command{
	Use:   "describe-cluster",
	Short: "Describe kubernetes cluster",
	Run: func(cmd *cobra.Command, args []string) {
		k8s := Kubernetes{
			Name:    "k8s-demo-cluster",
			Version: "1.31",
			Users:   []string{"alex", "den"},
		}
		k8s.GetClusterInfo()
	},
}

func init() {
	rootCmd.AddCommand(goBasicCmd)
	rootCmd.AddCommand(addNewUser)
	rootCmd.AddCommand(describeCluster)
}

// My go basic fucntions here
type Kubernetes struct {
	Name       string     `json:"name"`
	Version    string     `json:"version"`
	Users      []string   `json:"users,omitempty"`
	NodeNumber func() int `json:"-"`
}

func (k8s Kubernetes) GetUsers() {
	for _, user := range k8s.Users {
		fmt.Println(user)
	}
}

func (k8s *Kubernetes) AddNewUser(user string) {
	k8s.Users = append(k8s.Users, user)
}

func (k8s *Kubernetes) GetClusterInfo() {
	fmt.Println("Cluster name: ", k8s.Name)
	fmt.Println("Kubernetes version: ", k8s.Version)
}
