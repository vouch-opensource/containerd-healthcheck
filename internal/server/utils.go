package server

import "containerdhealthcheck/internal/models"

func findContainerTask(a []models.Check, x string) int {
	for i, n := range a {
		if x == n.ContainerTask {
			return i
		}
	}
	return len(a)
}
