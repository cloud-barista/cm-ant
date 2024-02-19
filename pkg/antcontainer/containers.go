package antcontainer

import (
	"context"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var CManager ContainerManager

type ContainerManager struct {
	DockerClient *client.Client
	Containers   map[string]ContainerConfig
}

type ContainerConfig struct {
	ID          string
	Image       string
	Environment []string
}

func init() {
	log.Println("container manager initialized")
	CManager = *NewContainerManager()
}

func NewContainerManager() *ContainerManager {

	newClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Println("Error creating Docker client:", err)
		log.Fatal(err)
	}

	if err != nil {
		log.Println("Error creating Docker client:", err)
		log.Fatal(err)
	}

	cm := &ContainerManager{
		DockerClient: newClient,
	}

	return cm
}

func (cm *ContainerManager) BuildJMeterDockerImage() error {
	buildOptions := types.ImageBuildOptions{
		Tags:       []string{"jmeter-custom-image"},
		Dockerfile: "../../Dockerfile",
	}

	buildContext, err := os.Open("../../.")
	if err != nil {
		log.Println("Error opening build context:", err)
		return err
	}
	defer buildContext.Close()

	buildResponse, err := cm.DockerClient.ImageBuild(context.Background(), buildContext, buildOptions)
	if err != nil {
		log.Println("Error building Docker image:", err)
		return err
	}
	defer buildResponse.Body.Close()
	return nil
}

// func (cm *ContainerManager) StartJMeterContainer() error {

// 	conainerName := uuid.New()

// 	resp, err := cm.DockerClient.ContainerCreate(context.Background(), &container.Config{}, nil, nil, nil, conainerName.String())
// 	if err != nil {
// 		log.Printf("error occurred while container creating; %v", err)
// 		return err
// 	}

// 	if err := cm.DockerClient.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
// 		return err
// 	}
// 	cm.DockerClient.Container

// 	containerInfo := ContainerConfig{
// 		ID:          resp.ID,
// 		Image:       cm.CreateImageName,
// 		Environment: []string{},
// 	}

// 	cm.Containers[containerInfo.ID] = containerInfo
// 	return nil
// }

// func (cm *ContainerManager) StopJMeterContainer(containerID string) error {
// 	if err := cm.DockerClient.ContainerStop(context.Background(), containerID, container.StopOptions{}); err != nil {
// 		return err
// 	}

// 	stopContainerId := ""

// 	for k := range cm.Containers {
// 		if k == containerID {
// 			stopContainerId = k
// 			break
// 		}
// 	}

// 	if stopContainerId != "" {
// 		delete(cm.Containers, stopContainerId)
// 	}
// 	return nil
// }
