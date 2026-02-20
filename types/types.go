package types

// PodList - Lista de pods do kubectl
type PodList struct {
    Items []Pod `yaml:"items"`
}

// Pod - Estrutura de um pod
type Pod struct {
    Metadata struct {
        Name      string            `yaml:"name"`
        Namespace string            `yaml:"namespace"`
        Labels    map[string]string `yaml:"labels"`
    } `yaml:"metadata"`
    Status struct {
        Phase             string `yaml:"phase"`
        ContainerStatuses []struct {
            Name         string `yaml:"name"`
            Ready        bool   `yaml:"ready"`
            RestartCount int    `yaml:"restartCount"`
            State        struct {
                Waiting struct {
                    Reason string `yaml:"reason"`
                } `yaml:"waiting"`
                Terminated struct {
                    Reason   string `yaml:"reason"`
                    ExitCode int    `yaml:"exitCode"`
                } `yaml:"terminated"`
            } `yaml:"state"`
            LastState struct {
                Terminated struct {
                    Reason   string `yaml:"reason"`
                    ExitCode int    `yaml:"exitCode"`
                } `yaml:"terminated"`
            } `yaml:"lastState"`
        } `yaml:"containerStatuses"`
    } `yaml:"status"`
}

// EventList - Lista de eventos
type EventList struct {
    Items []Event `yaml:"items"`
}

// Event - Estrutura de um evento
type Event struct {
    Type    string `yaml:"type"`
    Reason  string `yaml:"reason"`
    Message string `yaml:"message"`
    Count   int    `yaml:"count"`
}

// PodMetrics - Metricas de um pod
type PodMetrics struct {
    Namespace string
    Name      string
    CPU       string
    CPUValue  int64
    Memory    string
    MemValue  int64
}
