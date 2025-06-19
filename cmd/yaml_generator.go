package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

// setupViperConfig configures Viper to read from env file and environment variables
func setupViperConfig(envFile string, verbose bool) error {
	// Enable environment variable support first
	viper.AutomaticEnv()

	// Set environment variable prefix and replace - with _ for env vars
	viper.SetEnvPrefix("POD")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Bind environment variables to viper keys
	viper.BindEnv("pod-name")
	viper.BindEnv("container-name")
	viper.BindEnv("image")
	viper.BindEnv("tag")
	viper.BindEnv("port")
	viper.BindEnv("namespace")
	viper.BindEnv("configmap")
	viper.BindEnv("output")
	viper.BindEnv("verbose")

	// Read from env file if specified
	if envFile != "" {
		// Read the env file manually and set values
		content, err := os.ReadFile(envFile)
		if err != nil {
			return fmt.Errorf("failed to read env file: %w", err)
		}

		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Remove comments from value (everything after #)
				if commentIndex := strings.Index(value, "#"); commentIndex != -1 {
					value = strings.TrimSpace(value[:commentIndex])
				}

				// Remove quotes if present
				value = strings.Trim(value, `"'`)

				// Convert POD_* env vars to viper keys
				if strings.HasPrefix(key, "POD_") {
					// Remove POD_ prefix and convert to lowercase with dashes
					envKey := key[4:] // Remove "POD_" prefix
					viperKey := strings.ToLower(strings.ReplaceAll(envKey, "_", "-"))

					// Special handling for NAME -> pod-name
					if envKey == "NAME" {
						viperKey = "pod-name"
					}

					if verbose {
						log.Debug().Str("original_key", key).Str("env_key", envKey).Str("viper_key", viperKey).Str("value", value).Msg("Converting env var to viper key")
					}

					// Handle boolean values
					if viperKey == "verbose" {
						boolValue := value == "true" || value == "1" || value == "yes"
						viper.Set(viperKey, boolValue)
						if verbose {
							log.Debug().Str("key", key).Str("viper_key", viperKey).Bool("value", boolValue).Msg("Set boolean viper key from env file")
						}
					} else if viperKey == "port" {
						// Handle integer values
						if intValue, err := strconv.Atoi(value); err == nil {
							viper.Set(viperKey, intValue)
							if verbose {
								log.Debug().Str("key", key).Str("viper_key", viperKey).Int("value", intValue).Msg("Set integer viper key from env file")
							}
						} else {
							viper.Set(viperKey, value)
							if verbose {
								log.Debug().Str("key", key).Str("viper_key", viperKey).Str("value", value).Msg("Set viper key from env file")
							}
						}
					} else {
						viper.Set(viperKey, value)
						if verbose {
							log.Debug().Str("key", key).Str("viper_key", viperKey).Str("value", value).Msg("Set viper key from env file")
						}
					}
				}
			}
		}

		if verbose {
			log.Debug().Str("env_file", envFile).Msg("Loaded configuration from env file")
		}
	}

	return nil
}

var generatePodYaml = &cobra.Command{
	Use:   "generate-pod-yaml",
	Short: "Generate YAML file for Kubernetes Pod",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Msg("Starting YAML generation for Pod")

		// Get env file flag
		envFile, _ := cmd.Flags().GetString("env-file")

		// Check if verbose flag was set on command line
		cmdLineVerbose := false
		if cmd.Flags().Changed("verbose") {
			cmdLineVerbose, _ = cmd.Flags().GetBool("verbose")
		}

		// Setup Viper configuration with command line verbose setting
		if err := setupViperConfig(envFile, cmdLineVerbose); err != nil {
			log.Error().Err(err).Msg("Failed to setup configuration")
			return
		}

		// Get flag values using Cobra (which uses pflag under the hood)
		// Viper will automatically override with env vars if they exist
		podName := viper.GetString("pod-name")
		containerName := viper.GetString("container-name")
		image := viper.GetString("image")
		tag := viper.GetString("tag")
		port := viper.GetInt("port")
		namespace := viper.GetString("namespace")
		configMapName := viper.GetString("configmap")
		outputFile := viper.GetString("output")
		verbose := viper.GetBool("verbose")

		// Allow command line flags to override env file settings
		if cmd.Flags().Changed("verbose") {
			verbose, _ = cmd.Flags().GetBool("verbose")
		}
		if cmd.Flags().Changed("output") {
			outputFile, _ = cmd.Flags().GetString("output")
		}

		// If no env file was used, get values directly from command line flags
		if envFile == "" {
			podName, _ = cmd.Flags().GetString("pod-name")
			containerName, _ = cmd.Flags().GetString("container-name")
			image, _ = cmd.Flags().GetString("image")
			tag, _ = cmd.Flags().GetString("tag")
			port, _ = cmd.Flags().GetInt("port")
			namespace, _ = cmd.Flags().GetString("namespace")
			configMapName, _ = cmd.Flags().GetString("configmap")
		}

		// Set log level based on verbose flag BEFORE any debug logging
		if verbose {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
			log.Debug().Msg("Verbose logging enabled - showing detailed information")
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}

		// Debug logging to show what Viper read
		log.Debug().
			Str("viper_pod_name", podName).
			Str("viper_container_name", containerName).
			Str("viper_image", image).
			Str("viper_tag", tag).
			Int("viper_port", port).
			Str("viper_namespace", namespace).
			Str("viper_configmap", configMapName).
			Str("viper_output", outputFile).
			Bool("viper_verbose", verbose).
			Msg("Values read by Viper")

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
			Str("env_file", envFile).
			Msg("Parsed all command line flags and environment variables")

		// Validate required parameters
		if podName == "" || containerName == "" || image == "" || tag == "" || port == 0 {
			log.Error().Msg("Missing required flags: --pod-name, --container-name, --image, --tag, --port")
			fmt.Println("Error: please provide all required flags")
			fmt.Println("Usage: ./controller generate-pod-yaml --pod-name=my-pod --container-name=my-container --image=nginx --tag=latest --port=80")
			fmt.Println("Or use environment variables: POD_NAME, POD_CONTAINER_NAME, POD_IMAGE, POD_TAG, POD_PORT")
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
		if outputFile != "" && outputFile != "stdout" {
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
	generatePodYaml.Flags().String("env-file", "", "Environment file path (optional)")
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
