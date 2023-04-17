package handlers

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/caiofralmeida/kube-image-cacher/internal/registry"
	"github.com/docker/docker/api/types"
	dockercli "github.com/docker/docker/client"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type PodImageCacher struct {
	Client     client.Client
	ECRService *ecr.ECR
	Registry   registry.Registry
	decoder    *admission.Decoder
}

var authToken *string

func (h *PodImageCacher) Handle(ctx context.Context, req admission.Request) admission.Response {

	pod := &corev1.Pod{}
	err := h.decoder.Decode(req, pod)

	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	for k, container := range pod.Spec.Containers {
		pod.Spec.Containers[k] = h.mutateImage(ctx, container)
	}

	jsonPod, err := json.Marshal(pod)
	return admission.PatchResponseFromRaw(req.Object.Raw, jsonPod)
}

func (h *PodImageCacher) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *PodImageCacher) mutateImage(ctx context.Context, container corev1.Container) corev1.Container {

	fromRegistry := strings.Contains(container.Image, h.Registry.URL)

	if fromRegistry {
		return container
	}

	_, err := h.ECRService.DescribeRepositories(&ecr.DescribeRepositoriesInput{
		RepositoryNames: aws.StringSlice([]string{"httpd"}),
	})

	if tmp := new(ecr.RepositoryNotFoundException); errors.As(err, &tmp) {
		fmt.Println("creating the new repository")
		h.ECRService.CreateRepository(&ecr.CreateRepositoryInput{
			RepositoryName: aws.String("httpd"),
		})
	}

	if authToken == nil {
		authOut, _ := h.ECRService.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
		authToken = authOut.AuthorizationData[0].AuthorizationToken
	}

	decode, _ := base64.StdEncoding.DecodeString(*authToken)
	userPassword := strings.Split(string(decode), ":")

	docker, err := dockercli.NewEnvClient()

	var authConfig = types.AuthConfig{
		Username:      userPassword[0],
		Password:      userPassword[1],
		ServerAddress: h.Registry.URL,
	}
	authConfigBytes, _ := json.Marshal(authConfig)
	authConfigEncoded := base64.URLEncoding.EncodeToString(authConfigBytes)

	image, err := docker.ImagePull(ctx, container.Image, types.ImagePullOptions{})
	defer image.Close()
	io.Copy(os.Stdout, image)

	destination := fmt.Sprintf("%s/%s:latest", h.Registry.URL, container.Image)

	errTag := docker.ImageTag(ctx, container.Image, destination)

	rd, errPush := docker.ImagePush(ctx, destination, types.ImagePushOptions{RegistryAuth: authConfigEncoded})
	if errPush == nil {
		defer rd.Close()
	}

	errRd := print(rd)

	fmt.Println(rd, errTag, errPush, errRd)

	container.Image = "httpd:stable"
	return container
}

type ErrorLine struct {
	Error       string      `json:"error"`
	ErrorDetail ErrorDetail `json:"errorDetail"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}

func print(rd io.Reader) error {
	var lastLine string

	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		lastLine = scanner.Text()
		fmt.Println(scanner.Text())
	}

	errLine := &ErrorLine{}
	json.Unmarshal([]byte(lastLine), errLine)
	if errLine.Error != "" {
		return errors.New(errLine.Error)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
