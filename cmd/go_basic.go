package cmd

import (
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var goBasicCmd = &cobra.Command{
	Use:   "go-basic",
	Short: "Run golang basic code",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Msg("Starting go-basic command")

		//Go basic code to run functions
		k8s := Kubernetes{
			Name:    "k8s-demo-cluster",
			Version: "1.31",
			Users:   []string{"alex", "den"},
			NodeNumber: func() int {
				return 10
			},
		}

		log.Info().Str("cluster", k8s.Name).Str("version", k8s.Version).Msg("Initialized Kubernetes cluster")

		//print users
		log.Info().Msg("Getting current users")
		k8s.GetUsers()

		//add new user to struct
		log.Info().Str("user", "anonymous").Msg("Adding new user")
		k8s.AddNewUser("anonymous")

		//print users one more time
		log.Info().Msg("Getting updated users list")
		k8s.GetUsers()

		log.Info().Msg("go-basic command completed successfully")
	},
}

var addNewUser = &cobra.Command{
	Use:   "add-user [user]",
	Short: "Add new user to kubernetes cluster",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Strs("users", args).Msg("Starting add-user command")

		if len(args) < 1 {
			log.Error().Msg("No username provided")
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

		log.Info().Str("cluster", k8s.Name).Msg("Initialized Kubernetes cluster")

		//add new user to struct
		for _, username := range args {
			log.Info().Str("user", username).Msg("Adding user to cluster")
			k8s.AddNewUser(username)
		}

		//print users one more time
		log.Info().Msg("Getting updated users list")
		k8s.GetUsers()

		log.Info().Int("total_users", len(k8s.Users)).Msg("add-user command completed successfully")
	},
}

var describeCluster = &cobra.Command{
	Use:   "describe-cluster",
	Short: "Describe kubernetes cluster",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Msg("Starting describe-cluster command")

		k8s := Kubernetes{
			Name:    "k8s-demo-cluster",
			Version: "1.31",
			Users:   []string{"alex", "den"},
		}

		log.Info().Str("cluster", k8s.Name).Str("version", k8s.Version).Int("users_count", len(k8s.Users)).Msg("Describing cluster")
		k8s.GetClusterInfo()

		log.Info().Msg("describe-cluster command completed successfully")
	},
}

var defineNodeCount = &cobra.Command{
	Use:   "add-node",
	Short: "Add node to kubernetes cluster",
	Run: func(cmd *cobra.Command, args []string) {
		nodeCount, err := strconv.Atoi(args[0])
		if err != nil {
			log.Error().Msg("Invalid node count! Please provide a valid number.")
			return
		}

		k8s := Kubernetes{
			Name:    "k8s-demo-cluster",
			Version: "1.31",
			Users:   []string{"alex", "den"},
		}

		k8s.defineNodeCount(nodeCount)

		log.Info().Msg("add-node command completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(goBasicCmd)
	rootCmd.AddCommand(addNewUser)
	rootCmd.AddCommand(describeCluster)
	rootCmd.AddCommand(defineNodeCount)
	rootCmd.AddCommand(createPod)

	// Add flags for create-pod command
	createPod.Flags().String("name", "", "Pod name")
	createPod.Flags().String("image", "", "Container image")
	createPod.Flags().String("tag", "", "Image tag")
	createPod.Flags().Int("port", 0, "Container port")
}

type Kubernetes struct {
	Name       string     `json:"name"`
	Version    string     `json:"version"`
	Users      []string   `json:"users,omitempty"`
	NodeNumber func() int `json:"-"`
}

type Pod struct {
	Name      string `json:"name"`
	ImageRepo string `json:"image_repo"`
	ImageTag  string `json:"image_tag"`
	Port      int    `json:"port"`
}

var createPod = &cobra.Command{
	Use:   "create-pod",
	Short: "Create a pod in the Kubernetes cluster",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Msg("Starting create-pod command")

		name, _ := cmd.Flags().GetString("name")
		image, _ := cmd.Flags().GetString("image")
		tag, _ := cmd.Flags().GetString("tag")
		port, _ := cmd.Flags().GetInt("port")

		if name == "" || image == "" || tag == "" || port == 0 {
			log.Error().Msg("Missing required flags: --name, --image, --tag, --port")
			fmt.Println("Error: please provide a name, image, tag, and port")
			return
		}

		pod := Pod{
			Name:      name,
			ImageRepo: image,
			ImageTag:  tag,
			Port:      port,
		}

		log.Info().Str("name", pod.Name).Str("image", pod.ImageRepo).Str("tag", pod.ImageTag).Int("port", pod.Port).Msg("Creating pod...")
		// Add logic to create the pod in the Kubernetes cluster
	},
}

func (k8s Kubernetes) GetUsers() {
	log.Info().Int("users_count", len(k8s.Users)).Msg("Getting users list")
	for _, user := range k8s.Users {
		fmt.Println(user)
	}
	log.Info().Msg("Users list displayed successfully")
}

func (k8s *Kubernetes) AddNewUser(user string) {
	log.Info().Str("user", user).Int("current_users", len(k8s.Users)).Msg("Adding new user")
	k8s.Users = append(k8s.Users, user)
	log.Info().Str("user", user).Int("new_users_count", len(k8s.Users)).Msg("User added successfully")
}

func (k8s *Kubernetes) GetClusterInfo() {
	log.Info().Str("cluster", k8s.Name).Str("version", k8s.Version).Msg("Getting cluster information")
	fmt.Println("Cluster name: ", k8s.Name)
	fmt.Println("Kubernetes version: ", k8s.Version)
	log.Info().Msg("Cluster information displayed successfully")
}

func (k8s *Kubernetes) defineNodeCount(nodeCount int) {
	log.Info().Msg("Defining number of nodes...")
	k8s.NodeNumber = func() int {
		return nodeCount
	}
	fmt.Println("Current number of nodes on cluster is: ", nodeCount)
	if log.Info() != nil {
		log.Info().Msg("Node count defined successfully")
	} else {
		log.Error().Msg("Node count not defined")
	}
}
