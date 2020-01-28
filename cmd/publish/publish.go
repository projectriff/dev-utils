package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	devutil "github.com/projectriff/developer-utils/pkg"
	client "github.com/projectriff/stream-client-go"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	payload     string
	payloadBase64 string
	contentType string
	header      []string
	streamGVRp  = schema.GroupVersionResource{
		Group:    "streaming.projectriff.io",
		Version:  "v1alpha1",
		Resource: "streams",
	}
	secretGVRp = schema.GroupVersionResource{
		Version:  "v1",
		Resource: "secrets",
	}
	namespaceP string
)

func main() {
	if err := publishCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var publishCmd = &cobra.Command{
	Use:     "publish <stream-name> <payload>",
	Short:   "publish events to the given stream",
	Long:    "",
	Example: "publish letters --content-type text/plain --payload my-value",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		stop := devutil.SetupSignalHandler()
		go func() {
			select {
			case <-stop:
				cancel()
			}
		}()

		k8sClient := devutil.NewK8sClient()
		secretName, err := k8sClient.GetNestedString(args[0], namespaceP, streamGVRp, "status", "binding", "secretRef", "name")
		if err != nil {
			fmt.Println("error while finding binding secret reference", err)
			os.Exit(1)
		}

		encodedTopic, err := k8sClient.GetNestedString(secretName, namespaceP, secretGVRp, "data", "topic")
		if err != nil {
			fmt.Println("error while determining gateway topic for stream", err)
			os.Exit(1)
		}

		topic, err := base64.StdEncoding.DecodeString(encodedTopic)
		if err != nil {
			fmt.Println("error decoding topic", err)
			os.Exit(1)
		}

		encodedGateway, err := k8sClient.GetNestedString(secretName, namespaceP, secretGVRp, "data", "gateway")
		if err != nil {
			fmt.Println("error while determining gateway address for stream", err)
			os.Exit(1)
		}

		gateway, err := base64.StdEncoding.DecodeString(encodedGateway)
		if err != nil {
			fmt.Println("error decoding gateway address", err)
			os.Exit(1)
		}

		acceptableContentType, err := k8sClient.GetNestedString(args[0], namespaceP, streamGVRp, "spec", "contentType")
		if err != nil {
			fmt.Println("error while determining acceptableContentType for stream", err)
			os.Exit(1)
		}

		m, err := getMapFromHeaders(header)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		sc, err := client.NewStreamClient(string(gateway), string(topic), acceptableContentType)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		resolvedPayload, err := resolvePayload(payload, payloadBase64)
		_, err = sc.Publish(ctx, strings.NewReader(resolvedPayload), nil, contentType, m)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func resolvePayload(payload string, payloadBase64 string) (string, error) {
	if payload != "" && payloadBase64 != "" {
		return "", errors.New("the options --payload and --payload-base64 are mutually exclusive")
	}
	if payloadBase64 != "" {
		if !IsBase64(payloadBase64) {
			return "", errors.New("the payload is not base64 encoded")
		}
		return payloadBase64, nil
	}
	return payload, nil
}

func IsBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

func init() {
	publishCmd.Flags().StringVarP(&namespaceP, "namespace", "n", "", "namespace of the stream")
	publishCmd.Flags().StringVarP(&payload, "payload", "p", "", "the content/payload to publish to stream")
	publishCmd.Flags().StringVarP(&contentType, "content-type", "c", "", "mime type of content")
	publishCmd.Flags().StringArrayVarP(&header, "header", "", header, "headers for the payload")
	publishCmd.Flags().StringVarP(&payloadBase64, "payload-base64", "b", "", "base64 encoded payload")
	err := publishCmd.MarkFlagRequired("content-type")
	if err != nil {
		panic(err)
	}
}

func getMapFromHeaders(headers []string) (map[string]string, error) {
	returnVal := map[string]string{}
	for _, h := range headers {
		splitH := strings.Split(h, ":")
		if len(splitH) != 2 {
			return nil, errors.New(fmt.Sprintf("illegal header: %s, expected form: k1:v1", h))
		}
		returnVal[splitH[0]] = splitH[1]
	}
	return returnVal, nil
}
