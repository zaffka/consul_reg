package main

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/hashicorp/consul/api"
)

var serviceVersion = "to_be_replaced_at_build"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cli, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		slog.Error("failed to create consul client", "err_msg", err.Error())

		return
	}

	service := "test-service"
	microserviceID := "test-service-instance" + "-" + serviceVersion
	reg := &api.AgentServiceRegistration{
		ID:      microserviceID,
		Name:    service,
		Tags:    []string{"dev"},
		Port:    8080,
		Address: "localhost",
		Check: &api.AgentServiceCheck{
			CheckID:                        microserviceID,
			TTL:                            "60s",
			DeregisterCriticalServiceAfter: "120s",
		},
	}

	agent := cli.Agent()
	if err := agent.ServiceRegister(reg); err != nil {
		slog.Error("failed to create consul client", "err_msg", err.Error())

		return
	}
	defer func() {
		if err := agent.ServiceDeregister(microserviceID); err != nil {
			slog.Error("failed to deregister a service", "err_msg", err.Error())
		}
	}()

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			if err := agent.UpdateTTL(microserviceID,
				fmt.Sprintf("%v", time.Now()), "pass"); err != nil {
				slog.Error("failed to update ttl", "err_msg", err.Error())
			}

			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}
		}
	}()

	<-ctx.Done()
}
