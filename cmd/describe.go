package cmd

import (
    "fmt"
    "os/exec"
    "strings"

    "ck/types"
    "github.com/spf13/cobra"
    "gopkg.in/yaml.v3"
)

var describeCmd = &cobra.Command{
    Use:   "describe [pod (apenas o nome do pod)]",
    Short: "Mostra detalhes de um pod (resumido)",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        podName := args[0]
        describePod(podName)
    },
}

func init() {
    RootCmd.AddCommand(describeCmd)
    describeCmd.Flags().StringVarP(&Namespace, "namespace", "n", "", "Namespace do pod")
}

func describePod(podName string) {
    var args []string
    if Namespace == "" {
        args = []string{"get", "pod", podName, "-o", "yaml"}
    } else {
        args = []string{"get", "pod", podName, "-n", Namespace, "-o", "yaml"}
    }

    out, err := exec.Command("kubectl", args...).Output()
    if err != nil {
        fmt.Println("Erro ao buscar pod:", err)
        return
    }

    var pod types.Pod
    err = yaml.Unmarshal(out, &pod)
    if err != nil {
        fmt.Println("Erro ao parsear YAML:", err)
        return
    }

    fmt.Printf("=== POD: %s ===\n", podName)
    fmt.Printf("Namespace:      %s\n", pod.Metadata.Namespace)

    fmt.Printf("Labels:         ")
    if len(pod.Metadata.Labels) == 0 {
        fmt.Println("<none>")
    } else {
        labels := []string{}
        for k, v := range pod.Metadata.Labels {
            labels = append(labels, fmt.Sprintf("%s=%s", k, v))
        }
        fmt.Println(strings.Join(labels, ", "))
    }

    fmt.Printf("Status:         %s\n", pod.Status.Phase)

    for _, cs := range pod.Status.ContainerStatuses {
        fmt.Printf("\n--- Container: %s ---\n", cs.Name)
        fmt.Printf("Restart Count:  %d\n", cs.RestartCount)

        if cs.State.Waiting.Reason != "" {
            fmt.Printf("Reason:         %s\n", cs.State.Waiting.Reason)
        } else if cs.State.Terminated.Reason != "" {
            fmt.Printf("Reason:         %s (Exit Code: %d)\n", cs.State.Terminated.Reason, cs.State.Terminated.ExitCode)
        } else {
            fmt.Printf("Reason:         Running\n")
        }

        if cs.LastState.Terminated.Reason != "" {
            fmt.Printf("Last State:     %s (Exit Code: %d)\n", cs.LastState.Terminated.Reason, cs.LastState.Terminated.ExitCode)
        }
    }

    fmt.Println("\n=== EVENTOS ===")
    getEvents(podName)
}

func getEvents(podName string) {
    var args []string
    if Namespace == "" {
        args = []string{"get", "events", "--field-selector", fmt.Sprintf("involvedObject.name=%s", podName), "-o", "yaml"}
    } else {
        args = []string{"get", "events", "-n", Namespace, "--field-selector", fmt.Sprintf("involvedObject.name=%s", podName), "-o", "yaml"}
    }

    out, err := exec.Command("kubectl", args...).Output()
    if err != nil {
        fmt.Println("Erro ao buscar eventos:", err)
        return
    }

    var eventList types.EventList
    err = yaml.Unmarshal(out, &eventList)
    if err != nil {
        fmt.Println("Erro ao parsear eventos:", err)
        return
    }

    if len(eventList.Items) == 0 {
        fmt.Println("Nenhum evento encontrado")
        return
    }

    for _, event := range eventList.Items {
        fmt.Printf("%-8s %-15s %s\n", event.Type, event.Reason, event.Message)
    }
}
