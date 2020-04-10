compile-%:
	go run main.go | dot -T$*> out.$*
	open out.$*
