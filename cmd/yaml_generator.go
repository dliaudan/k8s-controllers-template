package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
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

		// Get flag values using Cobra (which uses pflag under the hood)
		podName, _ := cmd.Flags().GetString("pod-name")
		containerName, _ := cmd.Flags().GetString("container-name")
		image, _ := cmd.Flags().GetString("image")
		tag, _ := cmd.Flags().GetString("tag")
		port, _ := cmd.Flags().GetInt("port")
		namespace, _ := cmd.Flags().GetString("namespace")
		configMapName, _ := cmd.Flags().GetString("configmap")
		outputFile, _ := cmd.Flags().GetString("output")
		verbose, _ := cmd.Flags().GetBool("verbose")

		// Set log level based on verbose flag
		if verbose {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
			log.Debug().Msg("Verbose logging enabled - showing detailed information")
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}

		log.Debug().
			Str("pod_name", podName).
			Str("container_name", containerName).
			Str("image", image).
			Str("tag", tag).
			Int("port", port).
			Str("namespace", namespace).
			Str("configmap", configMapName).
			Str("output_file", outputFile).
			Bool("verbose", verbose).
			Msg("Parsed all command line flags")

		// Validate required parameters
		if podName == "" || containerName == "" || image == "" || tag == "" || port == 0 {
			log.Error().Msg("Missing required flags: --pod-name, --container-name, --image, --tag, --port")
			fmt.Println("Error: please provide all required flags")
			fmt.Println("Usage: ./controller generate-pod-yaml --pod-name=my-pod --container-name=my-container --image=nginx --tag=latest --port=80")
			return
		}

		log.Debug().Msg("All required parameters validated successfully")

		// Create full image name
		fullImage := fmt.Sprintf("%s:%s", image, tag)
		log.Debug().Str("full_image", fullImage).Msg("Constructed full image name")

		log.Info().
			Str("pod_name", podName).
			Str("container_name", containerName).
			Str("image", fullImage).
			Int("port", port).
			Str("namespace", namespace).
			Str("configmap", configMapName).
			Msg("Generating Pod YAML")

		// Create Pod specification
		log.Debug().Msg("Creating Pod specification structure")
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

		log.Debug().
			Str("api_version", pod.APIVersion).
			Str("kind", pod.Kind).
			Str("namespace", pod.Metadata.Namespace).
			Int("containers_count", len(pod.Spec.Containers)).
			Msg("Pod specification created")

		// Add ConfigMap if specified
		if configMapName != "" {
			log.Debug().Str("configmap_name", configMapName).Msg("Adding ConfigMap volume to Pod")

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

			log.Debug().
				Str("volume_name", "config-volume").
				Str("mount_path", "/etc/config").
				Msg("ConfigMap volume and mount configured")
		} else {
			log.Debug().Msg("No ConfigMap specified - skipping volume configuration")
		}

		// Serialize to YAML
		log.Debug().Msg("Starting YAML serialization")
		yamlData, err := yaml.Marshal(pod)
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal YAML")
			return
		}
		log.Debug().Int("yaml_size", len(yamlData)).Msg("YAML serialization completed")

		// Write to file or output to stdout
		if outputFile != "" {
			log.Debug().Str("output_file", outputFile).Msg("Writing YAML to file")
			err = os.WriteFile(outputFile, yamlData, 0644)
			if err != nil {
				log.Error().Err(err).Str("file", outputFile).Msg("Failed to write YAML file")
				return
			}
			log.Info().Str("file", outputFile).Msg("YAML file generated successfully")
			fmt.Printf("YAML file generated: %s\n", outputFile)
		} else {
			log.Debug().Msg("Outputting YAML to stdout")
			fmt.Println("---")
			fmt.Println(string(yamlData))
		}

		log.Debug().Msg("Pod YAML generation process completed")
		log.Info().Msg("Pod YAML generation completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(generatePodYaml)

	// Add flags for YAML generation using pflag
	generatePodYaml.Flags().String("pod-name", "", "Pod name (required)")
	generatePodYaml.Flags().String("container-name", "", "Container name (required)")
	generatePodYaml.Flags().String("image", "", "Container image (required)")
	generatePodYaml.Flags().String("tag", "", "Image tag (required)")
	generatePodYaml.Flags().Int("port", 0, "Container port (required)")
	generatePodYaml.Flags().String("namespace", "default", "Kubernetes namespace")
	generatePodYaml.Flags().String("configmap", "", "ConfigMap name (optional)")
	generatePodYaml.Flags().String("output", "", "Output file path (optional, defaults to stdout)")
	generatePodYaml.Flags().Bool("verbose", false, "Enable verbose logging")
}
