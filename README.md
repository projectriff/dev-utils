# Developer Tools
This repository provides tools for riff users to develop and debug functions. These tools are bundled in a container image that is meant to be run in the development cluster.

## Using the tools
These tools can be used by running a simple pod in the k8s cluster with a configuration like this:
```bash
kubectl run dev-utils --image=projectriff/dev-utils --generator=run-pod/v1
```
Then depending upon which riff runtime you are using, create the appropriate clusterrolebindings:
```bash
kubectl create clusterrolebinding dev-util-stream --clusterrole=riff-streaming-readonly-role --serviceaccount=default:default
kubectl create clusterrolebinding dev-util-core --clusterrole=riff-core-readonly-role --serviceaccount=default:default
kubectl create clusterrolebinding dev-util-knative --clusterrole=riff-knative-readonly-role --serviceaccount=default:default
```
The `publish` and `subscribe` tools will additionally require read access to secrets in your development namespace:
```bash
kubectl create role view-secrets-role --namespace ${NAMESPACE} --resource secrets --verb get,watch,list
kubectl create rolebinding dev-util-secrets --namespace ${NAMESPACE} --role=view-secrets-role --serviceaccount=default:default
```

## Included tools
1. **publish:** To publish an event to the given stream.
The command takes the form:
    ```
    publish <stream-name> -n <namespace> --payload <payload-as-string> --content-type <content-type> --header "<header-name>: <header-value>"
    ```
    where `stream-name`, `payload` and `content-type` are mandatory and `header` can be used multiple times.
1. **subscribe:** To subscribe for events from the given stream.
The command takes the form:
    ```
    subscribe <stream-name> --from-beginning
    ```
    If the `--from-beginning` option is present, display all the events in the stream, otherwise only new events are displayed in the following json format:
    ```
    {"payload": "base64 encoded user payload","content-type": "the content type of the message","headers": {"user provided header": "while publishing"}}
    ```
1. [jq](https://stedolan.github.io/jq/): To process JSON.

1. [base64](http://manpages.ubuntu.com/manpages/bionic/man1/base64.1.html): Encode and decode base64 strings.

1. [curl](https://curl.haxx.se/): To make HTTP requests.

The namespace parameter is optional for all the commands. If not specified, the namespace of the `dev-utils` pod will be assumed.

## Examples
These tools can be invoked using kubectl exec. some examples follow:
```bash
kubectl exec dev-utils -it -- publish letters --content-type text/plain --payload foo
kubectl exec dev-utils -it -- subscribe letters --from-beginning
```
