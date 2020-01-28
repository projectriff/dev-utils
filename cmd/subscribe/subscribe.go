package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	devutil "github.com/projectriff/developer-utils/pkg"
	client "github.com/projectriff/stream-client-go"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	fromBeginning bool
	streamGVRc      = schema.GroupVersionResource{
		Group:    "streaming.projectriff.io",
		Version:  "v1alpha1",
		Resource: "streams",
	}
	secretGVRc = schema.GroupVersionResource{
		Version:  "v1",
		Resource: "secrets",
	}
	namespaceC string
)

type Event struct {
	Payload     string            `json:"payload"`
	ContentType string            `json:"contentType"`
	Headers     map[string]string `json:"headers"`
}

func main() {
	if err := subscribeCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var eventHandler = func(ctx context.Context, payload io.Reader, contentType string, headers map[string]string) error {
	bytes, err := ioutil.ReadAll(payload)
	if err != nil {
		return err
	}

	payloadStr := base64.StdEncoding.EncodeToString(bytes)

	evt := Event{
		Payload:     payloadStr,
		ContentType: contentType,
		Headers:     headers,
	}
	marshaledEvt, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", marshaledEvt)
	return nil
}

var subscribeCmd = &cobra.Command{
	Use:     "subscribe <stream-name>",
	Short:   "subscribe for events from the given stream",
	Long:    "",
	Example: "subscribe letters --from-beginning",
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
		secretName, err := k8sClient.GetNestedString(args[0], namespaceC, streamGVRc, "status", "binding", "secretRef", "name")
		if err != nil {
			fmt.Println("error while finding binding secret reference", err)
			os.Exit(1)
		}

		encodedTopic, err := k8sClient.GetNestedString(secretName, namespaceC, secretGVRc, "data", "topic")
		if err != nil {
			fmt.Println("error while determining gateway topic for stream", err)
			os.Exit(1)
		}

		topic, err := base64.StdEncoding.DecodeString(encodedTopic)
		if err != nil {
			fmt.Println("error decoding topic", err)
			os.Exit(1)
		}

		encodedGateway, err := k8sClient.GetNestedString(secretName, namespaceC, secretGVRc, "data", "gateway")
		if err != nil {
			fmt.Println("error while determining gateway address for stream", err)
			os.Exit(1)
		}

		gateway, err := base64.StdEncoding.DecodeString(encodedGateway)
		if err != nil {
			fmt.Println("error decoding gateway address", err)
			os.Exit(1)
		}

		contentType, err := k8sClient.GetNestedString(args[0], namespaceC, streamGVRc, "spec", "contentType")
		if err != nil {
			fmt.Println("error while determining contentType for stream", err)
			os.Exit(1)
		}

		sc, err := client.NewStreamClient(string(gateway), string(topic), contentType)
		if err != nil {
			fmt.Println("error while creating stream client", err)
			os.Exit(1)
		}

		var eventErrHandler client.EventErrHandler
		eventErrHandler = func(_ context.CancelFunc, err error) {
			fmt.Printf("ERROR: %v\n", err)
		}
		_, err = sc.Subscribe(ctx, fmt.Sprintf("g%d", time.Now().UnixNano()), fromBeginning, eventHandler, eventErrHandler)
		if err != nil {
			fmt.Println("error while subscribing", err)
			os.Exit(1)
		}
		<-ctx.Done()
	},
}

func init() {
	subscribeCmd.Flags().BoolVarP(&fromBeginning, "from-beginning", "b", false, "read everything in the stream")
	subscribeCmd.Flags().StringVarP(&namespaceC, "namespace", "n", "", "namespace of the stream")
}
