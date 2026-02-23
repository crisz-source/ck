package k8s

import (
	"context"
	"fmt"
	"time"

	"ck/notify"

	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// WatchPods inicia o monitoramento de pods em tempo real.
// Usa Informer (Watch + cache) — o mesmo padrão que Operators usam.
// Bloqueia até o contexto ser cancelado (Ctrl+C).
func WatchPods(ctx context.Context, clientset *kubernetes.Clientset, namespace string) {

	var factory informers.SharedInformerFactory
	if namespace == "" {
		factory = informers.NewSharedInformerFactory(clientset, 30*time.Second)
	} else {
		factory = informers.NewSharedInformerFactoryWithOptions(
			clientset,
			30*time.Second,
			informers.WithNamespace(namespace),
		)
	}

	podInformer := factory.Core().V1().Pods().Informer()

	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{

		// Pod novo apareceu — checa se já nasceu com problema
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				return
			}
			checkPodProblems(pod, "ADDED")
		},

		// Pod mudou — aqui pega restarts, crashes, OOMKills
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldPod, ok1 := oldObj.(*corev1.Pod)
			newPod, ok2 := newObj.(*corev1.Pod)
			if !ok1 || !ok2 {
				return
			}
			detectChanges(oldPod, newPod)
		},

		// Pod deletado
		DeleteFunc: func(obj interface{}) {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				return
			}
			fmt.Printf("🗑️  Pod deletado: %s/%s\n", pod.Namespace, pod.Name)
		},
	})

	stopCh := ctx.Done()
	factory.Start(stopCh)
	factory.WaitForCacheSync(stopCh)

	if namespace == "" {
		fmt.Println("👁️  Monitorando pods de TODOS os namespaces...")
	} else {
		fmt.Printf("👁️  Monitorando pods do namespace: %s\n", namespace)
	}

	threshold := viper.GetInt32("watch.restart_threshold")
	if threshold == 0 {
		threshold = 3
	}
	fmt.Printf("🔔 Alerta quando restarts >= %d\n", threshold)
	fmt.Printf("📧 Email para: %s\n", viper.GetString("notify.email.to"))
	fmt.Println("   Pressione Ctrl+C para parar")
	fmt.Println("─────────────────────────────────")

	// Bloqueia aqui até Ctrl+C
	<-stopCh
}

func detectChanges(oldPod, newPod *corev1.Pod) {
	threshold := viper.GetInt32("watch.restart_threshold")
	if threshold == 0 {
		threshold = 3
	}

	for i, newCS := range newPod.Status.ContainerStatuses {
		// Pega o container status antigo pra comparar
		var oldRestarts int32
		if i < len(oldPod.Status.ContainerStatuses) {
			oldRestarts = oldPod.Status.ContainerStatuses[i].RestartCount
		}

		if newCS.RestartCount > oldRestarts {
			// Descobre o motivo do restart
			reason := "Unknown"
			if newCS.LastTerminationState.Terminated != nil {
				reason = newCS.LastTerminationState.Terminated.Reason
			}

			eventType := "RESTART"
			if reason == "OOMKilled" {
				eventType = "OOM_KILLED"
			}

			// Checa se é CrashLoopBackOff
			if newCS.State.Waiting != nil &&
				newCS.State.Waiting.Reason == "CrashLoopBackOff" {
				eventType = "CRASHLOOP"
			}

			event := notify.PodEvent{
				PodName:      newPod.Name,
				Namespace:    newPod.Namespace,
				EventType:    eventType,
				RestartCount: newCS.RestartCount,
				Reason:       reason,
				Timestamp:    time.Now(),
			}

			// Sempre mostra no terminal
			notify.PrintAlert(event)

			// Envia email se passou do threshold
			if newCS.RestartCount >= threshold {
				if err := notify.SendEmail(event); err != nil {
					fmt.Printf("⚠️  Erro ao enviar email: %v\n", err)
				} else {
					fmt.Println("✅ Email enviado!")
				}
			}
		}
	}
}

// checkPodProblems verifica se um pod recém-detectado já tem problemas
func checkPodProblems(pod *corev1.Pod, eventSource string) {
	for _, cs := range pod.Status.ContainerStatuses {
		// Pod já tá em CrashLoopBackOff?
		if cs.State.Waiting != nil &&
			cs.State.Waiting.Reason == "CrashLoopBackOff" {

			event := notify.PodEvent{
				PodName:      pod.Name,
				Namespace:    pod.Namespace,
				EventType:    "CRASHLOOP",
				RestartCount: cs.RestartCount,
				Reason:       "CrashLoopBackOff (detectado no " + eventSource + ")",
				Timestamp:    time.Now(),
			}

			notify.PrintAlert(event)
		}
	}
}
