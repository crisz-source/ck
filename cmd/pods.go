package cmd

import (
    "fmt"
    "os/exec"

    "ck/types"
    "github.com/spf13/cobra"
    "gopkg.in/yaml.v3"
)

var podsCmd = &cobra.Command{
    Use:   "pods",
    Short: "Lista pods com problema",
    Run: func(cmd *cobra.Command, args []string) {
        listProblematicPods()
    },
}

func init() {
    RootCmd.AddCommand(podsCmd)
    podsCmd.Flags().StringVarP(&Namespace, "namespace", "n", "", "Filtrar por namespace")
}

func listProblematicPods() {
    var args []string
    if Namespace == "" {
        args = []string{"get", "pods", "-A", "-o", "yaml"}
    } else {
        args = []string{"get", "pods", "-n", Namespace, "-o", "yaml"}
    }

    out, err := exec.Command("kubectl", args...).Output()
    if err != nil {
        fmt.Println("Erro ao executar kubectl:", err)
        return
    }

    var podList types.PodList
    err = yaml.Unmarshal(out, &podList)
    if err != nil {
        fmt.Println("Erro ao parsear YAML:", err)
        return
    }

    fmt.Println("NAMESPACE\t\tNAME\t\t\t\t\tSTATUS")
    fmt.Println("---------\t\t----\t\t\t\t\t------")

    problemCount := 0
    for _, pod := range podList.Items {
        status := pod.Status.Phase
        hasProblem := false

        if status != "Running" && status != "Succeeded" {
            hasProblem = true
        }

        for _, cs := range pod.Status.ContainerStatuses {
            if cs.State.Waiting.Reason != "" {
                status = cs.State.Waiting.Reason
                hasProblem = true
            }
            if cs.RestartCount > 5 {
                status = fmt.Sprintf("%s (restarts: %d)", status, cs.RestartCount)
                hasProblem = true
            }
        }

        if hasProblem {
            fmt.Printf("%s\t\t%s\t\t%s\n", pod.Metadata.Namespace, pod.Metadata.Name, status)
            problemCount++
        }
    }

    fmt.Println("---------")
    fmt.Printf("Total: %d pods com problema\n", problemCount)
}
