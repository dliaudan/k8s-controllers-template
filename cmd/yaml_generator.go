package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// PodSpec represents the Kubernetes Pod specification
type PodSpec struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

type Metadata struct {
	Name      string            `yaml:"name"`
	Namespace string            `yaml:"namespace"`
	Labels    map[string]string `yaml:"labels,omitempty"`
}

type Spec struct {
	Containers []Container `yaml:"containers"`
	Volumes    []Volume    `yaml:"volumes,omitempty"`
}

type Container struct {
	Name         string        `yaml:"name"`
	Image        string        `yaml:"image"`
	Ports        []Port        `yaml:"ports,omitempty"`
	VolumeMounts []VolumeMount `yaml:"volumeMounts,omitempty"`
}

type Port struct {
	ContainerPort int    `yaml:"containerPort"`
	Protocol      string `yaml:"protocol,omitempty"`
}

type Volume struct {
	Name      string                 `yaml:"name"`
	ConfigMap *ConfigMapVolumeSource `yaml:"configMap,omitempty"`
}

type ConfigMapVolumeSource struct {
	Name string `yaml:"name"`
}

type VolumeMount struct {
	Name      string `yaml:"name"`
	MountPath string `yaml:"mountPath"`
}

var generatePodYaml = &cobra.Command{
	Use:   "generate-pod-yaml",
	Short: "Generate YAML file for Kubernetes Pod",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Msg("Starting YAML generation for Pod")

		// Get flag values from Cobra
		podName, _ := cmd.Flags().GetString("pod-name")
		containerName, _ := cmd.Flags().GetString("container-name")
		image, _ := cmd.Flags().GetString("image")
		tag, _ := cmd.Flags().GetString("tag")
		port, _ := cmd.Flags().GetInt("port")
		namespace, _ := cmd.Flags().GetString("namespace")
		configMapName, _ := cmd.Flags().GetString("configmap")
		outputFile, _ := cmd.Flags().GetString("output")

		// Validate required parameters
		if podName == "" || containerName == "" || image == "" || tag == "" || port == 0 {
			log.Error().Msg("Missing required flags: --pod-name, --container-name, --image, --tag, --port")
			fmt.Println("Error: please provide all required flags")
			fmt.Println("Usage: ./controller generate-pod-yaml --pod-name=my-pod --container-name=my-container --image=nginx --tag=latest --port=80")
			return
		}

		// Create full image name
		fullImage := fmt.Sprintf("%s:%s", image, tag)

		log.Info().
			Str("pod_name", podName).
			Str("container_name", containerName).
			Str("image", fullImage).
			Int("port", port).
			Str("namespace", namespace).
			Str("configmap", configMapName).
			Msg("Generating Pod YAML")

		// Create Pod specification
		pod := PodSpec{
			APIVersion: "v1",
			Kind:       "Pod",
			Metadata: Metadata{
				Name:      podName,
				Namespace: namespace,
				Labels: map[string]string{
					"app": podName,
				},
			},
			Spec: Spec{
				Containers: []Container{
					{
						Name:  containerName,
						Image: fullImage,
						Ports: []Port{
							{
								ContainerPort: port,
								Protocol:      "TCP",
							},
						},
					},
				},
			},
		}

		// Add ConfigMap if specified
		if configMapName != "" {
			pod.Spec.Volumes = []Volume{
				{
					Name: "config-volume",
					ConfigMap: &ConfigMapVolumeSource{
						Name: configMapName,
					},
				},
			}

			pod.Spec.Containers[0].VolumeMounts = []VolumeMount{
				{
					Name:      "config-volume",
					MountPath: "/etc/config",
				},
			}
		}

		// Serialize to YAML
		yamlData, err := yaml.Marshal(pod)
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal YAML")
			return
		}

		// Write to file or output to stdout
		if outputFile != "" {
			err = os.WriteFile(outputFile, yamlData, 0644)
			if err != nil {
				log.Error().Err(err).Str("file", outputFile).Msg("Failed to write YAML file")
				return
			}
			log.Info().Str("file", outputFile).Msg("YAML file generated successfully")
			fmt.Printf("YAML file generated: %s\n", outputFile)
		} else {
			fmt.Println("---")
			fmt.Println(string(yamlData))
		}

		log.Info().Msg("Pod YAML generation completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(generatePodYaml)

	// Add flags for YAML generation using Cobra (which uses pflag under the hood)
	generatePodYaml.Flags().String("pod-name", "", "Pod name (required)")
	generatePodYaml.Flags().String("container-name", "", "Container name (required)")
	generatePodYaml.Flags().String("image", "", "Container image (required)")
	generatePodYaml.Flags().String("tag", "", "Image tag (required)")
	generatePodYaml.Flags().Int("port", 0, "Container port (required)")
	generatePodYaml.Flags().String("namespace", "default", "Kubernetes namespace")
	generatePodYaml.Flags().String("configmap", "", "ConfigMap name (optional)")
	generatePodYaml.Flags().String("output", "", "Output file path (optional, defaults to stdout)")
}
